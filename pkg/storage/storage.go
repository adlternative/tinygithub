package storage

import (
	"errors"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"

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

func (s *Storage) GetRepository(userName, repoName string) (*Repo, error) {
	if !strings.HasSuffix(repoName, ".git") {
		repoName = repoName + ".git"
	}
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

func (s *Storage) CreateRepository(ctx *gin.Context, userName, repoName string) (*Repo, error) {
	if !strings.HasSuffix(repoName, ".git") {
		repoName = repoName + ".git"
	}
	userDir := path.Clean(path.Join(s.path, userName))
	repoPath := path.Clean(path.Join(userDir, repoName))

	var pathErr *fs.PathError
	_, err := os.Stat(repoPath)
	if err == nil {
		return nil, fmt.Errorf("repo %s exists", repoPath)
	} else if err != os.ErrNotExist && !errors.As(err, &pathErr) {
		return nil, fmt.Errorf("repo %s stat failed: %w", repoPath, err)
	}

	err = os.MkdirAll(userDir, 0750)
	if err != nil {
		return nil, fmt.Errorf("mkdir user dir %s failed", userDir)
	}
	var stderrBuf strings.Builder

	gitCmd := cmd.NewGitCommand("init").WithGitDir(repoPath).
		WithOptions("--bare").WithStderr(&stderrBuf)
	err = gitCmd.Run(ctx)
	if err != nil {
		log.WithError(err).Errorf("git command failed with: Error:%v, stderr:%v", err, stderrBuf.String())
		return nil, err
	}

	return NewRepository(repoPath), nil
}

func (s *Storage) RemoveRepository(ctx *gin.Context, userName, repoName string) error {
	if !strings.HasSuffix(repoName, ".git") {
		repoName = repoName + ".git"
	}
	userDir := path.Clean(path.Join(s.path, userName))
	repoPath := path.Clean(path.Join(userDir, repoName))

	var pathErr *fs.PathError
	_, err := os.Stat(repoPath)
	if err != nil {
		if err == os.ErrNotExist || errors.As(err, &pathErr) {
			return nil
		}
		return err
	}

	log.Infof("storage removeRepository %s", repoPath)
	err = os.RemoveAll(repoPath)
	return err
}

func BackUp(repoPath, backUpPath string) error {
	_, err := os.Stat(repoPath)
	if err != nil {
		return err
	}

	// check backup not exist
	_, err = os.Stat(backUpPath)
	if err == nil {
		return fmt.Errorf("backUpPath %s existed", backUpPath)
	}

	// Copy the repository to the backup directory
	backUpCmd := exec.Command("cp", "-R", repoPath, backUpPath)
	if err = backUpCmd.Run(); err != nil {
		return err
	}

	return nil
}

func Restore(repoPath, backUpPath string) error {
	// check repo exists
	_, err := os.Stat(repoPath)
	if err == nil {
		return nil
	}

	// check backup exists
	_, err = os.Stat(backUpPath)
	if err != nil {
		return fmt.Errorf("stat backUpPath error: %w", err)
	}

	// restore the repository from the backup
	restoreCmd := exec.Command("cp", "-R", backUpPath, repoPath)
	if err = restoreCmd.Run(); err != nil {
		return err
	}
	return nil
}
