package storage

type Repo struct {
	path string
}

func (r *Repo) Path() string {
	return r.path
}

func NewRepository(path string) *Repo {
	return &Repo{
		path: path,
	}
}
