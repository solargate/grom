package notifications

import "errors"

var (
	ErrNotFound            = errors.New("push subscription not found")
	ErrInvalidSubscription = errors.New("invalid push subscription")
)
