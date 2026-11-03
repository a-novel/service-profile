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

// ProfileGetDao reads a profile by its ID.
type ProfileGetDao interface {
	Exec(ctx context.Context, request *dao.ProfileSelectRequest) (*dao.Profile, error)
}

// ProfileGetRequest names the profile to read.
type ProfileGetRequest struct {
	ID uuid.UUID `validate:"required"`
}

// ProfileGet reads any profile by its ID.
type ProfileGet struct {
	dao ProfileGetDao
}

func NewProfileGet(dao ProfileGetDao) *ProfileGet {
	return &ProfileGet{dao: dao}
}

func (service *ProfileGet) Exec(ctx context.Context, request *ProfileGetRequest) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "core.ProfileGet")
	defer span.End()

	span.SetAttributes(attribute.String("profile.id", request.ID.String()))

	err := validate.Struct(request)
	if err != nil {
		return nil, otel.ReportError(span, errors.Join(err, ErrInvalidRequest))
	}

	entity, err := service.dao.Exec(ctx, &dao.ProfileSelectRequest{ID: request.ID})
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("select profile: %w", err))
	}

	return newProfile(entity), nil
}
