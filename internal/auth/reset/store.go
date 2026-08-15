package reset

import "time"

// TokenRecord is a stored password-reset token (hash only).
type TokenRecord struct {
	TokenHash string    `yaml:"token_hash" json:"token_hash"`
	UserID    string    `yaml:"user_id" json:"user_id"`
	ExpiresAt time.Time `yaml:"expires_at" json:"expires_at"`
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
}

// TokenStore persists password-reset tokens.
type TokenStore interface {
	// ReplaceForUser deletes any existing tokens for userID, then stores record.
	ReplaceForUser(userID string, record TokenRecord) error
	// GetByHash returns a valid (non-expired) record, or ErrInvalidToken.
	// Expired records are deleted lazily.
	GetByHash(hash string) (*TokenRecord, error)
	// DeleteByHash removes a token by its hash. Missing keys are not an error.
	DeleteByHash(hash string) error
	// DeleteAllForUser removes every reset token for userID. Missing tokens are fine.
	DeleteAllForUser(userID string) error
}
