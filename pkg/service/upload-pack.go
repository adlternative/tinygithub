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

	gitCmd := cmd.NewGitCommand(serviceName).WithGitDir(repo.Path()).
		WithOptions("--stateless-rpc").
		WithArgs(repo.Path()).WithStderr(&stderrBuf).WithStdout(c.Writer).WithStdin(r)

	if protocol := c.GetHeader("Git-Protocol"); protocol != "" {
		version := strings.TrimPrefix(protocol, "version=")
		if version == "2" || version == "1" {
			gitCmd.WithEnv(fmt.Sprintf("GIT_PROTOCOL=version=%s", version))
		}
	}

	if err = gitCmd.Run(c); err != nil {
		return fmt.Errorf("git command failed with: err:%w, stderr:%v", err, stderrBuf.String())
	}

	return nil
}
