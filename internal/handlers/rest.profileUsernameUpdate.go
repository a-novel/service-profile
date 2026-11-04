package handlers

import (
	"context"
	"net/http"

	"github.com/a-novel-kit/golib/httpf"
	"github.com/a-novel-kit/golib/logging"
	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/core"
)

// RestProfileUsernameUpdateService is the core operation [RestProfileUsernameUpdate] delegates to.
type RestProfileUsernameUpdateService interface {
	Exec(ctx context.Context, request *core.ProfileUsernameUpdateRequest) (*core.Profile, error)
}

// profileUsernameUpdateErrors maps a refused username to its status.
var profileUsernameUpdateErrors = httpf.ErrMap{
	core.ErrInvalidRequest: http.StatusBadRequest,
}

// RestProfileUsernameUpdateRequest is the JSON body the update-username endpoint accepts.
type RestProfileUsernameUpdateRequest struct {
	Username string `json:"username"`
}

// RestProfileUsernameUpdate is the HTTP handler for the update-username endpoint. It responds with the
// caller's profile, which the caller's first write creates.
type RestProfileUsernameUpdate struct {
	service RestProfileUsernameUpdateService
	logger  logging.Log
}

func NewRestProfileUsernameUpdate(
	service RestProfileUsernameUpdateService, logger logging.Log,
) *RestProfileUsernameUpdate {
	return &RestProfileUsernameUpdate{service: service, logger: logger}
}

func (handler *RestProfileUsernameUpdate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer().Start(r.Context(), "rest.ProfileUsernameUpdate")
	defer span.End()

	userID, err := requestOwner(r)
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, ownerErrors, err)

		return
	}

	var request RestProfileUsernameUpdateRequest

	err = httpf.DecodeJSON(r.Body, &request)
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, httpf.ErrMap{nil: http.StatusBadRequest}, err)

		return
	}

	profile, err := handler.service.Exec(ctx, &core.ProfileUsernameUpdateRequest{
		UserID:   userID,
		Username: request.Username,
	})
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, profileUsernameUpdateErrors, err)

		return
	}

	httpf.SendJSONStatus(w, span, http.StatusOK, loadProfile(profile))
}
