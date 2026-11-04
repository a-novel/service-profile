package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/a-novel-kit/golib/httpf"
	"github.com/a-novel-kit/golib/logging"
	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/core"
)

// RestProfilePictureUpdateService is the core operation [RestProfilePictureUpdate] delegates to.
type RestProfilePictureUpdateService interface {
	Exec(ctx context.Context, request *core.ProfilePictureUpdateRequest) (*core.Profile, error)
}

// profilePictureUpdateErrors maps a refused picture to its status. A picture the contract refuses
// answers 422, with its reason as the problem's reason tag. A body over the router's limit answers 413.
var profilePictureUpdateErrors = httpf.ErrMap{
	core.ErrInvalidRequest: http.StatusBadRequest,
	core.ErrPictureRefused: http.StatusUnprocessableEntity,
}

// RestProfilePictureUpdate is the HTTP handler for the update-picture endpoint. Its body is the
// encoded image, whose bytes name its format whatever the Content-Type says. It responds with the
// caller's profile, which the caller's first write creates.
type RestProfilePictureUpdate struct {
	service RestProfilePictureUpdateService
	logger  logging.Log
}

func NewRestProfilePictureUpdate(
	service RestProfilePictureUpdateService, logger logging.Log,
) *RestProfilePictureUpdate {
	return &RestProfilePictureUpdate{service: service, logger: logger}
}

func (handler *RestProfilePictureUpdate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer().Start(r.Context(), "rest.ProfilePictureUpdate")
	defer span.End()

	userID, err := requestOwner(r)
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, ownerErrors, err)

		return
	}

	// The picture is decoded whole, so the body is read whole, within the router's limit.
	picture, err := io.ReadAll(r.Body)
	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, httpf.ErrMap{nil: http.StatusBadRequest}, err)

		return
	}

	profile, err := handler.service.Exec(ctx, &core.ProfilePictureUpdateRequest{UserID: userID, Picture: picture})

	var refusal core.PictureRefusedError
	if errors.As(err, &refusal) {
		err = httpf.WithTag(err, "reason", string(refusal))
	}

	if err != nil {
		httpf.HandleError(ctx, handler.logger, w, span, profilePictureUpdateErrors, err)

		return
	}

	httpf.SendJSONStatus(w, span, http.StatusOK, loadProfile(profile))
}
