package social

type Repository interface {
	FindByID(id string) (*Follow, error)
	ListByFollower(followerID string) ([]Follow, error)
	ListActiveFollowing(followerID string) ([]Follow, error)
	ListActiveByTarget(targetHandle string) ([]Follow, error)
	FindExisting(followerID, targetHandle string) (*Follow, error)
	Create(follow Follow) (*Follow, error)
	UpdateStatus(id, status string) (*Follow, error)
	FindByFollowActivityID(activityID string) (*Follow, error)
	UpdateActivityID(id, activityID string) (*Follow, error)
	Delete(id string) error
	// DeleteInvolving removes follows where the user is follower or target (any status).
	DeleteInvolving(followerID, localHandle string) error
}
