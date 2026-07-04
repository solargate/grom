package users

import "time"

type User struct {
	ID           string    `yaml:"id" json:"id"`
	Nickname     string    `yaml:"nickname" json:"nickname"`
	Name         string    `yaml:"name,omitempty" json:"name,omitempty"`
	Email        string    `yaml:"email" json:"email"`
	PasswordHash string    `yaml:"password_hash" json:"-"`
	CreatedAt    time.Time `yaml:"created_at" json:"created_at"`
}

type userFile struct {
	Users []User `yaml:"users"`
}
