package core

import (
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
)
