package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/dao"
)

// ProfileUsernameUpdateDao writes the username of a user's profile, and creates the profile on the
// user's first write.
type ProfileUsernameUpdateDao interface {
	Exec(ctx context.Context, request *dao.ProfileUsernameUpsertRequest) (*dao.Profile, error)
}

// ProfileUsernameUpdateRequest carries the caller's new username.
type ProfileUsernameUpdateRequest struct {
	UserID uuid.UUID `validate:"required"`
	// Username is the name to show. The spaces at both ends are trimmed before [ValidateUsername] checks
	// it and the profile stores it. Empty once trimmed, it clears the username.
	Username string `validate:"username"`
}

// ProfileUsernameUpdate sets the caller's username.
type ProfileUsernameUpdate struct {
	dao ProfileUsernameUpdateDao
}

func NewProfileUsernameUpdate(dao ProfileUsernameUpdateDao) *ProfileUsernameUpdate {
	return &ProfileUsernameUpdate{dao: dao}
}

func (service *ProfileUsernameUpdate) Exec(
	ctx context.Context, request *ProfileUsernameUpdateRequest,
) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "core.ProfileUsernameUpdate")
	defer span.End()

	span.SetAttributes(attribute.String("profile.user_id", request.UserID.String()))

	err := validate.Struct(request)
	if err != nil {
		return nil, otel.ReportError(span, errors.Join(err, ErrInvalidRequest))
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("generate profile ID: %w", err))
	}

	var username *string
	if trimmed := strings.Trim(request.Username, " "); trimmed != "" {
		username = &trimmed
	}

	entity, err := service.dao.Exec(ctx, &dao.ProfileUsernameUpsertRequest{
		ID:       id,
		UserID:   request.UserID,
		Username: username,
		Now:      time.Now(),
	})
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("upsert username: %w", err))
	}

	return newProfile(entity), nil
}
