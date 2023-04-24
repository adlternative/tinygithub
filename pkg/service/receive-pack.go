package service

import (
	"compress/gzip"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"io"
	"strings"

	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-gonic/gin"
)

func ReceivePack(c *gin.Context, storage *storage.Storage, userName, repoName string) error {
	serviceName := "receive-pack"

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
	// git -c <repoPath> receive-pack --stateless-rpc <repoPath>

	gitCmd := cmd.NewGitCommand(serviceName).WithGitDir(repo.Path()).
		WithOptions("--stateless-rpc").
		WithArgs(repo.Path()).WithStderr(&stderrBuf).WithStdout(c.Writer).WithStdin(r)

	if err = gitCmd.Run(c); err != nil {
		return fmt.Errorf("git command failed with: err:%w, stderr:%v", err, stderrBuf.String())
	}

	return nil
}
