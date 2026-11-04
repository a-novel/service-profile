package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/a-novel/service-profile/internal/config"
	"github.com/a-novel/service-profile/internal/core"
	"github.com/a-novel/service-profile/internal/handlers"
	handlersmocks "github.com/a-novel/service-profile/internal/handlers/mocks"
)

func TestRestProfileUsernameUpdate(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	body := `{"username": "Jean Paul"}`

	testCases := []struct {
		name string

		caller *uuid.UUID
		body   string

		serviceMock *serviceMock[*core.Profile]

		expectStatus   int
		expectResponse any
	}{
		{
			name: "Success",

			caller: &userID,
			body:   body,

			serviceMock: &serviceMock[*core.Profile]{resp: coreProfile},

			expectStatus:   http.StatusOK,
			expectResponse: jsonProfile,
		},
		{
			name: "Error/Anonymous",

			body: body,

			expectStatus: http.StatusForbidden,
		},
		{
			name: "Error/MalformedBody",

			caller: &userID,
			body:   "not json",

			expectStatus: http.StatusBadRequest,
		},
		{
			name: "Error/InvalidUsername",

			caller: &userID,
			body:   body,

			serviceMock: &serviceMock[*core.Profile]{err: core.ErrInvalidRequest},

			expectStatus: http.StatusBadRequest,
		},
		{
			name: "Error/Internal",

			caller: &userID,
			body:   body,

			serviceMock: &serviceMock[*core.Profile]{err: errFoo},

			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service := handlersmocks.NewMockRestProfileUsernameUpdateService(t)

			if testCase.serviceMock != nil {
				service.EXPECT().
					Exec(mock.Anything, &core.ProfileUsernameUpdateRequest{UserID: userID, Username: "Jean Paul"}).
					Return(testCase.serviceMock.resp, testCase.serviceMock.err)
			}

			request := httptest.NewRequestWithContext(
				t.Context(), http.MethodPut, "/v1/profile", strings.NewReader(testCase.body),
			)

			w := httptest.NewRecorder()
			handlers.NewRestProfileUsernameUpdate(service, config.LoggerDev).ServeHTTP(w, asUser(request, testCase.caller))

			requireResponse(t, w.Result(), testCase.expectStatus, testCase.expectResponse)
			service.AssertExpectations(t)
		})
	}
}
