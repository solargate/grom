package notifications

import (
	"strings"
	"time"
)

const (
	TypeWorkoutLiked     = "workout.liked"
	TypeWorkoutCommented = "workout.commented"
)

// PushSubscription is a Web Push endpoint registered by a client installation.
type PushSubscription struct {
	UserID         string    `yaml:"user_id" json:"user_id"`
	InstallationID string    `yaml:"installation_id" json:"installation_id"`
	Endpoint       string    `yaml:"endpoint" json:"endpoint"`
	P256dh         string    `yaml:"p256dh" json:"p256dh"`
	Auth           string    `yaml:"auth" json:"auth"`
	Platform       string    `yaml:"platform,omitempty" json:"platform,omitempty"`
	DeviceName     string    `yaml:"device_name,omitempty" json:"device_name,omitempty"`
	UpdatedAt      time.Time `yaml:"updated_at" json:"updated_at"`
}

// Event is a transport-agnostic notification payload delivered to devices.
type Event struct {
	Type             string `json:"type"`
	ActorDisplayName string `json:"actor_display_name"`
	WorkoutID        string `json:"workout_id"`
	WorkoutTitle     string `json:"workout_title"`
	Owner            string `json:"owner"`
	Slot             string `json:"slot"`
	EventID          string `json:"event_id"`
}

// Repository persists push subscriptions.
type Repository interface {
	Upsert(sub PushSubscription) error
	ListByUser(userID string) ([]PushSubscription, error)
	DeleteByUserAndInstallation(userID, installationID string) error
	DeleteAllForUser(userID string) error
	DeleteByEndpoint(endpoint string) error
	ListAll() ([]PushSubscription, error)
}

// ActorDisplayName picks a human-readable label: name → nickname → handle.
func ActorDisplayName(name, nickname, handle string) string {
	if s := strings.TrimSpace(name); s != "" {
		return s
	}
	if s := strings.TrimSpace(nickname); s != "" {
		return s
	}
	return strings.TrimSpace(handle)
}

// SlotLiked returns the collapse slot key for likes on a workout.
func SlotLiked(owner, workoutID string) string {
	return "liked:" + owner + ":" + workoutID
}

// SlotCommented returns the collapse slot key for comments on a workout.
func SlotCommented(owner, workoutID string) string {
	return "commented:" + owner + ":" + workoutID
}
