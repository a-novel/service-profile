package core

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// ErrInvalidRequest is returned by a service Exec when its request fails struct validation.
var ErrInvalidRequest = errors.New("invalid request")

func init() {
	err := validate.RegisterValidation("username", ValidateUsername)
	if err != nil {
		panic(err)
	}
}
