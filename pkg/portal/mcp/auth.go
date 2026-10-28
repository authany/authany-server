package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("mcp")

type TokenVerifier interface {
	VerifyResourceAccessToken(ctx context.Context, rawToken string, resourceURI string) (*model.SessionInfo, jwt.Token, error)
}

// Caller is the user an MCP request acts for, and what the client may do.
type Caller struct {
	Session  *model.SessionInfo
	ClientID string
	Scopes   []string
}

type callerKey struct{}

func withCaller(ctx context.Context, c *Caller) context.Context {
	return context.WithValue(ctx, callerKey{}, c)
}

func callerFromContext(ctx context.Context) *Caller {
	c, _ := ctx.Value(callerKey{}).(*Caller)
	return c
}

// Authenticator makes the MCP endpoint an OAuth resource server: it accepts
// only access tokens issued for Resource, and answers with the challenges of
// the MCP authorization spec so that clients can discover how to authorize.
type Authenticator struct {
	Resource            string
	ResourceMetadataURL string
	Verifier            TokenVerifier
}

func (a *Authenticator) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		rawToken, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			a.challenge(w, http.StatusUnauthorized, "", ScopesSupported, "")
			return
		}

		sessionInfo, token, err := a.Verifier.VerifyResourceAccessToken(ctx, rawToken, a.Resource)
		if err != nil {
			logger.GetLogger(ctx).Debug(ctx, "rejected MCP access token", slog.String("reason", err.Error()))
			a.challenge(w, http.StatusUnauthorized, "invalid_token", ScopesSupported, "The access token is invalid or expired")
			return
		}

		caller := &Caller{
			Session:  sessionInfo,
			ClientID: stringClaim(token, "client_id"),
			Scopes:   strings.Fields(stringClaim(token, "scope")),
		}

		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			if missing := missingScopes(caller.Scopes, calledTools(body)); len(missing) > 0 {
				a.challenge(w, http.StatusForbidden, "insufficient_scope", missing,
					fmt.Sprintf("This operation needs the scope %s", strings.Join(missing, " ")))
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(withCaller(ctx, caller)))
	})
}

func (a *Authenticator) challenge(w http.ResponseWriter, status int, errCode string, scopes []string, description string) {
	params := []string{fmt.Sprintf("resource_metadata=%q", a.ResourceMetadataURL)}
	if errCode != "" {
		params = append(params, fmt.Sprintf("error=%q", errCode))
	}
	if len(scopes) > 0 {
		params = append(params, fmt.Sprintf("scope=%q", strings.Join(scopes, " ")))
	}
	if description != "" {
		params = append(params, fmt.Sprintf("error_description=%q", description))
	}
	w.Header().Set("WWW-Authenticate", "Bearer "+strings.Join(params, ", "))
	http.Error(w, http.StatusText(status), status)
}

func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	return token, token != ""
}

func stringClaim(token jwt.Token, name string) string {
	v, ok := token.Get(name)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

type jsonrpcMessage struct {
	Method string `json:"method"`
	Params struct {
		Name string `json:"name"`
	} `json:"params"`
}

// calledTools returns the tools a JSON-RPC message (or batch) calls. A body
// that does not parse calls nothing here; the MCP handler rejects it.
func calledTools(body []byte) []string {
	var msgs []jsonrpcMessage
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &msgs); err != nil {
			return nil
		}
	} else {
		var msg jsonrpcMessage
		if err := json.Unmarshal(trimmed, &msg); err != nil {
			return nil
		}
		msgs = []jsonrpcMessage{msg}
	}

	var names []string
	for _, msg := range msgs {
		if msg.Method == "tools/call" {
			names = append(names, msg.Params.Name)
		}
	}
	return names
}

// missingScopes returns every scope the called tools need that was not granted,
// so that one challenge covers the whole operation.
func missingScopes(granted []string, tools []string) []string {
	var missing []string
	for _, name := range tools {
		t, ok := toolsByName[name]
		if !ok {
			continue
		}
		if !hasScope(granted, t.scope) && !slices.Contains(missing, t.scope) {
			missing = append(missing, t.scope)
		}
	}
	return missing
}

var errNoCaller = errors.New("mcp: no authenticated caller")
