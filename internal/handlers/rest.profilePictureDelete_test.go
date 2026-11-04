package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/a-novel/service-profile/internal/config"
	"github.com/a-novel/service-profile/internal/core"
	"github.com/a-novel/service-profile/internal/handlers"
	handlersmocks "github.com/a-novel/service-profile/internal/handlers/mocks"
)

func TestRestProfilePictureDelete(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	testCases := []struct {
		name string

		caller *uuid.UUID

		serviceMock *serviceMock[*core.Profile]

		expectStatus   int
		expectResponse any
	}{
		{
			name: "Success",

			caller: &userID,

			serviceMock: &serviceMock[*core.Profile]{resp: &core.Profile{ID: profileID}},

			expectStatus:   http.StatusOK,
			expectResponse: map[string]any{"id": profileID.String()},
		},
		{
			name: "Error/Anonymous",

			expectStatus: http.StatusForbidden,
		},
		{
			name: "Error/NotFound",

			caller: &userID,

			serviceMock: &serviceMock[*core.Profile]{err: core.ErrProfilePictureDeleteNotFound},

			expectStatus: http.StatusNotFound,
		},
		{
			name: "Error/Internal",

			caller: &userID,

			serviceMock: &serviceMock[*core.Profile]{err: errFoo},

			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service := handlersmocks.NewMockRestProfilePictureDeleteService(t)

			if testCase.serviceMock != nil {
				service.EXPECT().
					Exec(mock.Anything, &core.ProfilePictureDeleteRequest{UserID: userID}).
					Return(testCase.serviceMock.resp, testCase.serviceMock.err)
			}

			request := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/v1/profile/picture", nil)

			w := httptest.NewRecorder()
			handlers.NewRestProfilePictureDelete(service, config.LoggerDev).ServeHTTP(w, asUser(request, testCase.caller))

			requireResponse(t, w.Result(), testCase.expectStatus, testCase.expectResponse)
			service.AssertExpectations(t)
		})
	}
}
