package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/a-novel/service-profile/internal/config"
	"github.com/a-novel/service-profile/internal/core"
	"github.com/a-novel/service-profile/internal/handlers"
	handlersmocks "github.com/a-novel/service-profile/internal/handlers/mocks"
)

func TestRestProfileGet(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	target := "/v1/profile?id=" + profileID.String()

	testCases := []struct {
		name string

		target string

		serviceMock *serviceMock[*core.Profile]

		expectStatus   int
		expectResponse any
	}{
		{
			name: "Success",

			target: target,

			serviceMock: &serviceMock[*core.Profile]{resp: coreProfile},

			expectStatus:   http.StatusOK,
			expectResponse: jsonProfile,
		},
		{
			// A profile without a username or a picture leaves both out.
			name: "Success/Bare",

			target: target,

			serviceMock: &serviceMock[*core.Profile]{resp: &core.Profile{ID: profileID}},

			expectStatus:   http.StatusOK,
			expectResponse: map[string]any{"id": profileID.String()},
		},
		{
			name: "Error/MalformedID",

			target: "/v1/profile?id=not-a-uuid",

			expectStatus: http.StatusBadRequest,
		},
		{
			name: "Error/InvalidRequest",

			target: target,

			serviceMock: &serviceMock[*core.Profile]{err: core.ErrInvalidRequest},

			expectStatus: http.StatusBadRequest,
		},
		{
			name: "Error/NotFound",

			target: target,

			serviceMock: &serviceMock[*core.Profile]{err: core.ErrProfileSelectNotFound},

			expectStatus: http.StatusNotFound,
		},
		{
			name: "Error/Internal",

			target: target,

			serviceMock: &serviceMock[*core.Profile]{err: errFoo},

			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service := handlersmocks.NewMockRestProfileGetService(t)

			if testCase.serviceMock != nil {
				service.EXPECT().
					Exec(mock.Anything, &core.ProfileGetRequest{ID: profileID}).
					Return(testCase.serviceMock.resp, testCase.serviceMock.err)
			}

			// An anonymous session reads any profile.
			request := asUser(httptest.NewRequestWithContext(t.Context(), http.MethodGet, testCase.target, nil), nil)

			w := httptest.NewRecorder()
			handlers.NewRestProfileGet(service, config.LoggerDev).ServeHTTP(w, request)

			requireResponse(t, w.Result(), testCase.expectStatus, testCase.expectResponse)
			service.AssertExpectations(t)
		})
	}
}
