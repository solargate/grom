package workouts

type Repository interface {
	Create(nickname string, workout *Workout) (*Workout, error)
	BeginCreate(nickname string, workout *Workout) (*Workout, string, func(), error)
	WriteMetadata(nickname string, workout *Workout) error
	Update(nickname string, workout *Workout) (*Workout, error)
	Get(nickname, workoutID string) (*Workout, error)
	List(nickname string) ([]Workout, error)
	// ListPage returns up to limit workouts older than cursor (nil = first page),
	// ordered by start_date DESC, id DESC. hasMore is true when more items exist.
	ListPage(nickname string, cursor *Cursor, limit int) (items []Workout, hasMore bool, err error)
	Delete(nickname, workoutID string) error
	RemoveEquipmentFromAll(nickname, equipmentID string) error
	HasStravaActivityID(nickname, stravaActivityID string) (bool, error)
	WorkoutDirName(nickname, workoutID string) (string, error)
}
