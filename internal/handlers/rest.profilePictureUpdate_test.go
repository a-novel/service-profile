package handlers_test

import (
	"bytes"
	"errors"
	"io"
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

func TestRestProfilePictureUpdate(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	picture := []byte("picture")

	testCases := []struct {
		name string

		caller *uuid.UUID
		// bodyLimit stands for the router's body limit, when set.
		bodyLimit int64

		serviceMock *serviceMock[*core.Profile]

		expectStatus   int
		expectResponse any
	}{
		{
			name: "Success",

			caller: &userID,

			serviceMock: &serviceMock[*core.Profile]{resp: coreProfile},

			expectStatus:   http.StatusOK,
			expectResponse: jsonProfile,
		},
		{
			name: "Error/Anonymous",

			expectStatus: http.StatusForbidden,
		},
		{
			name: "Error/TooLarge",

			caller:    &userID,
			bodyLimit: int64(len(picture)) - 1,

			expectStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name: "Error/Refused",

			caller: &userID,

			serviceMock: &serviceMock[*core.Profile]{err: core.ErrPictureRatio},

			expectStatus: http.StatusUnprocessableEntity,
			expectResponse: map[string]any{
				"type":   "about:blank",
				"title":  "Unprocessable Entity",
				"status": float64(http.StatusUnprocessableEntity),
				"tags":   map[string]any{"reason": "ratio"},
			},
		},
		{
			name: "Error/InvalidRequest",

			caller: &userID,

			serviceMock: &serviceMock[*core.Profile]{err: core.ErrInvalidRequest},

			expectStatus: http.StatusBadRequest,
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

			service := handlersmocks.NewMockRestProfilePictureUpdateService(t)

			if testCase.serviceMock != nil {
				service.EXPECT().
					Exec(mock.Anything, &core.ProfilePictureUpdateRequest{UserID: userID, Picture: picture}).
					Return(testCase.serviceMock.resp, testCase.serviceMock.err)
			}

			w := httptest.NewRecorder()

			var body io.Reader = bytes.NewReader(picture)
			if testCase.bodyLimit > 0 {
				body = http.MaxBytesReader(w, io.NopCloser(body), testCase.bodyLimit)
			}

			request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/v1/profile/picture", body)
			handlers.NewRestProfilePictureUpdate(service, config.LoggerDev).ServeHTTP(w, asUser(request, testCase.caller))

			requireResponse(t, w.Result(), testCase.expectStatus, testCase.expectResponse)
			service.AssertExpectations(t)
		})
	}
}
