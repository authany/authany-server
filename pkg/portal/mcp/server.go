package mcp

import (
	"context"
	"log/slog"
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/authgear/authgear-server/pkg/version"
)

type toolDef struct {
	tool  *sdk.Tool
	scope string
	add   func(s *sdk.Server, c *apiClient, t *sdk.Tool)
}

func define[In any](t *sdk.Tool, scope string, h func(c *apiClient, in In) (any, error)) toolDef {
	return toolDef{
		tool:  t,
		scope: scope,
		add: func(s *sdk.Server, c *apiClient, t *sdk.Tool) {
			sdk.AddTool(s, t, func(ctx context.Context, _ *sdk.CallToolRequest, in In) (*sdk.CallToolResult, any, error) {
				out, err := h(c, in)
				logger.GetLogger(ctx).Info(ctx, "mcp tool call",
					slog.String("tool", t.Name),
					slog.String("user_id", c.caller.Session.UserID),
					slog.String("client_id", c.caller.ClientID),
					slog.Bool("ok", err == nil),
				)
				return nil, out, err
			})
		},
	}
}

var tools = []toolDef{
	define(&sdk.Tool{
		Name:        "whoami",
		Title:       "Who am I",
		Description: "Show the signed-in Authany account and the scopes this client was granted.",
		Annotations: readOnly(),
	}, "", whoami),
	define(&sdk.Tool{
		Name:        "list_projects",
		Title:       "List projects",
		Description: "List the projects you are a collaborator of.",
		Annotations: readOnly(),
	}, ScopeProjectsRead, listProjects),
	define(&sdk.Tool{
		Name:        "get_project",
		Title:       "Get project",
		Description: "Show a project's plan, domains, collaborators, applications and sign-in methods. Secrets are never included.",
		Annotations: readOnly(),
	}, ScopeProjectsRead, getProject),
	define(&sdk.Tool{
		Name:        "project_charts",
		Title:       "Project analytics",
		Description: "Active users, total users and sign-up conversion of a project over a date range.",
		Annotations: readOnly(),
	}, ScopeProjectsRead, projectCharts),
	define(&sdk.Tool{
		Name:        "search_users",
		Title:       "Search users",
		Description: "Search or list the users of a project.",
		Annotations: readOnly(),
	}, ScopeUsersRead, searchUsers),
	define(&sdk.Tool{
		Name:        "get_user",
		Title:       "Get user",
		Description: "Show a user of a project by user ID or login ID: status, login IDs, social logins, roles, groups and lockout.",
		Annotations: readOnly(),
	}, ScopeUsersRead, getUser),
	define(&sdk.Tool{
		Name:        "list_user_sessions",
		Title:       "List user sessions",
		Description: "List the active sessions of a user of a project.",
		Annotations: readOnly(),
	}, ScopeUsersRead, listUserSessions),
	define(&sdk.Tool{
		Name:        "list_roles_groups",
		Title:       "List roles and groups",
		Description: "List the roles and groups defined in a project.",
		Annotations: readOnly(),
	}, ScopeUsersRead, listRolesGroups),
	define(&sdk.Tool{
		Name:        "query_audit_log",
		Title:       "Query audit log",
		Description: "Query a project's audit log, newest first, optionally by time range, activity type or user.",
		Annotations: readOnly(),
	}, ScopeUsersRead, queryAuditLog),
}

var toolsByName = func() map[string]toolDef {
	m := make(map[string]toolDef, len(tools))
	for _, t := range tools {
		m[t.tool.Name] = t
	}
	return m
}()

var schemaCache = sdk.NewSchemaCache()

// NewHandler serves MCP over stateless Streamable HTTP. api serves the portal
// GraphQL and Admin API routes without their session middleware; the caller set
// by Authenticator is the session.
func NewHandler(api http.Handler) http.Handler {
	return sdk.NewStreamableHTTPHandler(func(r *http.Request) *sdk.Server {
		caller := callerFromContext(r.Context())
		if caller == nil {
			panic(errNoCaller)
		}
		client := &apiClient{handler: api, ctx: r.Context(), caller: caller}

		s := sdk.NewServer(&sdk.Implementation{
			Name:    "authany",
			Title:   "Authany",
			Version: version.Version,
		}, &sdk.ServerOptions{
			Instructions: "Manage your Authany projects: projects, users, sessions, roles, groups and audit logs. " +
				"You act as the signed-in user and can only reach projects they collaborate on.",
			SchemaCache: schemaCache,
		})
		for _, t := range tools {
			t.add(s, client, t.tool)
		}
		return s
	}, &sdk.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	})
}
