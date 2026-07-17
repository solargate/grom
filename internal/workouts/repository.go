package workouts

type Repository interface {
	Create(nickname string, workout *Workout) (*Workout, error)
	BeginCreate(nickname string, workout *Workout) (*Workout, string, func(), error)
	WriteMetadata(nickname string, workout *Workout) error
	Get(nickname, workoutID string) (*Workout, error)
	List(nickname string) ([]Workout, error)
	Delete(nickname, workoutID string) error
	RemoveEquipmentFromAll(nickname, equipmentID string) error
	HasStravaActivityID(nickname, stravaActivityID string) (bool, error)
	WorkoutDirName(nickname, workoutID string) (string, error)
}
