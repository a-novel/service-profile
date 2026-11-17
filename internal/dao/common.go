package dao

import "errors"

var (
	// ErrProfileNotFound is returned when no profile has the requested ID, or the user holds none yet.
	ErrProfileNotFound = errors.New("profile not found")
	// ErrProfilePictureNotFound is returned when no profile has the requested ID, or the profile holds
	// no picture.
	ErrProfilePictureNotFound = errors.New("profile picture not found")
)
