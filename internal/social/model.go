package social

import "time"

const (
	StatusPending  = "pending"
	StatusActive   = "active"
	StatusRejected = "rejected"
)

type Follow struct {
	ID               string    `yaml:"id" json:"id"`
	FollowerID       string    `yaml:"follower_id" json:"follower_id"`
	TargetActorURI   string    `yaml:"target_actor_uri" json:"target_actor_uri"`
	TargetHandle     string    `yaml:"target_handle" json:"target_handle"`
	TargetNickname   string    `yaml:"target_nickname" json:"target_nickname"`
	TargetName       string    `yaml:"target_name" json:"target_name"`
	TargetAvatarURL  string    `yaml:"target_avatar_url,omitempty" json:"target_avatar_url,omitempty"`
	TargetIsLocal    bool      `yaml:"target_is_local" json:"target_is_local"`
	Status           string    `yaml:"status" json:"status"`
	FollowActivityID string    `yaml:"follow_activity_id,omitempty" json:"follow_activity_id,omitempty"`
	CreatedAt        time.Time `yaml:"created_at" json:"created_at"`
}

type followsFile struct {
	Follows []Follow `yaml:"follows"`
}
