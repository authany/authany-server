package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/portal/session"
)

const (
	testResource    = "https://mcp.example.com/mcp"
	testMetadataURL = "https://mcp.example.com/.well-known/oauth-protected-resource/mcp"
)

type fakeVerifier struct {
	scope string
}

func (v fakeVerifier) VerifyResourceAccessToken(_ context.Context, rawToken string, resourceURI string) (*model.SessionInfo, jwt.Token, error) {
	if rawToken != "good" || resourceURI != testResource {
		return nil, nil, errors.New("bad token")
	}
	token := jwt.New()
	_ = token.Set("client_id", "agent")
	_ = token.Set("scope", v.scope)
	return &model.SessionInfo{IsValid: true, UserID: "user1"}, token, nil
}

func mcpRequest(token string, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json, text/event-stream")
	r.Header.Set("MCP-Protocol-Version", "2025-06-18")
	return r
}

func toolCall(name string, args string) string {
	return `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"` + name + `","arguments":` + args + `}}`
}

func TestHasScope(t *testing.T) {
	Convey("hasScope", t, func() {
		So(hasScope(nil, ""), ShouldBeTrue)
		So(hasScope([]string{ScopeUsersRead}, ScopeUsersRead), ShouldBeTrue)
		So(hasScope([]string{ScopeUsersWrite}, ScopeUsersRead), ShouldBeTrue)
		So(hasScope([]string{ScopeUsersRead}, ScopeUsersWrite), ShouldBeFalse)
		So(hasScope([]string{ScopeProjectsWrite}, ScopeUsersRead), ShouldBeFalse)
	})
}

