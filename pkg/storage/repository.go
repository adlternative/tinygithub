package storage

type Repository struct {
	path string
}

func (r *Repository) Path() string {
	return r.path
}

func NewRepository(path string) *Repository {
	return &Repository{
		path: path,
	}
}
