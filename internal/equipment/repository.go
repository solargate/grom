package equipment

type Repository interface {
	List(nickname string) ([]Equipment, error)
	FindByID(nickname, id string) (*Equipment, error)
	FindByIDs(nickname string, ids []string) ([]Equipment, error)
	Create(nickname string, item *Equipment) (*Equipment, error)
	Update(nickname string, item *Equipment) (*Equipment, error)
	Delete(nickname, id string) error
}
