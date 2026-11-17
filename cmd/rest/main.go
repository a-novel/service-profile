// Command rest serves the public REST API to the callers its service tokens allow. It refuses to start
// without a token issuer unless SERVICE_ISSUER_OPTIONAL opts a local stack out. cmd/migrations applies
// the database schema.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	"github.com/a-novel/service-profile/internal/core"
	"github.com/a-novel/service-profile/internal/dao"
	"github.com/a-novel/service-profile/internal/handlers"
)

func main() {
	cfg := config.AppPresetDefault

	withServiceToken := lo.Must(newServiceTokenGate(cfg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otel.SetAppName(cfg.App.Name)

	lo.Must0(otel.Init(cfg.Otel))
	defer cfg.Otel.Flush()

	// A server started during a planned downtime refuses work, so it leaves the database alone.
	if !downtime.Started(cfg.App.DowntimeStart, time.Now()) {
		ctx = lo.Must(postgres.NewContext(ctx, cfg.Postgres))
		defer cfg.Postgres.Close()
	}

	// =================================================================================================================
	// CLIENTS
	// =================================================================================================================

	jsonKeysClient := lo.Must(servicejsonkeys.NewClient(
		cfg.Dependencies.ServiceJSONKeys.Address(), cfg.Dependencies.ServiceJSONKeys.DialOptions()...,
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

	router := httpf.NewMux(cfg.Rest, cfg.RestLogger, cfg.App.DowntimeStart)

	router.Route("/v1", func(api chi.Router) {
		api.Get("/ping", httpf.Ping)
		api.Get("/healthcheck", handlerHealth)

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

// newServiceTokenGate returns the middleware that wraps the whole router, so a request reaches a route
// only with a service token cfg accepts. The liveness route stays open to the kubelet. The gate refuses
// an empty issuer unless cfg opts a local stack out, so the server never starts open by mistake.
func newServiceTokenGate(cfg config.App) (func(http.Handler) http.Handler, error) {
	gate, err := httpf.ServiceTokenGate(
		cfg.ServiceToken, httpf.NewPoolClient(httpf.PoolOptions{}), cfg.Logger, cfg.Rest.Liveness,
	)
	if err != nil {
		return nil, fmt.Errorf("service token gate: %w", err)
	}

	return gate, nil
}
