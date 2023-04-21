package storage

import (
	"fmt"
	"os"
	"path"

	"github.com/adlternative/tinygithub/pkg/config"
	"github.com/spf13/viper"
)

type Storage struct {
	path string
}

func NewStorage() (*Storage, error) {
	s := &Storage{
		path: viper.GetString(config.Storage),
	}
	if err := s.valid(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) Path() string {
	return s.path
}

func (s *Storage) GetRepository(userName, repoName string) (*Repository, error) {
	repoPath := path.Clean(path.Join(s.path, userName, repoName))
	info, err := os.Stat(repoPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("repoPath %s is not a dir", repoPath)
	}
	return NewRepository(repoPath), nil
}

func (s *Storage) valid() error {
	fi, err := os.Stat(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("storage path is not exist: %w", err)
		} else {
			return fmt.Errorf("bad storage path: %w", err)
		}
	}
	if !fi.IsDir() {
		return fmt.Errorf("storage path not dir: %s", s.path)
	}
	return nil
}
