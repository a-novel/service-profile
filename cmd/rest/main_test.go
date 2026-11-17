package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/a-novel-kit/golib/servicetoken"
	"github.com/a-novel-kit/golib/servicetoken/servicetokentest"

	"github.com/a-novel/service-profile/internal/config"
)

func TestNewServiceTokenGate(t *testing.T) {
	t.Parallel()

	const (
		audience = "service-profile"
		caller   = "system:serviceaccount:agora:caller"
	)

	issuer := servicetokentest.NewIssuer(t)
	gated := servicetoken.Config{Issuer: issuer.URL, Audience: audience, Callers: []string{caller}}

	testCases := []struct {
		name string

		serviceToken servicetoken.Config
		path         string
		token        string

		expectStatus int
		expectErr    error
	}{
		{
			name: "Success",

			serviceToken: gated,
			path:         "/v1/profile",
			token:        issuer.Sign(t, issuer.Claims(caller, audience)),

			expectStatus: http.StatusNoContent,
		},
		{
			// The kubelet probes the liveness route without a token.
			name: "Success/Liveness",

			serviceToken: gated,
			path:         "/v1/ping",

			expectStatus: http.StatusNoContent,
		},
		{
			// A local stack opts out, and the server takes every request.
			name: "Success/IssuerOptional",

			serviceToken: servicetoken.Config{IssuerOptional: true},
			path:         "/v1/profile",

			expectStatus: http.StatusNoContent,
		},
		{
			name: "Error/Missing",

			serviceToken: gated,
			path:         "/v1/profile",

			expectStatus: http.StatusUnauthorized,
		},
		{
			// A deployment that forgets the issuer stops at startup rather than admit every caller.
			name: "Error/IssuerRequired",

			expectErr: servicetoken.ErrIssuerRequired,
		},
	}

	// The router answers 204, so a 401 proves the gate refused the request before it.
	router := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.AppPresetDefault
			cfg.ServiceToken = testCase.serviceToken

			withServiceToken, err := newServiceTokenGate(cfg)
			require.ErrorIs(t, err, testCase.expectErr)

			if testCase.expectErr != nil {
				return
			}

			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, testCase.path, nil)
			if testCase.token != "" {
				request.Header.Set(servicetoken.Header, testCase.token)
			}

			w := httptest.NewRecorder()
			withServiceToken(router).ServeHTTP(w, request)

			require.Equal(t, testCase.expectStatus, w.Code)
		})
	}
}
