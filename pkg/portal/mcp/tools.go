package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

func pageSize(n int) int {
	switch {
	case n <= 0:
		return defaultPageSize
	case n > maxPageSize:
		return maxPageSize
	default:
		return n
	}
}

// userNodeID accepts a user ID as shown in the console, or its global ID.
func userNodeID(id string) string {
	if resolved := relay.FromGlobalID(id); resolved != nil && resolved.Type == "User" {
		return id
	}
	return relay.ToGlobalID("User", id)
}

func readOnly() *sdk.ToolAnnotations {
	return &sdk.ToolAnnotations{ReadOnlyHint: true}
}

func requireProject(project string) error {
	if strings.TrimSpace(project) == "" {
		return fmt.Errorf("project is required")
	}
	return nil
}

type noInput struct{}

type projectInput struct {
	Project string `json:"project" jsonschema:"Project ID, as listed by list_projects"`
}

func whoami(c *apiClient, _ noInput) (any, error) {
	var data struct {
		Viewer json.RawMessage `json:"viewer"`
	}
	err := c.portal(`query { viewer { id email formattedName } }`, nil, &data)
	if err != nil {
		return nil, err
	}
	return map[string]any{"viewer": data.Viewer, "granted_scopes": c.caller.Scopes}, nil
}

func listProjects(c *apiClient, _ noInput) (any, error) {
	var data struct {
		AppList []struct {
			AppID        string `json:"appID"`
			PublicOrigin string `json:"publicOrigin"`
		} `json:"appList"`
	}
	err := c.portal(`query { appList { appID publicOrigin } }`, nil, &data)
	if err != nil {
		return nil, err
	}
	projects := make([]map[string]any, len(data.AppList))
	for i, a := range data.AppList {
		projects[i] = map[string]any{"project": a.AppID, "public_origin": a.PublicOrigin}
	}
	return map[string]any{"projects": projects}, nil
}

func getProject(c *apiClient, in projectInput) (any, error) {
	if err := requireProject(in.Project); err != nil {
		return nil, err
	}
	var data struct {
		Node *struct {
			ID                 string          `json:"id"`
			PlanName           string          `json:"planName"`
			Viewer             json.RawMessage `json:"viewer"`
			Domains            json.RawMessage `json:"domains"`
			Collaborators      json.RawMessage `json:"collaborators"`
			EffectiveAppConfig json.RawMessage `json:"effectiveAppConfig"`
		} `json:"node"`
	}
	err := c.portal(`
query ($id: ID!) {
  node(id: $id) {
    ... on App {
      id
      planName
      viewer { role }
      domains { domain isCustom isVerified }
      collaborators { role user { email } }
      effectiveAppConfig
    }
  }
}`, map[string]any{"id": relay.ToGlobalID("App", in.Project)}, &data)
	if err != nil {
		return nil, err
	}
	if data.Node == nil {
		return nil, fmt.Errorf("project %q not found, or you are not a collaborator of it", in.Project)
	}

	summary, err := summarizeAppConfig(data.Node.EffectiveAppConfig)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"project":       in.Project,
		"plan":          data.Node.PlanName,
		"your_role":     data.Node.Viewer,
		"domains":       data.Node.Domains,
		"collaborators": data.Node.Collaborators,
		"config":        summary,
	}, nil
}

