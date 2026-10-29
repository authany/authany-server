package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/portal/session"
)

// apiClient runs portal GraphQL and Admin API operations in process, as the
// caller. It goes through the same handlers as the console, so access control
// and audit logging are the console's; the caller's MCP token is never passed on.
type apiClient struct {
	handler http.Handler
	ctx     context.Context
	caller  *Caller
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *apiClient) portal(query string, variables map[string]any, out any) error {
	return c.do("/api/graphql", query, variables, out)
}

func (c *apiClient) admin(projectID string, query string, variables map[string]any, out any) error {
	return c.do("/api/apps/"+relay.ToGlobalID("App", projectID)+"/graphql", query, variables, out)
}

func (c *apiClient) do(path string, query string, variables map[string]any, out any) error {
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return err
	}

	ctx := session.WithSessionInfo(c.ctx, c.caller.Session)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c.handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		return fmt.Errorf("you do not have access to this project")
	}
	if rec.Code != http.StatusOK {
		return fmt.Errorf("request failed with status %d: %s", rec.Code, strings.TrimSpace(truncate(rec.Body.String(), 500)))
	}

	var resp graphQLResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		return fmt.Errorf("unexpected response: %w", err)
	}
	if len(resp.Errors) > 0 {
		msgs := make([]string, len(resp.Errors))
		for i, e := range resp.Errors {
			msgs[i] = e.Message
		}
		return fmt.Errorf("%s", strings.Join(msgs, "; "))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(resp.Data, out)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
