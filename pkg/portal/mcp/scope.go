package mcp

import "slices"

// Scopes of the MCP resource. They must be defined on the resource in the
// project that authorizes MCP clients.
const (
	ScopeProjectsRead  = "projects:read"
	ScopeProjectsWrite = "projects:write"
	ScopeUsersRead     = "users:read"
	ScopeUsersWrite    = "users:write"
)

// ScopesSupported is the minimal set a client needs to start; write scopes are
// requested later through step-up authorization.
var ScopesSupported = []string{ScopeProjectsRead, ScopeUsersRead}

// impliedBy maps a scope to the broader scope that also grants it.
var impliedBy = map[string]string{
	ScopeProjectsRead: ScopeProjectsWrite,
	ScopeUsersRead:    ScopeUsersWrite,
}

func hasScope(granted []string, required string) bool {
	if required == "" || slices.Contains(granted, required) {
		return true
	}
	broader, ok := impliedBy[required]
	return ok && slices.Contains(granted, broader)
}
