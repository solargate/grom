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
	bucketIdxWorkoutsExternal = []byte("idx_workouts_external")
	bucketFedFollowers      = []byte("fed_followers")
	bucketFedInbox          = []byte("fed_inbox")
	bucketFedAuthors        = []byte("fed_authors")
	bucketSpeedCharts       = []byte("speed_charts")
	bucketFedSpeedCharts    = []byte("fed_speed_charts")
	bucketHeartRateCharts   = []byte("heart_rate_charts")
	bucketFedHeartRateCharts = []byte("fed_heart_rate_charts")
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
	bucketIdxWorkoutsExternal,
	bucketFedFollowers,
	bucketFedInbox,
	bucketFedAuthors,
	bucketSpeedCharts,
	bucketFedSpeedCharts,
	bucketHeartRateCharts,
	bucketFedHeartRateCharts,
}

const schemaVersion = "1"
