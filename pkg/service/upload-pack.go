package service

import (
	"compress/gzip"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"strings"

	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-gonic/gin"
)

func UploadPack(c *gin.Context, storage *storage.Storage, userName, repoName string) error {
	serviceName := "upload-pack"

	repo, err := storage.GetRepository(userName, repoName)
	if err != nil {
		return err
	}

	var r io.Reader
	r = c.Request.Body
	encoding := c.GetHeader("Content-Encoding")
	switch encoding {
	case "gzip":
		r, err = gzip.NewReader(r)
		if err != nil {
			return err
		}
	default:
	}

	c.Writer.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", serviceName))
	c.Writer.Header().Set("Cache-Control", "no-cache")

	var stderrBuf strings.Builder
	// git -c <repoPath> upload-pack --stateless-rpc <repoPath>
	command := exec.CommandContext(c, "git", fmt.Sprintf("--git-dir=%s", repo.Path()), serviceName, "--stateless-rpc", repo.Path())
	command.Stdin = r
	command.Stdout = c.Writer
	command.Stderr = &stderrBuf

	if protocol := c.GetHeader("Git-Protocol"); protocol != "" {
		version := strings.TrimPrefix(protocol, "version=")
		if version == "2" || version == "1" {
			command.Env = append(command.Env, fmt.Sprintf("GIT_PROTOCOL=version=%s", version))
		}
	}

	log.Debug("git command: ", command.String())

	if err = command.Run(); err != nil {
		return fmt.Errorf("git command failed with: err:%w, stderr:%v", err, stderrBuf.String())
	}

	return nil
}
