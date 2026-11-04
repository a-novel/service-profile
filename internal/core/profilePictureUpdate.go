package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	// Registers the JPEG decoder with the image package.
	_ "image/jpeg"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
	// Registers the WebP decoder with the image package.
	_ "golang.org/x/image/webp"

	"github.com/a-novel-kit/golib/otel"

	"github.com/a-novel/service-profile/internal/dao"
)

// ProfilePictureSideMax caps the side of a picture, in pixels. The platform draws its crop 512 px wide
// at most.
const ProfilePictureSideMax = 512

// pictureTypes maps each format the picture contract accepts, as the image package names it, to its
// media type. The bytes name the format; the request's Content-Type never does.
var pictureTypes = map[string]string{
	"jpeg": "image/jpeg",
	"webp": "image/webp",
}

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
// [PictureRefusedError].
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

	pictureType, err := decodePicture(ctx, request.Picture)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("decode picture: %w", err))
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

// decodePicture returns the media type of a picture the contract accepts, and a [PictureRefusedError]
// otherwise. It reads the dimensions from the header before it decodes the image, so a small file
// that declares a huge image costs no memory.
func decodePicture(ctx context.Context, picture []byte) (string, error) {
	_, span := otel.Tracer().Start(ctx, "core.decodePicture")
	defer span.End()

	config, format, err := image.DecodeConfig(bytes.NewReader(picture))
	pictureType, accepted := pictureTypes[format]

	switch {
	case err != nil || !accepted || (format == "webp" && webpAnimated(picture)):
		return "", refusePicture(span, ErrPictureFormat, err)
	case config.Width != config.Height:
		return "", refusePicture(span, ErrPictureRatio, nil)
	case config.Width < 1 || config.Width > ProfilePictureSideMax:
		return "", refusePicture(span, ErrPictureDimensions, nil)
	}

	_, _, err = image.Decode(bytes.NewReader(picture))
	if err != nil {
		return "", refusePicture(span, ErrPictureFormat, err)
	}

	return pictureType, nil
}

// refusePicture marks span failed by refusal, with the decoder's error when one caused it, and names
// the refusal as the span's error.type, so the traces count refusals by reason.
func refusePicture(span trace.Span, refusal PictureRefusedError, cause error) error {
	span.SetAttributes(semconv.ErrorTypeKey.String(string(refusal)))

	return otel.ReportError(span, errors.Join(refusal, cause))
}

// webpAnimated reports whether a WebP declares an animation: its VP8X chunk, which follows the RIFF
// header, sets the animation flag (WebP Container Specification, "Extended File Format").
func webpAnimated(picture []byte) bool {
	const animationFlag = 1 << 1

	return len(picture) > 20 && string(picture[12:16]) == "VP8X" && picture[20]&animationFlag != 0
}
