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
	// sportTypes nil means no sport filter; when non-nil, only matching SportType values are returned.
	ListPage(nickname string, cursor *Cursor, limit int, sportTypes map[string]struct{}) (items []Workout, hasMore bool, err error)
	Delete(nickname, workoutID string) error
	RemoveEquipmentFromAll(nickname, equipmentID string) error
	HasExternalID(nickname, name, id string) (bool, error)
	WorkoutDirName(nickname, workoutID string) (string, error)
}
