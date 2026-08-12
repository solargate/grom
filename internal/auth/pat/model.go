package pat

import "time"

const (
	TokenPrefix         = "grom_pat_"
	DefaultTTLDays      = 90
	MaxTokensPerUser    = 10
	MaxNameLen          = 64
	MinNameLen          = 1
	MaxExpiresDays      = 365
	TokenPrefixLen      = 12
	LastUsedTouchMinGap = time.Hour
)

const (
	ScopeWorkoutsRead    = "workouts:read"
	ScopeWorkoutsWrite   = "workouts:write"
	ScopeEquipmentRead   = "equipment:read"
	ScopeEquipmentWrite  = "equipment:write"
)

var ValidScopes = []string{
	ScopeWorkoutsRead,
	ScopeWorkoutsWrite,
	ScopeEquipmentRead,
	ScopeEquipmentWrite,
}

// TokenRecord is persisted metadata for a personal access token (secret is never stored).
type TokenRecord struct {
	ID          string     `yaml:"id" json:"id"`
	TokenHash   string     `yaml:"token_hash" json:"-"`
	TokenPrefix string     `yaml:"token_prefix" json:"token_prefix"`
	UserID      string     `yaml:"user_id" json:"user_id"`
	Name        string     `yaml:"name" json:"name"`
	Scopes      []string   `yaml:"scopes" json:"scopes"`
	CreatedAt   time.Time  `yaml:"created_at" json:"created_at"`
	ExpiresAt   *time.Time `yaml:"expires_at,omitempty" json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `yaml:"last_used_at,omitempty" json:"last_used_at,omitempty"`
}

func (r *TokenRecord) IsExpired(now time.Time) bool {
	if r == nil || r.ExpiresAt == nil {
		return false
	}
	return !r.ExpiresAt.After(now)
}

func HasScope(scopes []string, required string) bool {
	for _, s := range scopes {
		if s == required {
			return true
		}
	}
	return false
}
