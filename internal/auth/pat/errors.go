package pat

import "errors"

var (
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrNotFound        = errors.New("token not found")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrTooManyTokens   = errors.New("too many tokens")
	ErrInvalidScope    = errors.New("invalid scope")
	ErrInsufficientScope = errors.New("insufficient scope")
)
