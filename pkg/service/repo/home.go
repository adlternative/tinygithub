package repo

import (
	"bufio"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"github.com/adlternative/tinygithub/pkg/config"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/service/git/tree"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func Home(db *model.DBEngine, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionUserName := session.Get("username").(string)
		if sessionUserName == "" {
			c.Redirect(http.StatusFound, "/user/login")
			return
		}
		sessionUserID := session.Get("user_id").(uint)

		userName := c.Param("username")
		repoName := c.Param("reponame")

		var user model.User
		user.Name = sessionUserName
		user.ID = sessionUserID

		if err := db.Where("name = ?", userName).Preload("Repositories", "name = ?", repoName).First(&user).Error; err != nil {
			// 处理错误
			if err == gorm.ErrRecordNotFound {
				c.HTML(http.StatusNotFound, "404.html", nil)
				return
			} else {
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		if len(user.Repositories) == 0 {
			c.HTML(http.StatusNotFound, "404.html", nil)
			return
		} else if len(user.Repositories) > 1 {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": "multiple repo same name",
			})
			return
		}

		repo, err := store.GetRepository(userName, repoName)
		if err != nil {
			c.HTML(http.StatusNotFound, "404.html", nil)
			return
		}
		var stderrBuf strings.Builder
		// git -c <repoPath> upload-pack --advertise-refs --stateless-rpc <repoPath>
		// git -c <repoPath> receive-pack --advertise-refs --stateless-rpc <repoPath>

		gitCmd := cmd.NewGitCommand("ls-tree").WithGitDir(repo.Path()).
			WithArgs("HEAD").
			WithStderr(&stderrBuf)

		if err = gitCmd.Start(c); err != nil {
			log.WithError(err).Errorf("git command start failed with: err:%v, stderr:%v", err, stderrBuf.String())
			c.String(http.StatusInternalServerError, "git command start  failed with: err:%v, stderr:%v", err, stderrBuf.String())
			return
		}

		var entries []*tree.Entry
		scanner := bufio.NewScanner(gitCmd)

		for scanner.Scan() {
			entry, err := tree.Parse(scanner.Text())
			if err != nil {
				return
			}
			entries = append(entries, entry)
		}
		if err = scanner.Err(); err != nil {
			log.WithError(err).Errorf("scanner failed")
			c.String(http.StatusInternalServerError, "scanner failed with: err:%v, stderr:%v", err, stderrBuf.String())
			return
		}

		if err = gitCmd.Wait(); err != nil {
			log.WithError(err).Errorf("git command failed with stderr:%v", stderrBuf.String())
			c.String(http.StatusInternalServerError, "git command failed with: err:%v, stderr:%v", err, stderrBuf.String())
			return
		}

		c.HTML(http.StatusOK, "repo.html", gin.H{
			"RepoName":    user.Repositories[0].Name,
			"Description": user.Repositories[0].Desc,
			"Owner":       userName,
			"DownloadURL": fmt.Sprintf("http://%s:%s/%s/%s.git", viper.GetString(config.ServerIp), viper.GetString(config.ServerPort), userName, repoName),
			"TreeEntries": entries,
		})
	}
}
