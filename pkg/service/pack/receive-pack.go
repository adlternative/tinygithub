package pack

import (
	"compress/gzip"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"

	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-gonic/gin"
)

func ReceivePack(storage *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := "receive-pack"
		userName := c.Param("username")
		repoName := strings.TrimSuffix(c.Param("reponame"), ".git")

		repo, err := storage.GetRepository(userName, repoName)
		if err != nil {
			log.WithError(err).Errorf("GetRepository failed")
			c.String(http.StatusNotFound, "GetRepository failed: %v", err)
			return
		}
		var r io.Reader
		r = c.Request.Body
		encoding := c.GetHeader("Content-Encoding")
		switch encoding {
		case "gzip":
			r, err = gzip.NewReader(r)
			if err != nil {
				log.WithError(err).Errorf("gzip decode failed")
				c.String(http.StatusBadRequest, "gzip decode failed: %v", err)
				return
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
			log.WithError(err).Errorf("git command failed with: err:%v, stderr:%v", err, stderrBuf.String())
			c.String(http.StatusInternalServerError, "git command failed with: err:%v, stderr:%v", err, stderrBuf.String())
			return
		}
	}
}
