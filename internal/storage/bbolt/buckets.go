package bbolt

var (
	bucketMeta                = []byte("meta")
	bucketUsers               = []byte("users")
	bucketUserProfiles        = []byte("user_profiles")
	bucketIdxUsersEmail       = []byte("idx_users_email")
	bucketIdxUsersNick        = []byte("idx_users_nick")
	bucketEquipment           = []byte("equipment")
	bucketFollows             = []byte("follows")
	bucketIdxFollowsFollower  = []byte("idx_follows_follower")
	bucketIdxFollowsTarget    = []byte("idx_follows_target")
	bucketIdxFollowsActivity  = []byte("idx_follows_activity")
	bucketWorkouts            = []byte("workouts")
	bucketIdxWorkoutsID       = []byte("idx_workouts_id")
	bucketIdxWorkoutsExternal = []byte("idx_workouts_external")
	bucketFedFollowers        = []byte("fed_followers")
	bucketFedInbox            = []byte("fed_inbox")
	bucketFedAuthors          = []byte("fed_authors")
	bucketWorkoutLikes        = []byte("workout_likes")
	bucketFedWorkoutLikes     = []byte("fed_workout_likes")
	bucketLikeActivities      = []byte("like_activities")
	bucketWorkoutComments     = []byte("workout_comments")
	bucketFedWorkoutComments  = []byte("fed_workout_comments")
	bucketCommentActivities   = []byte("comment_activities")
	bucketSpeedCharts         = []byte("speed_charts")
	bucketFedSpeedCharts      = []byte("fed_speed_charts")
	bucketHeartRateCharts     = []byte("heart_rate_charts")
	bucketFedHeartRateCharts  = []byte("fed_heart_rate_charts")
	bucketResetTokens         = []byte("reset_tokens")
	bucketPersonalAccessTokens = []byte("personal_access_tokens")
)

var allBuckets = [][]byte{
	bucketMeta,
	bucketUsers,
	bucketUserProfiles,
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
	bucketWorkoutLikes,
	bucketFedWorkoutLikes,
	bucketLikeActivities,
	bucketWorkoutComments,
	bucketFedWorkoutComments,
	bucketCommentActivities,
	bucketSpeedCharts,
	bucketFedSpeedCharts,
	bucketHeartRateCharts,
	bucketFedHeartRateCharts,
	bucketResetTokens,
	bucketPersonalAccessTokens,
}

const schemaVersion = "1"
