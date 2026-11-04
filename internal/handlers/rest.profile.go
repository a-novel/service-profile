package handlers

import (
	"time"

	"github.com/google/uuid"

	"github.com/a-novel/service-profile/internal/core"
)

// Profile is the REST representation of a profile. It never names the user who owns it.
type Profile struct {
	ID uuid.UUID `json:"id"`

	// Username is the name the profile shows, absent until its owner sets one.
	Username *string `json:"username,omitempty"`
	// PictureVersion is when the picture last changed, absent while the profile has no picture.
	PictureVersion *time.Time `json:"pictureVersion,omitempty"`
}

func loadProfile(profile *core.Profile) Profile {
	return Profile{
		ID:             profile.ID,
		Username:       profile.Username,
		PictureVersion: profile.PictureVersion,
	}
}
