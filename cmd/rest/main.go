// Command rest serves the public REST API, for any HTTP client. cmd/migrations applies the database
// schema.
package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/samber/lo"

	serviceauthentication "github.com/a-novel/service-authentication/v2/pkg/go"
	servicejsonkeys "github.com/a-novel/service-json-keys/v2/pkg/go"

	"github.com/a-novel-kit/golib/downtime"
	"github.com/a-novel-kit/golib/httpf"
	"github.com/a-novel-kit/golib/otel"
	"github.com/a-novel-kit/golib/postgres"

	"github.com/a-novel/service-profile/internal/config"
	"github.com/a-novel/service-profile/internal/config/env"
	"github.com/a-novel/service-profile/internal/core"
	"github.com/a-novel/service-profile/internal/dao"
	"github.com/a-novel/service-profile/internal/handlers"
)

func main() {
	cfg := config.AppPresetDefault

	// Without an issuer, every caller reaches the routes, so only the standalone image starts so.
	if cfg.ServiceToken.Issuer == "" && !cfg.ServiceIssuerOptional {
		log.Fatalln("SERVICE_TOKEN_ISSUER is empty: set it, or SERVICE_ISSUER_OPTIONAL=true outside production")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otel.SetAppName(cfg.App.Name)

	lo.Must0(otel.Init(cfg.Otel))
	defer cfg.Otel.Flush()

	if env.OTLPEndpoint == "" {
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	}

	// A server started during a planned downtime refuses work, so it leaves the database alone.
	if !downtime.Started(cfg.App.DowntimeStart, time.Now()) {
		ctx = lo.Must(postgres.NewContext(ctx, cfg.Postgres))

		database := lo.Must(cfg.Postgres.DB(ctx))
		defer closeDatabase(database)
	}

	// =================================================================================================================
	// CLIENTS
	// =================================================================================================================

	jsonKeysClient := lo.Must(servicejsonkeys.NewClient(
		net.JoinHostPort(cfg.Dependencies.ServiceJsonKeysHost, strconv.Itoa(cfg.Dependencies.ServiceJsonKeysPort)),
		lo.Must(cfg.Dependencies.ServiceJsonKeysCredentials.Options(ctx))...,
	))
	defer jsonKeysClient.Close()

	claimsVerifier := lo.Must(servicejsonkeys.NewClaimsVerifier[serviceauthentication.Claims](jsonKeysClient))
	withAuth := serviceauthentication.NewAuthHandler(claimsVerifier, cfg.Permissions, cfg.Logger)

	// =================================================================================================================
	// DAO
	// =================================================================================================================

	daoProfileSelect := dao.NewPgProfileSelect()
	daoProfileSelectByUser := dao.NewPgProfileSelectByUser()
	daoProfileUsernameUpsert := dao.NewPgProfileUsernameUpsert()
	daoProfilePictureUpsert := dao.NewPgProfilePictureUpsert()
	daoProfilePictureDelete := dao.NewPgProfilePictureDelete()
	daoProfilePictureSelect := dao.NewPgProfilePictureSelect()

	// =================================================================================================================
	// SERVICES
	// =================================================================================================================

	serviceProfileGet := core.NewProfileGet(daoProfileSelect)
	serviceProfileGetSelf := core.NewProfileGetSelf(daoProfileSelectByUser)
	serviceProfileUsernameUpdate := core.NewProfileUsernameUpdate(daoProfileUsernameUpsert)
	serviceProfilePictureUpdate := core.NewProfilePictureUpdate(daoProfilePictureUpsert)
	serviceProfilePictureDelete := core.NewProfilePictureDelete(daoProfilePictureDelete)
	serviceProfilePictureGet := core.NewProfilePictureGet(daoProfilePictureSelect)

	// =================================================================================================================
	// HANDLERS
	// =================================================================================================================

	handlerPing := handlers.NewRestPing()
	handlerHealth := handlers.NewRestHealth(jsonKeysClient)
	handlerProfileGet := handlers.NewRestProfileGet(serviceProfileGet, cfg.Logger)
	handlerProfileGetSelf := handlers.NewRestProfileGetSelf(serviceProfileGetSelf, cfg.Logger)
	handlerProfileUsernameUpdate := handlers.NewRestProfileUsernameUpdate(serviceProfileUsernameUpdate, cfg.Logger)
	handlerProfilePictureUpdate := handlers.NewRestProfilePictureUpdate(serviceProfilePictureUpdate, cfg.Logger)
	handlerProfilePictureDelete := handlers.NewRestProfilePictureDelete(serviceProfilePictureDelete, cfg.Logger)
	handlerProfilePictureGet := handlers.NewRestProfilePictureGet(serviceProfilePictureGet, cfg.Logger)

	// =================================================================================================================
	// ROUTER
	// =================================================================================================================

	router := httpf.NewRouter(cfg.Rest, cfg.Otel, cfg.RestLogger, cfg.App.DowntimeStart)
	withServiceToken := handlers.RestServiceToken(
		cfg.ServiceToken, httpf.NewPoolClient(httpf.PoolOptions{}), cfg.Logger, cfg.Rest.Liveness,
	)

	router.Route("/v1", func(api chi.Router) {
		api.Get("/ping", handlerPing.ServeHTTP)
		api.Get("/healthcheck", handlerHealth.ServeHTTP)

		withAuth(api, config.PermissionProfileRead).Get("/profile", handlerProfileGet.ServeHTTP)
		withAuth(api, config.PermissionProfileRead).Get("/profile/self", handlerProfileGetSelf.ServeHTTP)
		withAuth(api, config.PermissionProfileRead).Get("/profile/picture", handlerProfilePictureGet.ServeHTTP)
		withAuth(api, config.PermissionProfileWrite).Put("/profile", handlerProfileUsernameUpdate.ServeHTTP)
		withAuth(api, config.PermissionProfileWrite).Put("/profile/picture", handlerProfilePictureUpdate.ServeHTTP)
		withAuth(api, config.PermissionProfileWrite).Delete("/profile/picture", handlerProfilePictureDelete.ServeHTTP)
	})

	// =================================================================================================================
	// RUN
	// =================================================================================================================

	log.Println("Starting REST server on :" + strconv.Itoa(cfg.Rest.Port))

	lo.Must0(httpf.ListenAndServe(ctx, cfg.Rest, withServiceToken(router)))
}

func closeDatabase(database io.Closer) {
	err := database.Close()
	if err != nil {
		log.Println("Close Postgres: " + err.Error())
	}
}
