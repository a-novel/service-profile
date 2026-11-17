package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	servicejsonkeys "github.com/a-novel/service-json-keys/v2/pkg/go"

	"github.com/a-novel/service-profile/internal/handlers"
	handlersmocks "github.com/a-novel/service-profile/internal/handlers/mocks"
)

func TestRestHealth(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		// postgres hands the probe the test database; without it, the probe meets no database.
		postgres    bool
		jsonKeysErr error

		expectStatus   int
		expectPostgres string
		expectJSONKeys string
	}{
		{
			name:           "Success",
			postgres:       true,
			expectStatus:   http.StatusOK,
			expectPostgres: "up",
			expectJSONKeys: "up",
		},
		{
			name:           "Error/PostgresDown",
			expectStatus:   http.StatusServiceUnavailable,
			expectPostgres: "down",
			expectJSONKeys: "up",
		},
		{
			name:           "Error/JSONKeysDown",
			postgres:       true,
			jsonKeysErr:    errors.New("unavailable"),
			expectStatus:   http.StatusServiceUnavailable,
			expectPostgres: "up",
			expectJSONKeys: "down",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			if testCase.postgres {
				ctx = postgresContext(t)
			}

			jsonKeys := handlersmocks.NewMockRestHealthAPIJSONKeys(t)
			jsonKeys.EXPECT().
				Status(mock.Anything, &servicejsonkeys.StatusRequest{}).
				Return(&servicejsonkeys.StatusResponse{}, testCase.jsonKeysErr)

			request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/v1/healthcheck", nil)

			w := httptest.NewRecorder()
			handlers.NewRestHealth(jsonKeys).ServeHTTP(w, request)

			response := w.Result()
			require.Contains(t, response.Header.Get("Content-Type"), "application/json")
			// Exact matching keeps raw dependency errors out of the public report.
			requireResponse(t, response, testCase.expectStatus, map[string]any{
				"client:postgres": map[string]any{"status": testCase.expectPostgres},
				"api:jsonKeys":    map[string]any{"status": testCase.expectJSONKeys},
			})
		})
	}
}
