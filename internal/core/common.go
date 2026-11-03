package core

import (
	"errors"

	"github.com/a-novel/service-profile/internal/dao"
)

var (
	// ErrProfileSelectNotFound is returned when no profile has the requested ID.
	ErrProfileSelectNotFound = dao.ErrProfileSelectNotFound
	// ErrProfileSelectByUserNotFound is returned when the caller holds no profile yet.
	ErrProfileSelectByUserNotFound = dao.ErrProfileSelectByUserNotFound
	// ErrProfilePictureDeleteNotFound is returned when the caller holds no profile yet.
	ErrProfilePictureDeleteNotFound = dao.ErrProfilePictureDeleteNotFound
	// ErrProfilePictureSelectNotFound is returned when no profile has the requested ID, or the profile
	// holds no picture.
	ErrProfilePictureSelectNotFound = dao.ErrProfilePictureSelectNotFound

	// ErrPictureRefused is returned by [ProfilePictureUpdate], as a [PictureRefusedError], for a picture
	// the contract refuses.
	ErrPictureRefused = errors.New("picture refused")
)

// PictureRefusedError names the rule a refused picture broke, and matches [ErrPictureRefused]. Its
// value is the reason a client reads, and the error.type of the span that refuses the picture.
type PictureRefusedError string

const (
	// ErrPictureFormat refuses bytes that are neither a still WebP nor a JPEG, or do not decode
	// whole.
	ErrPictureFormat PictureRefusedError = "format"
	// ErrPictureRatio refuses a picture that is not square.
	ErrPictureRatio PictureRefusedError = "ratio"
	// ErrPictureDimensions refuses a picture whose side is not 1 to [ProfilePictureSideMax] pixels.
	ErrPictureDimensions PictureRefusedError = "dimensions"
)

func (refusal PictureRefusedError) Error() string {
	return ErrPictureRefused.Error() + ": " + string(refusal)
}

// Is reports the refusal as [ErrPictureRefused].
func (refusal PictureRefusedError) Is(target error) bool {
	return target == ErrPictureRefused
}
