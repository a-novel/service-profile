package core

import (
	"time"

	"github.com/google/uuid"

	"github.com/a-novel/service-profile/internal/dao"
)

// Profile is a user's public profile. It never names the user who owns it.
type Profile struct {
	ID uuid.UUID

	// Username is the name the profile shows, nil until its owner sets one.
	Username *string
	// PictureVersion is when the picture last changed, nil while the profile has no picture. Picture
	// URLs carry it, so a cache keeps each version for good.
	PictureVersion *time.Time
}

// ProfilePicture is the picture of a profile.
type ProfilePicture struct {
	// Type is the media type the picture is encoded in, image/webp or image/jpeg.
	Type string
	// Picture holds the encoded image.
	Picture []byte

	// Version is when the picture last changed.
	Version time.Time
}

func newProfile(entity *dao.Profile) *Profile {
	return &Profile{
		ID:             entity.ID,
		Username:       entity.Username,
		PictureVersion: entity.PictureUpdatedAt,
	}
}
