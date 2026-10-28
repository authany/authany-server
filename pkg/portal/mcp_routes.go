package portal

import (
	"net/http"
	"net/url"
	"path"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/mcp"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/portal/transport"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// NewMCPRouter serves the MCP endpoint at the path of resourceURI, and its
// OAuth protected resource metadata (RFC 9728). The authorization server is the
// project the portal signs in with (AUTHGEAR_ENDPOINT).
func NewMCPRouter(p *deps.RootProvider, resourceURI string) (http.Handler, error) {
	resource, err := url.Parse(resourceURI)
	if err != nil {
		return nil, err
	}
	endpointPath := resource.Path
	if endpointPath == "" {
		endpointPath = "/"
	}
	// RFC 9728 section 3.1: the well-known path, followed by the resource's path.
	metadataPath := path.Join("/.well-known/oauth-protected-resource", endpointPath)
	metadataURL := (&url.URL{Scheme: resource.Scheme, Host: resource.Host, Path: metadataPath}).String()

	router := httproute.NewRouter()
	router.Health(p.Handler(newHealthzHandler))

	rootChain := httproute.Chain(
		p.Middleware(newOtelMiddleware),
		p.Middleware(newPanicMiddleware),
		p.Middleware(newBodyLimitMiddleware),
		p.Middleware(newSentryMiddleware),
	)

	metadata := auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
		Resource:               resourceURI,
		AuthorizationServers:   []string{p.AuthgearConfig.Endpoint},
		ScopesSupported:        mcp.ScopesSupported,
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "Authany",
	})
	metadataRoute := httproute.Route{Middleware: rootChain}.WithMethods("GET", "OPTIONS")
	router.Add(metadataRoute.WithPathPattern(metadataPath), metadata)
	if metadataPath != "/.well-known/oauth-protected-resource" {
		router.Add(metadataRoute.WithPathPattern("/.well-known/oauth-protected-resource"), metadata)
	}

	authenticator := &mcp.Authenticator{
		Resource:            resourceURI,
		ResourceMetadataURL: metadataURL,
		Verifier: &session.SessionInfoMiddleware{
			AuthgearConfig: p.AuthgearConfig,
			HTTPClient:     session.NewHTTPClient(),
			Clock:          clock.NewSystemClock(),
		},
	}
	mcpChain := httproute.Chain(
		rootChain,
		httproute.MiddlewareFunc(httputil.NoStore),
		authenticator,
	)
	mcpRoute := httproute.Route{Middleware: mcpChain}.
		WithMethods("GET", "POST", "DELETE").
		WithPathPattern(endpointPath)
	router.Add(mcpRoute, mcp.NewHandler(newMCPAPIRouter(p)))

	return router.HTTPHandler(), nil
}

// newMCPAPIRouter serves the console's GraphQL and Admin API routes in process
// for the MCP tools. It has no session middleware: the MCP authenticator has
// already put the caller's session in the request context.
func newMCPAPIRouter(p *deps.RootProvider) http.Handler {
	router := httproute.NewRouter()
	route := httproute.Route{Middleware: httproute.Chain()}
	router.Add(transport.ConfigureGraphQLRoute(route), p.Handler(newGraphQLHandler))
	router.Add(transport.ConfigureAdminAPIRoute(route), p.Handler(newAdminAPIHandler))
	return router.HTTPHandler()
}
