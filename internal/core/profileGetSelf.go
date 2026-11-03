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

// ProfileGetSelfDao reads the profile of a user.
type ProfileGetSelfDao interface {
	Exec(ctx context.Context, request *dao.ProfileSelectByUserRequest) (*dao.Profile, error)
}

// ProfileGetSelfRequest names the caller.
type ProfileGetSelfRequest struct {
	UserID uuid.UUID `validate:"required"`
}

// ProfileGetSelf reads the caller's own profile, which exists from the caller's first write.
type ProfileGetSelf struct {
	dao ProfileGetSelfDao
}

func NewProfileGetSelf(dao ProfileGetSelfDao) *ProfileGetSelf {
	return &ProfileGetSelf{dao: dao}
}

func (service *ProfileGetSelf) Exec(ctx context.Context, request *ProfileGetSelfRequest) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "core.ProfileGetSelf")
	defer span.End()

	span.SetAttributes(attribute.String("profile.user_id", request.UserID.String()))

	err := validate.Struct(request)
	if err != nil {
		return nil, otel.ReportError(span, errors.Join(err, ErrInvalidRequest))
	}

	entity, err := service.dao.Exec(ctx, &dao.ProfileSelectByUserRequest{UserID: request.UserID})
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("select profile: %w", err))
	}

	return newProfile(entity), nil
}
