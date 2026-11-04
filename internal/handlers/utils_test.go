package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	serviceauthentication "github.com/a-novel/service-authentication/v2/pkg/go"

	"github.com/a-novel-kit/golib/postgres"
	postgrespresets "github.com/a-novel-kit/golib/postgres/presets"

	"github.com/a-novel/service-profile/internal/config/configtest"
	"github.com/a-novel/service-profile/internal/core"
)

var (
	userID    = uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	profileID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

	pictureVersion = time.Date(2021, 1, 2, 3, 4, 5, 600000000, time.UTC)

	// coreProfile is the profile the service mocks return.
	coreProfile = &core.Profile{ID: profileID, Username: lo.ToPtr("Jean Paul"), PictureVersion: &pictureVersion}
	// jsonProfile is coreProfile as the handlers send it.
	jsonProfile = map[string]any{
		"id":             profileID.String(),
		"username":       "Jean Paul",
		"pictureVersion": "2021-01-02T03:04:05.6Z",
	}
)

// serviceMock is what the service mock of a handler test case returns. A nil one means the handler
// must not call the service.
type serviceMock[Resp any] struct {
	resp Resp
	err  error
}

// asUser returns the request as the auth middleware hands it over: carrying the claims of userID, or
// of an anonymous session when userID is nil.
func asUser(request *http.Request, userID *uuid.UUID) *http.Request {
	claims := &serviceauthentication.Claims{UserID: userID}

	return request.WithContext(serviceauthentication.SetClaimsContext(request.Context(), claims))
}

// requireResponse checks the status of a handler's response, and its JSON body when one is expected.
func requireResponse(t *testing.T, response *http.Response, expectStatus int, expectBody any) {
	t.Helper()

	require.Equal(t, expectStatus, response.StatusCode)

	if expectBody == nil {
		return
	}

	data, err := io.ReadAll(response.Body)
	require.NoError(t, errors.Join(err, response.Body.Close()))

	var body any
	require.NoError(t, json.Unmarshal(data, &body))
	require.Equal(t, expectBody, body)
}

const (
	// serviceTokenAudience is the audience the service token tests answer to.
	serviceTokenAudience = "service-profile"
	// serviceTokenCaller is the one caller the service token tests allow.
	serviceTokenCaller = "system:serviceaccount:agora:caller"
	// serviceTokenStranger is a caller outside the allowed ones.
	serviceTokenStranger = "system:serviceaccount:agora:stranger"
)

// probeSetup selects the state of the PostgreSQL connection a health probe meets.
type probeSetup int

const (
	// probeHealthy gives the probe an open connection.
	probeHealthy probeSetup = iota
	// probeWithoutPostgres leaves PostgreSQL out of the context.
	probeWithoutPostgres
	// probeClosedPostgres gives the probe a closed connection pool.
	probeClosedPostgres
	// probeInTransaction runs the probe inside an open transaction, which it must leave open.
	probeInTransaction
	// probeCancelled gives the probe a cancelled context.
	probeCancelled
)

// newProbeContext returns the context a health probe runs with under the given setup. Each call
// opens a pool of its own, so closing it leaves the other tests' connections alone.
func newProbeContext(t *testing.T, setup probeSetup) context.Context {
	t.Helper()

	ctx := t.Context()
	if setup == probeWithoutPostgres {
		return ctx
	}

	ctx, err := postgres.NewContext(ctx, postgrespresets.NewDefault(configtest.PostgresPreset.Options()...))
	require.NoError(t, err)

	pg, err := postgres.GetContext(ctx)
	require.NoError(t, err)

	db, ok := pg.(*bun.DB)
	require.True(t, ok)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	switch setup {
	case probeClosedPostgres:
		require.NoError(t, db.Close())
	case probeInTransaction:
		// The transaction outlives the test context, so the rollback proves the probe left it open.
		tx, err := db.BeginTx(context.WithoutCancel(ctx), nil)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, tx.Rollback()) })

		ctx = context.WithValue(ctx, postgres.ContextKey{}, tx)
	case probeCancelled:
		cancelled, cancel := context.WithCancel(ctx)
		cancel()

		ctx = cancelled
	}

	return ctx
}
