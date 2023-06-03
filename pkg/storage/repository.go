package storage

type Repo struct {
	path string

	userName string
	repoName string
}

func (r *Repo) UserName() string {
	return r.userName
}

func (r *Repo) RepoName() string {
	return r.repoName
}

func (r *Repo) Path() string {
	return r.path
}

func NewRepository(path string, userName, repoName string) *Repo {
	return &Repo{
		path:     path,
		userName: userName,
		repoName: repoName,
	}
}
