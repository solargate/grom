package users

import "errors"

var (
	ErrEmailTaken      = errors.New("email already registered")
	ErrNicknameTaken   = errors.New("nickname already taken")
	ErrInvalidNickname = errors.New("invalid nickname")
	ErrUserNotFound    = errors.New("user not found")
)
