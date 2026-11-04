package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/a-novel-kit/golib/httpf"
	"github.com/a-novel-kit/golib/logging"
	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/core"
)

// RestProfileGetService is the core operation [RestProfileGet] delegates to.
type RestProfileGetService interface {
	Exec(ctx context.Context, request *core.ProfileGetRequest) (*core.Profile, error)
}

// profileGetErrors maps a refused profile read to its status.
var profileGetErrors = httpf.ErrMap{
	core.ErrInvalidRequest:        http.StatusBadRequest,
	core.ErrProfileSelectNotFound: http.StatusNotFound,
}

// RestProfileGetRequest is the query parameters the get-profile endpoint accepts.
type RestProfileGetRequest struct {
	ID uuid.UUID `schema:"id"`
}

// RestProfileGet is the HTTP handler for the get-profile endpoint. It responds with any profile, to
// any caller holding an access token.
type RestProfileGet struct {
	service RestProfileGetService
	logger  logging.Log
}

func NewRestProfileGet(service RestProfileGetService, logger logging.Log) *RestProfileGet {
	return &RestProfileGet{service: service, logger: logger}
}

func (handler *RestProfileGet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer().Start(r.Context(), "rest.ProfileGet")
	defer span.End()

	var request RestProfileGetRequest

	err := muxDecoder.Decode(&request, r.URL.Query())
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, httpf.ErrMap{nil: http.StatusBadRequest}, err)

		return
	}

	profile, err := handler.service.Exec(ctx, &core.ProfileGetRequest{ID: request.ID})
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, profileGetErrors, err)

		return
	}

	httpf.SendJSONStatus(w, span, http.StatusOK, loadProfile(profile))
}
