package handlers

import (
	"context"
	"net/http"

	"github.com/a-novel-kit/golib/httpf"
	"github.com/a-novel-kit/golib/logging"
	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/core"
)

// RestProfileGetSelfService is the core operation [RestProfileGetSelf] delegates to.
type RestProfileGetSelfService interface {
	Exec(ctx context.Context, request *core.ProfileGetSelfRequest) (*core.Profile, error)
}

// profileGetSelfErrors maps a refused read of the caller's profile to its status. A caller who never
// wrote holds no profile.
var profileGetSelfErrors = httpf.ErrMap{
	core.ErrInvalidRequest:              http.StatusBadRequest,
	core.ErrProfileSelectByUserNotFound: http.StatusNotFound,
}

// RestProfileGetSelf is the HTTP handler for the get-own-profile endpoint. It responds with the
// signed-in caller's profile.
type RestProfileGetSelf struct {
	service RestProfileGetSelfService
	logger  logging.Log
}

func NewRestProfileGetSelf(service RestProfileGetSelfService, logger logging.Log) *RestProfileGetSelf {
	return &RestProfileGetSelf{service: service, logger: logger}
}

func (handler *RestProfileGetSelf) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer().Start(r.Context(), "rest.ProfileGetSelf")
	defer span.End()

	userID, err := requestOwner(r)
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, ownerErrors, err)

		return
	}

	profile, err := handler.service.Exec(ctx, &core.ProfileGetSelfRequest{UserID: userID})
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, profileGetSelfErrors, err)

		return
	}

	httpf.SendJSONStatus(w, span, http.StatusOK, loadProfile(profile))
}
