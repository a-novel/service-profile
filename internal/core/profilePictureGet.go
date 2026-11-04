package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/dao"
)

// ProfilePictureGetDao reads the picture of a profile.
type ProfilePictureGetDao interface {
	Exec(ctx context.Context, request *dao.ProfilePictureSelectRequest) (*dao.ProfilePicture, error)
}

// ProfilePictureGetRequest names the profile whose picture to read.
type ProfilePictureGetRequest struct {
	ID uuid.UUID `validate:"required"`
}

// ProfilePictureGet reads the picture of any profile by the profile's ID.
type ProfilePictureGet struct {
	dao ProfilePictureGetDao
}

func NewProfilePictureGet(dao ProfilePictureGetDao) *ProfilePictureGet {
	return &ProfilePictureGet{dao: dao}
}

func (service *ProfilePictureGet) Exec(
	ctx context.Context, request *ProfilePictureGetRequest,
) (*ProfilePicture, error) {
	ctx, span := otel.Tracer().Start(ctx, "core.ProfilePictureGet")
	defer span.End()

	span.SetAttributes(attribute.String("profile.id", request.ID.String()))

	err := validate.Struct(request)
	if err != nil {
		return nil, otel.ReportError(span, errors.Join(err, ErrInvalidRequest))
	}

	entity, err := service.dao.Exec(ctx, &dao.ProfilePictureSelectRequest{ID: request.ID})
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("select picture: %w", err))
	}

	return &ProfilePicture{Type: entity.Type, Picture: entity.Picture, Version: entity.UpdatedAt}, nil
}
