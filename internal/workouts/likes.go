package workouts

import "sort"

type WorkoutLikeUser struct {
	Handle    string `yaml:"handle" json:"handle"`
	Nickname  string `yaml:"nickname" json:"nickname"`
	Name      string `yaml:"name" json:"name"`
	IsLocal   bool   `yaml:"is_local" json:"is_local"`
	AvatarURL string `yaml:"avatar_url,omitempty" json:"avatar_url,omitempty"`
}

type WorkoutLikes struct {
	Likes int               `yaml:"likes" json:"likes"`
	Users []WorkoutLikeUser `yaml:"users" json:"users"`
}

type LikesRepository interface {
	GetLocal(ownerNickname, workoutID string) (*WorkoutLikes, error)
	PutLocal(ownerNickname, workoutID string, likes *WorkoutLikes) error
	DeleteLocal(ownerNickname, workoutID string) error

	GetFederated(viewerNickname, ownerHandle, workoutID string) (*WorkoutLikes, error)
	PutFederated(viewerNickname, ownerHandle, workoutID string, likes *WorkoutLikes) error
	DeleteFederated(viewerNickname, ownerHandle, workoutID string) error

	GetLikeActivityID(actorNickname, objectID string) (string, error)
	PutLikeActivityID(actorNickname, objectID, activityID string) error
	DeleteLikeActivityID(actorNickname, objectID string) error
}

func NormalizeWorkoutLikes(likes *WorkoutLikes) WorkoutLikes {
	if likes == nil {
		return WorkoutLikes{Users: []WorkoutLikeUser{}}
	}
	seen := make(map[string]WorkoutLikeUser, len(likes.Users))
	for _, user := range likes.Users {
		if user.Handle == "" {
			continue
		}
		seen[user.Handle] = user
	}
	users := make([]WorkoutLikeUser, 0, len(seen))
	for _, user := range seen {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].Handle < users[j].Handle
	})
	return WorkoutLikes{
		Likes: len(users),
		Users: users,
	}
}

func LikesContainUser(likes *WorkoutLikes, handle string) bool {
	if likes == nil || handle == "" {
		return false
	}
	for _, user := range likes.Users {
		if user.Handle == handle {
			return true
		}
	}
	return false
}

func AddWorkoutLikeUser(likes *WorkoutLikes, user WorkoutLikeUser) WorkoutLikes {
	norm := NormalizeWorkoutLikes(likes)
	if user.Handle == "" {
		return norm
	}
	if LikesContainUser(&norm, user.Handle) {
		return norm
	}
	norm.Users = append(norm.Users, user)
	return NormalizeWorkoutLikes(&norm)
}

func RemoveWorkoutLikeUser(likes *WorkoutLikes, handle string) WorkoutLikes {
	norm := NormalizeWorkoutLikes(likes)
	if handle == "" {
		return norm
	}
	filtered := make([]WorkoutLikeUser, 0, len(norm.Users))
	for _, user := range norm.Users {
		if user.Handle == handle {
			continue
		}
		filtered = append(filtered, user)
	}
	norm.Users = filtered
	return NormalizeWorkoutLikes(&norm)
}
