package dao

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Profile is a user's public profile, as every read returns it: without the user who owns it, and
// without its picture's bytes.
type Profile struct {
	bun.BaseModel `bun:"table:profiles,alias:profiles"`

	ID uuid.UUID `bun:"id,pk,type:uuid"`

	// Username is the name the profile shows, nil until its owner sets one.
	Username *string `bun:"username"`
	// PictureUpdatedAt is when the picture last changed, nil while the profile has no picture.
	PictureUpdatedAt *time.Time `bun:"picture_updated_at"`
}

// ProfilePicture is the picture of a profile.
type ProfilePicture struct {
	bun.BaseModel `bun:"table:profiles,alias:profiles"`

	// Picture holds the encoded image.
	Picture []byte `bun:"picture"`
	// Type is the media type the picture is encoded in, image/webp or image/jpeg.
	Type string `bun:"picture_type"`

	// UpdatedAt is when the picture last changed.
	UpdatedAt time.Time `bun:"picture_updated_at"`
}
