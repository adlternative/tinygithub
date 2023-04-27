package service

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func InfoRefs(storage *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		userName := c.Param("username")
		// check user exist
		repoName := c.Param("reponame")

		// check repo exist
		serviceName := c.Query("service")

		c.Writer.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", serviceName))
		c.Writer.Header().Set("Cache-Control", "no-cache")

		repo, err := storage.GetRepository(userName, repoName)
		if err != nil {
			log.WithError(err).Errorf("GetRepository failed")
			c.String(http.StatusNotFound, "GetRepository failed: %v", err)
			return
		}

		switch serviceName {
		case "git-upload-pack":
			fallthrough
		case "git-receive-pack":
			serviceName = strings.TrimPrefix(serviceName, "git-")
		default:
			log.WithError(err).Errorf("unknown git service %s", serviceName)
			c.String(http.StatusBadRequest, "unknown git service %s", serviceName)
			return
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
			log.WithError(err).Errorf("git command failed with: err:%v, stderr:%v", err, stderrBuf.String())
			c.String(http.StatusInternalServerError, "git command failed with: err:%v, stderr:%v", err, stderrBuf.String())
			return
		}
	}
}
