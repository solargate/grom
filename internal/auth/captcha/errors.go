package captcha

import "errors"

var (
	// ErrDisabled means captcha is not enabled on this instance.
	ErrDisabled = errors.New("captcha is disabled")
	// ErrMissing means the request omitted the altcha payload.
	ErrMissing = errors.New("captcha payload is required")
	// ErrInvalid means the payload failed signature or solution checks.
	ErrInvalid = errors.New("invalid captcha")
	// ErrExpired means the challenge TTL elapsed.
	ErrExpired = errors.New("captcha expired")
	// ErrReplay means the same valid payload was submitted again.
	ErrReplay = errors.New("captcha already used")
	// ErrRateLimited means too many challenge requests from one IP.
	ErrRateLimited = errors.New("too many captcha challenges")
)
