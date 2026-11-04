package handlers_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/a-novel/service-profile/internal/config"
	"github.com/a-novel/service-profile/internal/core"
	"github.com/a-novel/service-profile/internal/handlers"
	handlersmocks "github.com/a-novel/service-profile/internal/handlers/mocks"
)

func TestRestProfilePictureGet(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	target := "/v1/profile/picture?id=" + profileID.String()
	picture := &core.ProfilePicture{Type: "image/webp", Picture: []byte("picture"), Version: pictureVersion}
	etag := `"1609556645600000"`

	testCases := []struct {
		name string

		target      string
		ifNoneMatch string

		serviceMock *serviceMock[*core.ProfilePicture]

		expectStatus int
		expectBody   string
	}{
		{
			name: "Success",

			target: target,

			serviceMock: &serviceMock[*core.ProfilePicture]{resp: picture},

			expectStatus: http.StatusOK,
			expectBody:   "picture",
		},
		{
			name: "Success/NotModified",

			target:      target,
			ifNoneMatch: etag,

			serviceMock: &serviceMock[*core.ProfilePicture]{resp: picture},

			expectStatus: http.StatusNotModified,
		},
		{
			name: "Error/MalformedID",

			target: "/v1/profile/picture?id=not-a-uuid",

			expectStatus: http.StatusBadRequest,
		},
		{
			name: "Error/NotFound",

			target: target,

			serviceMock: &serviceMock[*core.ProfilePicture]{err: core.ErrProfilePictureSelectNotFound},

			expectStatus: http.StatusNotFound,
		},
		{
			name: "Error/Internal",

			target: target,

			serviceMock: &serviceMock[*core.ProfilePicture]{err: errFoo},

			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service := handlersmocks.NewMockRestProfilePictureGetService(t)

			if testCase.serviceMock != nil {
				service.EXPECT().
					Exec(mock.Anything, &core.ProfilePictureGetRequest{ID: profileID}).
					Return(testCase.serviceMock.resp, testCase.serviceMock.err)
			}

			// An anonymous session reads any picture.
			request := asUser(httptest.NewRequestWithContext(t.Context(), http.MethodGet, testCase.target, nil), nil)
			if testCase.ifNoneMatch != "" {
				request.Header.Set("If-None-Match", testCase.ifNoneMatch)
			}

			w := httptest.NewRecorder()
			handlers.NewRestProfilePictureGet(service, config.LoggerDev).ServeHTTP(w, request)

			response := w.Result()
			require.Equal(t, testCase.expectStatus, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			require.NoError(t, errors.Join(err, response.Body.Close()))

			if testCase.expectStatus >= http.StatusBadRequest {
				return
			}

			require.Equal(t, testCase.expectBody, string(body))
			require.Equal(t, etag, response.Header.Get("ETag"))
			require.Equal(t, "no-cache", response.Header.Get("Cache-Control"))

			// A 304 leaves out the headers that describe the body.
			if testCase.expectStatus == http.StatusOK {
				require.Equal(t, "image/webp", response.Header.Get("Content-Type"))
				require.Equal(t, "nosniff", response.Header.Get("X-Content-Type-Options"))
				require.Equal(t, "Sat, 02 Jan 2021 03:04:05 GMT", response.Header.Get("Last-Modified"))
			}
		})
	}
}