// summarizeAppConfig keeps the parts of authgear.yaml that describe how users
// sign in to the project. authgear.yaml holds no secrets, but most of it is
// noise for an agent.
func summarizeAppConfig(raw json.RawMessage) (map[string]any, error) {
	var cfg struct {
		HTTP struct {
			PublicOrigin   string   `json:"public_origin"`
			AllowedOrigins []string `json:"allowed_origins"`
		} `json:"http"`
		OAuth struct {
			Clients []struct {
				ClientID        string   `json:"client_id"`
				Name            string   `json:"name"`
				ApplicationType string   `json:"x_application_type"`
				RedirectURIs    []string `json:"redirect_uris"`
			} `json:"clients"`
		} `json:"oauth"`
		Authentication struct {
			Identities                  []string `json:"identities"`
			PrimaryAuthenticators       []string `json:"primary_authenticators"`
			SecondaryAuthenticators     []string `json:"secondary_authenticators"`
			SecondaryAuthenticationMode string   `json:"secondary_authentication_mode"`
		} `json:"authentication"`
		Identity struct {
			LoginID struct {
				Keys []struct {
					Type string `json:"type"`
				} `json:"keys"`
			} `json:"login_id"`
			OAuth struct {
				Providers []struct {
					Alias string `json:"alias"`
					Type  string `json:"type"`
				} `json:"providers"`
			} `json:"oauth"`
		} `json:"identity"`
		Localization struct {
			FallbackLanguage   string   `json:"fallback_language"`
			SupportedLanguages []string `json:"supported_languages"`
		} `json:"localization"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("unexpected project config: %w", err)
	}

	clients := make([]map[string]any, len(cfg.OAuth.Clients))
	for i, cl := range cfg.OAuth.Clients {
		clients[i] = map[string]any{
			"client_id":        cl.ClientID,
			"name":             cl.Name,
			"application_type": cl.ApplicationType,
			"redirect_uris":    cl.RedirectURIs,
		}
	}
	loginIDTypes := make([]string, len(cfg.Identity.LoginID.Keys))
	for i, k := range cfg.Identity.LoginID.Keys {
		loginIDTypes[i] = k.Type
	}
	providers := make([]map[string]any, len(cfg.Identity.OAuth.Providers))
	for i, p := range cfg.Identity.OAuth.Providers {
		providers[i] = map[string]any{"alias": p.Alias, "type": p.Type}
	}

	return map[string]any{
		"public_origin":                 cfg.HTTP.PublicOrigin,
		"allowed_origins":               cfg.HTTP.AllowedOrigins,
		"applications":                  clients,
		"identities":                    cfg.Authentication.Identities,
		"login_id_types":                loginIDTypes,
		"social_login_providers":        providers,
		"primary_authenticators":        cfg.Authentication.PrimaryAuthenticators,
		"secondary_authenticators":      cfg.Authentication.SecondaryAuthenticators,
		"secondary_authentication_mode": cfg.Authentication.SecondaryAuthenticationMode,
		"fallback_language":             cfg.Localization.FallbackLanguage,
		"supported_languages":           cfg.Localization.SupportedLanguages,
	}, nil
}

type chartsInput struct {
	Project    string `json:"project" jsonschema:"Project ID"`
	RangeFrom  string `json:"range_from" jsonschema:"First day, YYYY-MM-DD"`
	RangeTo    string `json:"range_to" jsonschema:"Last day, YYYY-MM-DD"`
	Periodical string `json:"periodical,omitempty" jsonschema:"Bucket size for active users: MONTHLY or WEEKLY (default MONTHLY)"`
}

func projectCharts(c *apiClient, in chartsInput) (any, error) {
	if err := requireProject(in.Project); err != nil {
		return nil, err
	}
	periodical := strings.ToUpper(in.Periodical)
	if periodical == "" {
		periodical = "MONTHLY"
	}
	if periodical != "MONTHLY" && periodical != "WEEKLY" {
		return nil, fmt.Errorf("periodical must be MONTHLY or WEEKLY")
	}
	var data json.RawMessage
	err := c.portal(`
query ($appID: ID!, $from: Date!, $to: Date!, $periodical: Periodical!) {
  activeUserChart(appID: $appID, periodical: $periodical, rangeFrom: $from, rangeTo: $to) { dataset { label data } }
  totalUserCountChart(appID: $appID, rangeFrom: $from, rangeTo: $to) { dataset { label data } }
  signupConversionRate(appID: $appID, rangeFrom: $from, rangeTo: $to) { totalSignup totalSignupUniquePageView conversionRate }
}`, map[string]any{
		"appID":      relay.ToGlobalID("App", in.Project),
		"from":       in.RangeFrom,
		"to":         in.RangeTo,
		"periodical": periodical,
	}, &data)
	if err != nil {
		return nil, err
	}
	return map[string]any{"project": in.Project, "charts": data}, nil
}

const userSummaryFields = `id createdAt lastLoginAt isDisabled disableReason isDeactivated isAnonymous standardAttributes`

type searchUsersInput struct {
	Project string `json:"project" jsonschema:"Project ID"`
	Keyword string `json:"keyword,omitempty" jsonschema:"Matches email, phone number, username, user ID or name; empty lists all users"`
	SortBy  string `json:"sort_by,omitempty" jsonschema:"CREATED_AT or LAST_LOGIN_AT (newest first)"`
	First   int    `json:"first,omitempty" jsonschema:"Page size, default 20, at most 100"`
	After   string `json:"after,omitempty" jsonschema:"Cursor from the previous page's end_cursor"`
}

func searchUsers(c *apiClient, in searchUsersInput) (any, error) {
	if err := requireProject(in.Project); err != nil {
		return nil, err
	}
	vars := map[string]any{"first": pageSize(in.First)}
	if in.Keyword != "" {
		vars["keyword"] = in.Keyword
	}
	if in.After != "" {
		vars["after"] = in.After
	}
	if in.SortBy != "" {
		sortBy := strings.ToUpper(in.SortBy)
		if sortBy != "CREATED_AT" && sortBy != "LAST_LOGIN_AT" {
			return nil, fmt.Errorf("sort_by must be CREATED_AT or LAST_LOGIN_AT")
		}
		vars["sortBy"] = sortBy
		vars["sortDirection"] = "DESC"
	}
	var data struct {
		Users connection `json:"users"`
	}
	err := c.admin(in.Project, `
query ($first: Int, $after: String, $keyword: String, $sortBy: UserSortBy, $sortDirection: SortDirection) {
  users(first: $first, after: $after, searchKeyword: $keyword, sortBy: $sortBy, sortDirection: $sortDirection) {
    totalCount
    pageInfo { hasNextPage endCursor }
    edges { node { `+userSummaryFields+` } }
  }
}`, vars, &data)
	if err != nil {
		return nil, err
	}
	return data.Users.page("users"), nil
}

type getUserInput struct {
	Project      string `json:"project" jsonschema:"Project ID"`
	UserID       string `json:"user_id,omitempty" jsonschema:"User ID; give this or login_id"`
	LoginIDKey   string `json:"login_id_key,omitempty" jsonschema:"Login ID key such as email, phone or username; used with login_id"`
	LoginIDValue string `json:"login_id,omitempty" jsonschema:"Login ID value, e.g. an email address"`
}

const userDetailFields = userSummaryFields + `
  updatedAt
  accountLockout { isLocked lockedUntil }
  loginIDs { type claims }
  oauthConnections { type claims }
  roles(first: 100) { edges { node { key name } } }
  groups(first: 100) { edges { node { key name } } }
  primaryPassword { id }
  secondaryTOTPAuthenticators { id }`

func getUser(c *apiClient, in getUserInput) (any, error) {
	if err := requireProject(in.Project); err != nil {
		return nil, err
	}
	var user json.RawMessage
	switch {
	case in.UserID != "":
		var data struct {
			Node json.RawMessage `json:"node"`
		}
		err := c.admin(in.Project, `query ($id: ID!) { node(id: $id) { ... on User { `+userDetailFields+` } } }`,
			map[string]any{"id": userNodeID(in.UserID)}, &data)
		if err != nil {
			return nil, err
		}
		user = data.Node
	case in.LoginIDValue != "":
		key := in.LoginIDKey
		if key == "" {
			key = "email"
		}
		var data struct {
			User json.RawMessage `json:"getUserByLoginID"`
		}
		err := c.admin(in.Project, `query ($key: String!, $value: String!) { getUserByLoginID(loginIDKey: $key, loginIDValue: $value) { `+userDetailFields+` } }`,
			map[string]any{"key": key, "value": in.LoginIDValue}, &data)
		if err != nil {
			return nil, err
		}
		user = data.User
	default:
		return nil, fmt.Errorf("give user_id or login_id")
	}
	if isJSONNull(user) {
		return nil, fmt.Errorf("user not found")
	}
	return map[string]any{"user": user}, nil
}

type userInput struct {
	Project string `json:"project" jsonschema:"Project ID"`
	UserID  string `json:"user_id" jsonschema:"User ID"`
}

func listUserSessions(c *apiClient, in userInput) (any, error) {
	if err := requireProject(in.Project); err != nil {
		return nil, err
	}
	if in.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	var data struct {
		Node *struct {
			Sessions connection `json:"sessions"`
		} `json:"node"`
	}
	err := c.admin(in.Project, `
query ($id: ID!) {
  node(id: $id) {
    ... on User {
      sessions(first: 100) {
        totalCount
        pageInfo { hasNextPage endCursor }
        edges { node { id type displayName applicationName clientID createdAt lastAccessedAt lastAccessedByIP lastAccessedByIPEnglishCountryName userAgent } }
      }
    }
  }
}`, map[string]any{"id": userNodeID(in.UserID)}, &data)
	if err != nil {
		return nil, err
	}
	if data.Node == nil {
		return nil, fmt.Errorf("user not found")
	}
	return data.Node.Sessions.page("sessions"), nil
}

func listRolesGroups(c *apiClient, in projectInput) (any, error) {
	if err := requireProject(in.Project); err != nil {
		return nil, err
	}
	var data struct {
		Roles  connection `json:"roles"`
		Groups connection `json:"groups"`
	}
	err := c.admin(in.Project, `
query {
  roles(first: 100) { totalCount pageInfo { hasNextPage endCursor } edges { node { id key name description } } }
  groups(first: 100) { totalCount pageInfo { hasNextPage endCursor } edges { node { id key name description } } }
}`, nil, &data)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"roles":  data.Roles.page("items"),
		"groups": data.Groups.page("items"),
	}, nil
}

type auditLogInput struct {
	Project       string   `json:"project" jsonschema:"Project ID"`
	RangeFrom     string   `json:"range_from,omitempty" jsonschema:"Start time, RFC 3339"`
	RangeTo       string   `json:"range_to,omitempty" jsonschema:"End time, RFC 3339"`
	ActivityTypes []string `json:"activity_types,omitempty" jsonschema:"Activity types such as USER_AUTHENTICATED or USER_SIGNED_UP"`
	UserID        string   `json:"user_id,omitempty" jsonschema:"Only entries about this user"`
	IncludeData   bool     `json:"include_data,omitempty" jsonschema:"Include each entry's full payload"`
	First         int      `json:"first,omitempty" jsonschema:"Page size, default 20, at most 100"`
	After         string   `json:"after,omitempty" jsonschema:"Cursor from the previous page's end_cursor"`
}

func queryAuditLog(c *apiClient, in auditLogInput) (any, error) {
	if err := requireProject(in.Project); err != nil {
		return nil, err
	}
	vars := map[string]any{"first": pageSize(in.First), "includeData": in.IncludeData}
	if in.After != "" {
		vars["after"] = in.After
	}
	if in.RangeFrom != "" {
		vars["from"] = in.RangeFrom
	}
	if in.RangeTo != "" {
		vars["to"] = in.RangeTo
	}
	if len(in.ActivityTypes) > 0 {
		types := make([]string, len(in.ActivityTypes))
		for i, t := range in.ActivityTypes {
			types[i] = strings.ToUpper(t)
		}
		vars["types"] = types
	}
	if in.UserID != "" {
		vars["userIDs"] = []string{userNodeID(in.UserID)}
	}
	var data struct {
		AuditLogs connection `json:"auditLogs"`
	}
	err := c.admin(in.Project, `
query ($first: Int, $after: String, $from: DateTime, $to: DateTime, $types: [AuditLogActivityType!], $userIDs: [ID!], $includeData: Boolean!) {
  auditLogs(first: $first, after: $after, rangeFrom: $from, rangeTo: $to, activityTypes: $types, userIDs: $userIDs, sortDirection: DESC) {
    totalCount
    pageInfo { hasNextPage endCursor }
    edges { node { id activityType createdAt clientID ipAddress userAgent user { id } data @include(if: $includeData) } }
  }
}`, vars, &data)
	if err != nil {
		return nil, err
	}
	return data.AuditLogs.page("entries"), nil
}

type connection struct {
	TotalCount *int `json:"totalCount"`
	PageInfo   struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
	Edges []struct {
		Node json.RawMessage `json:"node"`
	} `json:"edges"`
}

func (c connection) page(key string) map[string]any {
	nodes := make([]json.RawMessage, len(c.Edges))
	for i, e := range c.Edges {
		nodes[i] = e.Node
	}
	out := map[string]any{
		key:             nodes,
		"has_next_page": c.PageInfo.HasNextPage,
	}
	if c.PageInfo.HasNextPage {
		out["end_cursor"] = c.PageInfo.EndCursor
	}
	if c.TotalCount != nil {
		out["total_count"] = *c.TotalCount
	}
	return out
}

func isJSONNull(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s == "" || s == "null"
}
