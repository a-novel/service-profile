package handlers

import (
	"context"
	"net/http"

	"github.com/a-novel-kit/golib/httpf"
	"github.com/a-novel-kit/golib/logging"
	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/core"
)

// RestProfilePictureDeleteService is the core operation [RestProfilePictureDelete] delegates to.
type RestProfilePictureDeleteService interface {
	Exec(ctx context.Context, request *core.ProfilePictureDeleteRequest) (*core.Profile, error)
}

// profilePictureDeleteErrors maps a refused picture removal to its status. A caller who never wrote
// holds no profile.
var profilePictureDeleteErrors = httpf.ErrMap{
	core.ErrInvalidRequest:               http.StatusBadRequest,
	core.ErrProfilePictureDeleteNotFound: http.StatusNotFound,
}

// RestProfilePictureDelete is the HTTP handler for the delete-picture endpoint. It responds with the
// caller's profile, without a picture.
type RestProfilePictureDelete struct {
	service RestProfilePictureDeleteService
	logger  logging.Log
}

func NewRestProfilePictureDelete(
	service RestProfilePictureDeleteService, logger logging.Log,
) *RestProfilePictureDelete {
	return &RestProfilePictureDelete{service: service, logger: logger}
}

func (handler *RestProfilePictureDelete) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer().Start(r.Context(), "rest.ProfilePictureDelete")
	defer span.End()

	userID, err := requestOwner(r)
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, ownerErrors, err)

		return
	}

	profile, err := handler.service.Exec(ctx, &core.ProfilePictureDeleteRequest{UserID: userID})
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, profilePictureDeleteErrors, err)

		return
	}

	httpf.SendJSONStatus(w, span, http.StatusOK, loadProfile(profile))
}
