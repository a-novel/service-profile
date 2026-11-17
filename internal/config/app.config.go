package config

import (
	"time"

	serviceauthentication "github.com/a-novel/service-authentication/v2/pkg/go"

	"github.com/a-novel-kit/golib/grpcf"
	"github.com/a-novel-kit/golib/httpf"
	"github.com/a-novel-kit/golib/logging"
	"github.com/a-novel-kit/golib/otel"
	"github.com/a-novel-kit/golib/postgres"
	"github.com/a-novel-kit/golib/servicetoken"
)

// Main holds the top-level application settings.
type Main struct {
	// Name of the application, as it appears in logs and tracing.
	Name string
	// DowntimeStart is when a planned downtime starts; nil when none is planned. From then until
	// it is removed, the service refuses work and leaves its database alone.
	DowntimeStart *time.Time
}

// Dependencies configures how the service reaches the services it calls.
type Dependencies struct {
	// ServiceJSONKeysHost is the hostname of the JSON Keys service, which verifies access tokens.
	ServiceJSONKeysHost string
	// ServiceJSONKeysPort is the gRPC port of the JSON Keys service.
	ServiceJSONKeysPort int
	// ServiceJSONKeysCredentials authenticates the calls to the JSON Keys service.
	ServiceJSONKeysCredentials grpcf.CredentialsProvider
}

// App is the complete configuration the service reads at startup.
type App struct {
	// App holds the top-level application settings.
	App Main
	// Dependencies configures the services the service calls.
	Dependencies Dependencies
	// Permissions maps each role to the permissions it grants.
	Permissions serviceauthentication.Permissions
	// Rest configures the REST server.
	Rest httpf.ServerConfig
	// ServiceToken names the service tokens the REST server accepts.
	ServiceToken servicetoken.Config
	// ServiceIssuerOptional lets the REST server start with no ServiceToken issuer, and take every
	// request without a service token. Without it, a ServiceToken with no issuer refuses to start.
	ServiceIssuerOptional bool

	// Otel configures the OpenTelemetry exporter.
	Otel otel.Config
	// Logger is the application logger the REST handlers report their errors to.
	Logger logging.Log
	// RestLogger configures the REST request logging middleware.
	RestLogger logging.HTTPConfig
	// Postgres configures the PostgreSQL connection.
	Postgres postgres.Config
}
