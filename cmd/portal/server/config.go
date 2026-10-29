package server

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/kelseyhightower/envconfig"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Config struct {
	// ListenAddr sets the listen address of the portal server.
	PortalListenAddr         string `envconfig:"PORTAL_LISTEN_ADDR" default:"0.0.0.0:3003"`
	PortalInternalListenAddr string `envconfig:"PORTAL_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13003"`
	// SiteadminListenAddr sets the listen address of the Site Admin API server.
	SiteadminListenAddr         string `envconfig:"SITEADMIN_LISTEN_ADDR" default:"0.0.0.0:3005"`
	SiteadminInternalListenAddr string `envconfig:"SITEADMIN_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13005"`
	// MCPListenAddr sets the listen address of the MCP server.
	MCPListenAddr         string `envconfig:"MCP_LISTEN_ADDR" default:"0.0.0.0:3006"`
	MCPInternalListenAddr string `envconfig:"MCP_INTERNAL_LISTEN_ADDR" default:"0.0.0.0:13006"`
	// MCPResource is the canonical URI of the MCP endpoint, e.g. https://mcp.example.com/mcp.
	// Access tokens must be issued for it (RFC 8707) by the project the portal signs in with.
	MCPResource string `envconfig:"MCP_RESOURCE"`
	// ConfigSource configures the source of app configurations
	ConfigSource *configsource.Config `envconfig:"CONFIG_SOURCE"`
	// Authgear configures Authgear acting as authentication server for the portal.
	Authgear portalconfig.AuthgearConfig `envconfig:"AUTHGEAR"`
	// SiteadminAuthgear configures Authgear for the Site Admin API.
	// Allows the siteadmin server to authenticate against a different Authgear app than the portal.
	SiteadminAuthgear portalconfig.AuthgearConfig `envconfig:"SITEADMIN_AUTHGEAR"`
	// AdminAPI configures how portal interacts with Authgear Admin API.
	AdminAPI portalconfig.AdminAPIConfig `envconfig:"ADMIN_API"`
	// App configures the managed apps.
	App portalconfig.AppConfig `envconfig:"APP"`
	// SMTP configures SMTP.
	SMTP portalconfig.SMTPConfig `envconfig:"SMTP"`
	// Mail configures email settings.
	Mail portalconfig.MailConfig `envconfig:"MAIL"`

	// PORTAL_BUILTIN_RESOURCE_DIRECTORY is deprecated. It has no effect anymore.

	// CustomResourceDirectory sets the directory for customized resource files
	CustomResourceDirectory string `envconfig:"PORTAL_CUSTOM_RESOURCE_DIRECTORY"`

	// DomainImplementation indicates the domain implementation, only none or kubernetes are supported
	// if it sets to kubernetes, kubernetes resources will be created based on
	// APP_KUBERNETES_INGRESS_TEMPLATE_FILE when creating domains
	DomainImplementation portalconfig.DomainImplementationType `envconfig:"DOMAIN_IMPLEMENTATION"`

	// Kubernetes set the kubernetes related config if the portal is deployed in k8s
	// One of the purpose is for creating ingress when creating new domain
	Kubernetes portalconfig.KubernetesConfig `envconfig:"KUBERNETES"`

	// Search sets search related config.
	Search portalconfig.SearchConfig `envconfig:"SEARCH"`

	// AuditLog sets audit log related config.
	AuditLog portalconfig.AuditLogConfig `envconfig:"AUDIT_LOG"`

	// Analytic sets analytic dashboard related config.
	Analytic config.AnalyticConfig `envconfig:"ANALYTIC"`

	Stripe portalconfig.StripeConfig `envconfig:"STRIPE"`

	Osano portalconfig.OsanoConfig `envconfig:"OSANO"`

	GoogleTagManager portalconfig.GoogleTagManagerConfig `envconfig:"GTM"`

	PortalFrontendSentry portalconfig.PortalFrontendSentryConfig `envconfig:"PORTAL_FRONTEND_SENTRY"`
	PortalFeatures       portalconfig.PortalFeaturesConfig       `envconfig:"PORTAL_FEATURES"`

	*config.EnvironmentConfig
}

type LoadConfigOptions struct {
	ServePortal    bool
	ServeSiteadmin bool
	ServeMCP       bool
}

func LoadConfigFromEnv(opts LoadConfigOptions) (*Config, error) {
	config := &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, fmt.Errorf("cannot load server config: %w", err)
	}

	err = config.Validate(opts)
	if err != nil {
		return nil, fmt.Errorf("invalid server config: %w", err)
	}

	return config, nil
}

func (c *Config) Validate(opts LoadConfigOptions) error {
	ctx := &validation.Context{}

	sourceTypes := make([]string, len(configsource.Types))
	ok := false
	for i, t := range configsource.Types {
		if t == c.ConfigSource.Type {
			ok = true
			break
		}
		sourceTypes[i] = string(t)
	}
	if !ok {
		ctx.Child("CONFIG_SOURCE_TYPE").EmitErrorMessage(
			"invalid configuration source type; available: " + strings.Join(sourceTypes, ", "),
		)
	}

	if c.GlobalDatabase.DatabaseURL == "" {
		ctx.Child("DATABASE_URL").EmitErrorMessage("missing database URL")
	}

	if opts.ServePortal {
		if c.Authgear.ClientID == "" {
			ctx.Child("AUTHGEAR_CLIENT_ID").EmitErrorMessage("missing authgear client ID")
		}
		if c.Authgear.Endpoint == "" {
			ctx.Child("AUTHGEAR_ENDPOINT").EmitErrorMessage("missing authgear endpoint")
		}
	}

	if opts.ServeSiteadmin {
		if c.SiteadminAuthgear.AppID == "" {
			ctx.Child("SITEADMIN_AUTHGEAR_APP_ID").EmitErrorMessage("missing siteadmin authgear app ID")
		}
		if c.SiteadminAuthgear.Endpoint == "" {
			ctx.Child("SITEADMIN_AUTHGEAR_ENDPOINT").EmitErrorMessage("missing siteadmin authgear endpoint")
		}
	}

	if opts.ServeMCP {
		if c.Authgear.Endpoint == "" {
			ctx.Child("AUTHGEAR_ENDPOINT").EmitErrorMessage("missing authgear endpoint")
		}
		if err := validateMCPResource(c.MCPResource); err != nil {
			ctx.Child("MCP_RESOURCE").EmitErrorMessage(err.Error())
		}
	}

	// Stripe config is optional

	return ctx.Error("invalid server configuration")
}

// validateMCPResource checks the canonical server URI rules of the MCP
// authorization spec: absolute https (or http for local development), no fragment.
func validateMCPResource(resource string) error {
	if resource == "" {
		return fmt.Errorf("missing MCP resource URI")
	}
	u, err := url.Parse(resource)
	if err != nil {
		return fmt.Errorf("invalid MCP resource URI: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return fmt.Errorf("MCP resource URI must be an absolute http(s) URI")
	}
	if u.Fragment != "" || u.RawQuery != "" {
		return fmt.Errorf("MCP resource URI must not have a query or fragment")
	}
	if strings.HasSuffix(u.Path, "/") {
		return fmt.Errorf("MCP resource URI must not end with a slash")
	}
	return nil
}
