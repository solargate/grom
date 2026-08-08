package reset

import "errors"

var (
	// ErrNotConfigured means password reset / mailer is not enabled.
	ErrNotConfigured = errors.New("password reset is not configured")
	// ErrInvalidToken means the token is missing, unknown, or expired.
	ErrInvalidToken = errors.New("invalid or expired reset token")
	// ErrRateLimited means too many requests.
	ErrRateLimited = errors.New("too many requests")
	// ErrWeakPassword means the new password does not meet policy.
	ErrWeakPassword = errors.New("password must be at least 8 characters")
)
