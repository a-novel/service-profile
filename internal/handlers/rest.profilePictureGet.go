package handlers

import (
	"bytes"
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/a-novel-kit/golib/httpf"
	"github.com/a-novel-kit/golib/logging"
	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/core"
)

// RestProfilePictureGetService is the core operation [RestProfilePictureGet] delegates to.
type RestProfilePictureGetService interface {
	Exec(ctx context.Context, request *core.ProfilePictureGetRequest) (*core.ProfilePicture, error)
}

// profilePictureGetErrors maps a refused picture read to its status.
var profilePictureGetErrors = httpf.ErrMap{
	core.ErrInvalidRequest:               http.StatusBadRequest,
	core.ErrProfilePictureSelectNotFound: http.StatusNotFound,
}

// RestProfilePictureGetRequest is the query parameters the get-picture endpoint accepts.
type RestProfilePictureGetRequest struct {
	ID uuid.UUID `schema:"id"`
}

// RestProfilePictureGet is the HTTP handler for the get-picture endpoint. It responds with the bytes of
// any profile's picture, to any caller holding an access token.
//
// The picture carries the media type it was stored with, which X-Content-Type-Options keeps the
// browser from second-guessing. Its strong ETag (RFC 9110 section 8.8.3) and Last-Modified come from
// its version, so [http.ServeContent] answers a matching conditional request with 304. Cache-Control is
// no-cache (RFC 9111 section 5.2.2.4): a cache revalidates the picture on every use. The platform caches
// for good under a URL that names the version.
type RestProfilePictureGet struct {
	service RestProfilePictureGetService
	logger  logging.Log
}

func NewRestProfilePictureGet(service RestProfilePictureGetService, logger logging.Log) *RestProfilePictureGet {
	return &RestProfilePictureGet{service: service, logger: logger}
}

func (handler *RestProfilePictureGet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer().Start(r.Context(), "rest.ProfilePictureGet")
	defer span.End()

	var request RestProfilePictureGetRequest

	err := muxDecoder.Decode(&request, r.URL.Query())
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, httpf.ErrMap{nil: http.StatusBadRequest}, err)

		return
	}

	picture, err := handler.service.Exec(ctx, &core.ProfilePictureGetRequest{ID: request.ID})
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, profilePictureGetErrors, err)

		return
	}

	header := w.Header()
	header.Set("Content-Type", picture.Type)
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Cache-Control", "no-cache")
	header.Set("ETag", `"`+strconv.FormatInt(picture.Version.UnixMicro(), 10)+`"`)

	http.ServeContent(w, r, "", picture.Version, bytes.NewReader(picture.Picture))
}
