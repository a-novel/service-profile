package core

import (
	"github.com/a-novel/service-profile/internal/dao"
)

var (
	// ErrProfileNotFound is returned when no profile has the requested ID, or the caller holds none yet.
	ErrProfileNotFound = dao.ErrProfileNotFound
	// ErrProfilePictureNotFound is returned when no profile has the requested ID, or the profile holds
	// no picture.
	ErrProfilePictureNotFound = dao.ErrProfilePictureNotFound
)
