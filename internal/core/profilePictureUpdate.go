package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/a-novel-kit/golib/otel"
	"github.com/a-novel-kit/golib/picture"

	"github.com/a-novel/service-profile/internal/dao"
)

// ProfilePictureSideMax caps the side of a picture, in pixels. The platform draws its crop 512 px wide
// at most.
const ProfilePictureSideMax = 512

// profilePictureRule is the rule a profile picture follows: a square, 1 to [ProfilePictureSideMax]
// pixels wide.
var profilePictureRule = picture.Rule{RatioWidth: 1, RatioHeight: 1, MaxWidth: ProfilePictureSideMax}

// ProfilePictureUpdateDao writes the picture of a user's profile, and creates the profile on the
// user's first write.
type ProfilePictureUpdateDao interface {
	Exec(ctx context.Context, request *dao.ProfilePictureUpsertRequest) (*dao.Profile, error)
}

// ProfilePictureUpdateRequest carries the caller's new picture.
type ProfilePictureUpdateRequest struct {
	UserID uuid.UUID `validate:"required"`
	// Picture holds the encoded image, 256 KiB at most, as the schema caps it.
	Picture []byte `validate:"required,max=262144"`
}

// ProfilePictureUpdate replaces the caller's picture. It accepts a still WebP or a JPEG that decodes
// whole into a square of 1 to [ProfilePictureSideMax] pixels, and refuses any other picture with a
// [picture.RefusedError].
type ProfilePictureUpdate struct {
	dao ProfilePictureUpdateDao
}

func NewProfilePictureUpdate(dao ProfilePictureUpdateDao) *ProfilePictureUpdate {
	return &ProfilePictureUpdate{dao: dao}
}

func (service *ProfilePictureUpdate) Exec(ctx context.Context, request *ProfilePictureUpdateRequest) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "core.ProfilePictureUpdate")
	defer span.End()

	span.SetAttributes(attribute.String("profile.user_id", request.UserID.String()))

	err := validate.Struct(request)
	if err != nil {
		return nil, otel.ReportError(span, errors.Join(err, ErrInvalidRequest))
	}

	pictureType, err := picture.Check(ctx, request.Picture, profilePictureRule)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("check picture: %w", err))
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("generate profile ID: %w", err))
	}

	entity, err := service.dao.Exec(ctx, &dao.ProfilePictureUpsertRequest{
		ID:      id,
		UserID:  request.UserID,
		Picture: request.Picture,
		Type:    pictureType,
		Now:     time.Now(),
	})
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("upsert picture: %w", err))
	}

	return newProfile(entity), nil
}
