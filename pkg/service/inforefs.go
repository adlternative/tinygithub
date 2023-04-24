package service

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-gonic/gin"
	"strings"
)

func InfoRefs(c *gin.Context, storage *storage.Storage, userName, repoName string) error {
	// check repo exist
	serviceName := c.Query("service")

	c.Writer.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", serviceName))
	c.Writer.Header().Set("Cache-Control", "no-cache")

	repo, err := storage.GetRepository(userName, repoName)
	if err != nil {
		return err
	}

	switch serviceName {
	case "git-upload-pack":
		fallthrough
	case "git-receive-pack":
		serviceName = strings.TrimPrefix(serviceName, "git-")
	default:
		return fmt.Errorf("unkown git service %s", serviceName)
	}

	var stderrBuf strings.Builder
	// git -c <repoPath> upload-pack --advertise-refs --stateless-rpc <repoPath>
	// git -c <repoPath> receive-pack --advertise-refs --stateless-rpc <repoPath>

	gitCmd := cmd.NewGitCommand(serviceName).WithGitDir(repo.Path()).
		WithOptions("--advertise-refs", "--stateless-rpc", "--show-service").
		WithArgs(repo.Path()).WithStderr(&stderrBuf).WithStdout(c.Writer)

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
