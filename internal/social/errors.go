package social

import "errors"

var (
	ErrFollowNotFound   = errors.New("follow not found")
	ErrAlreadyFollowing = errors.New("already following this user")
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	ErrFollowNotActive  = errors.New("follow is not active")
)
