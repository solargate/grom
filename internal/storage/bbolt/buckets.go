package bbolt

var (
	bucketMeta              = []byte("meta")
	bucketUsers             = []byte("users")
	bucketIdxUsersEmail     = []byte("idx_users_email")
	bucketIdxUsersNick      = []byte("idx_users_nick")
	bucketEquipment         = []byte("equipment")
	bucketFollows           = []byte("follows")
	bucketIdxFollowsFollower = []byte("idx_follows_follower")
	bucketIdxFollowsTarget  = []byte("idx_follows_target")
	bucketIdxFollowsActivity = []byte("idx_follows_activity")
	bucketWorkouts          = []byte("workouts")
	bucketIdxWorkoutsID     = []byte("idx_workouts_id")
	bucketIdxWorkoutsStrava = []byte("idx_workouts_strava")
	bucketFedFollowers      = []byte("fed_followers")
	bucketFedInbox          = []byte("fed_inbox")
	bucketFedAuthors        = []byte("fed_authors")
)

var allBuckets = [][]byte{
	bucketMeta,
	bucketUsers,
	bucketIdxUsersEmail,
	bucketIdxUsersNick,
	bucketEquipment,
	bucketFollows,
	bucketIdxFollowsFollower,
	bucketIdxFollowsTarget,
	bucketIdxFollowsActivity,
	bucketWorkouts,
	bucketIdxWorkoutsID,
	bucketIdxWorkoutsStrava,
	bucketFedFollowers,
	bucketFedInbox,
	bucketFedAuthors,
}

const schemaVersion = "1"
