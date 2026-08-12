package pat

import "time"

// Repository persists personal access token metadata.
type Repository interface {
	Create(record TokenRecord) error
	ListByUser(userID string) ([]TokenRecord, error)
	CountByUser(userID string) (int, error)
	GetByHash(hash string) (*TokenRecord, error)
	DeleteByUserAndID(userID, id string) error
	UpdateLastUsed(id string, at time.Time) error
}