func TestAuthenticator(t *testing.T) {
	Convey("Authenticator", t, func() {
		var reached *Caller
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reached = callerFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})
		handler := func(scope string) http.Handler {
			a := &Authenticator{
				Resource:            testResource,
				ResourceMetadataURL: testMetadataURL,
				Verifier:            fakeVerifier{scope: scope},
			}
			return a.Handle(next)
		}

		Convey("asks for authorization when there is no token", func() {
			w := httptest.NewRecorder()
			handler("").ServeHTTP(w, mcpRequest("", `{}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
			So(w.Header().Get("WWW-Authenticate"), ShouldEqual,
				`Bearer resource_metadata="`+testMetadataURL+`", scope="projects:read users:read"`)
			So(reached, ShouldBeNil)
		})

		Convey("rejects an invalid token", func() {
			w := httptest.NewRecorder()
			handler("").ServeHTTP(w, mcpRequest("bad", `{}`))
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
			So(w.Header().Get("WWW-Authenticate"), ShouldContainSubstring, `error="invalid_token"`)
			So(reached, ShouldBeNil)
		})

		Convey("passes the caller on", func() {
			w := httptest.NewRecorder()
			handler("projects:read").ServeHTTP(w, mcpRequest("good", toolCall("list_projects", `{}`)))
			So(w.Code, ShouldEqual, http.StatusOK)
			So(reached.Session.UserID, ShouldEqual, "user1")
			So(reached.ClientID, ShouldEqual, "agent")
			So(reached.Scopes, ShouldResemble, []string{"projects:read"})
		})

		Convey("lets tools that need no scope through", func() {
			w := httptest.NewRecorder()
			handler("").ServeHTTP(w, mcpRequest("good", toolCall("whoami", `{}`)))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("answers a tool call without its scope with a step-up challenge", func() {
			w := httptest.NewRecorder()
			handler("projects:read").ServeHTTP(w, mcpRequest("good", toolCall("search_users", `{"project":"p"}`)))
			So(w.Code, ShouldEqual, http.StatusForbidden)
			h := w.Header().Get("WWW-Authenticate")
			So(h, ShouldContainSubstring, `resource_metadata="`+testMetadataURL+`"`)
			So(h, ShouldContainSubstring, `error="insufficient_scope"`)
			So(h, ShouldContainSubstring, `scope="users:read"`)
			So(reached, ShouldBeNil)
		})

		Convey("names every missing scope of a batch in one challenge", func() {
			w := httptest.NewRecorder()
			body := "[" + toolCall("search_users", `{}`) + "," + toolCall("get_project", `{}`) + "]"
			handler("").ServeHTTP(w, mcpRequest("good", body))
			So(w.Code, ShouldEqual, http.StatusForbidden)
			So(w.Header().Get("WWW-Authenticate"), ShouldContainSubstring, `scope="users:read projects:read"`)
		})

		Convey("accepts a broader scope", func() {
			w := httptest.NewRecorder()
			handler("users:write").ServeHTTP(w, mcpRequest("good", toolCall("search_users", `{"project":"p"}`)))
			So(w.Code, ShouldEqual, http.StatusOK)
		})

		Convey("leaves non-tool requests to the MCP handler", func() {
			w := httptest.NewRecorder()
			handler("").ServeHTTP(w, mcpRequest("good", `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}

// fakeAPI stands in for the portal's GraphQL and Admin API routes.
type fakeAPI struct {
	lastPath    string
	lastQuery   string
	lastVars    map[string]any
	lastSession *model.SessionInfo
	response    string
}

func (f *fakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.lastPath = r.URL.Path
	f.lastSession = session.GetValidSessionInfo(r.Context())
	var body struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.lastQuery = body.Query
	f.lastVars = body.Variables
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, f.response)
}

type toolResult struct {
	Result struct {
		IsError           bool            `json:"isError"`
		StructuredContent json.RawMessage `json:"structuredContent"`
		Content           []struct {
			Text string `json:"text"`
		} `json:"content"`
		Tools []struct {
			Name        string `json:"name"`
			Annotations struct {
				ReadOnlyHint bool `json:"readOnlyHint"`
			} `json:"annotations"`
		} `json:"tools"`
	} `json:"result"`
}

func TestHandler(t *testing.T) {
	Convey("MCP handler", t, func() {
		api := &fakeAPI{}
		a := &Authenticator{
			Resource:            testResource,
			ResourceMetadataURL: testMetadataURL,
			Verifier:            fakeVerifier{scope: "projects:read users:read"},
		}
		h := a.Handle(NewHandler(api))
		call := func(body string) toolResult {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, mcpRequest("good", body))
			So(w.Code, ShouldEqual, http.StatusOK)
			var res toolResult
			So(json.Unmarshal(w.Body.Bytes(), &res), ShouldBeNil)
			return res
		}

		Convey("lists the read-only tools", func() {
			res := call(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
			var names []string
			for _, tool := range res.Result.Tools {
				names = append(names, tool.Name)
				So(tool.Annotations.ReadOnlyHint, ShouldBeTrue)
			}
			So(names, ShouldContain, "list_projects")
			So(names, ShouldContain, "query_audit_log")
			So(len(names), ShouldEqual, len(tools))
		})

		Convey("runs portal queries as the caller", func() {
			api.response = `{"data":{"appList":[{"appID":"myapp","publicOrigin":"https://myapp.example.com"}]}}`
			res := call(toolCall("list_projects", `{}`))
			So(res.Result.IsError, ShouldBeFalse)
			So(api.lastPath, ShouldEqual, "/api/graphql")
			So(api.lastSession.UserID, ShouldEqual, "user1")
			So(string(res.Result.StructuredContent), ShouldEqual,
				`{"projects":[{"project":"myapp","public_origin":"https://myapp.example.com"}]}`)
		})

		Convey("runs Admin API queries against the project", func() {
			api.response = `{"data":{"users":{"totalCount":1,"pageInfo":{"hasNextPage":false,"endCursor":"c"},"edges":[{"node":{"id":"VXNlcjox"}}]}}}`
			res := call(toolCall("search_users", `{"project":"myapp","keyword":"alice@example.com"}`))
			So(res.Result.IsError, ShouldBeFalse)
			So(api.lastPath, ShouldEqual, "/api/apps/"+relay.ToGlobalID("App", "myapp")+"/graphql")
			So(api.lastVars["keyword"], ShouldEqual, "alice@example.com")
			So(api.lastVars["first"], ShouldEqual, float64(defaultPageSize))
			So(string(res.Result.StructuredContent), ShouldEqual,
				`{"has_next_page":false,"total_count":1,"users":[{"id":"VXNlcjox"}]}`)
		})

		Convey("accepts a plain user ID", func() {
			api.response = `{"data":{"node":{"id":"x"}}}`
			call(toolCall("get_user", `{"project":"myapp","user_id":"0b1c"}`))
			So(api.lastVars["id"], ShouldEqual, relay.ToGlobalID("User", "0b1c"))
		})

		Convey("reports GraphQL errors as tool errors", func() {
			api.response = `{"errors":[{"message":"app not found"}]}`
			res := call(toolCall("get_project", `{"project":"nope"}`))
			So(res.Result.IsError, ShouldBeTrue)
			So(res.Result.Content[0].Text, ShouldContainSubstring, "app not found")
		})

		Convey("summarizes the project config without unrelated parts", func() {
			api.response = `{"data":{"node":{"id":"x","planName":"free","viewer":{"role":"OWNER"},"domains":[],"collaborators":[],
				"effectiveAppConfig":{"http":{"public_origin":"https://myapp.example.com"},"oauth":{"clients":[{"client_id":"c1","name":"Web","x_application_type":"spa","redirect_uris":["https://myapp.example.com/cb"]}]},"hook":{"blocking_handlers":[{"url":"https://hooks.example.com"}]}}}}}`
			res := call(toolCall("get_project", `{"project":"myapp"}`))
			So(res.Result.IsError, ShouldBeFalse)
			out := string(res.Result.StructuredContent)
			So(out, ShouldContainSubstring, `"public_origin":"https://myapp.example.com"`)
			So(out, ShouldContainSubstring, `"client_id":"c1"`)
			So(out, ShouldNotContainSubstring, "hooks.example.com")
		})
	})
}
