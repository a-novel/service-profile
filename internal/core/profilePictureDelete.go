package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/dao"
)

// ProfilePictureDeleteDao removes the picture of a user's profile.
type ProfilePictureDeleteDao interface {
	Exec(ctx context.Context, request *dao.ProfilePictureDeleteRequest) (*dao.Profile, error)
}

// ProfilePictureDeleteRequest names the caller.
type ProfilePictureDeleteRequest struct {
	UserID uuid.UUID `validate:"required"`
}

// ProfilePictureDelete removes the caller's picture. A profile without a picture keeps none.
type ProfilePictureDelete struct {
	dao ProfilePictureDeleteDao
}

func NewProfilePictureDelete(dao ProfilePictureDeleteDao) *ProfilePictureDelete {
	return &ProfilePictureDelete{dao: dao}
}

func (service *ProfilePictureDelete) Exec(ctx context.Context, request *ProfilePictureDeleteRequest) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "core.ProfilePictureDelete")
	defer span.End()

	span.SetAttributes(attribute.String("profile.user_id", request.UserID.String()))

	err := validate.Struct(request)
	if err != nil {
		return nil, otel.ReportError(span, errors.Join(err, ErrInvalidRequest))
	}

	entity, err := service.dao.Exec(ctx, &dao.ProfilePictureDeleteRequest{UserID: request.UserID, Now: time.Now()})
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("delete picture: %w", err))
	}

	return newProfile(entity), nil
}
