package users

type Repository interface {
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	FindByNickname(nickname string) (*User, error)
	Search(query, excludeUserID string, limit int) ([]User, error)
	ListAll() ([]User, error)
	Create(nickname, name, email, password string) (*User, error)
	UpdateProfile(userID, name string) (*User, error)
	UpdatePassword(userID, passwordHash string) error

	GetProfile(userID string) (*Profile, error)
	PutProfile(userID string, profile Profile) error
	SetLastSportType(userID, sportType string) error
	SetLastEquipmentForSport(userID, sportType string, equipmentIDs []string) error
	RemoveEquipmentFromLastSets(userID, equipmentID string) error
}
