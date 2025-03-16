package errorz

import "errors"

var (
	ErrInvalidCallbackData = errors.New("invalid callback data")
	ErrInvalidCode         = errors.New("invalid code")
)
