package blob

import (
	"context"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/cmd"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/buffer"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func showBlob(ctx context.Context, repoPath, revision string) ([]byte, error) {
	var stderrBuf strings.Builder
	var stdoutBuf buffer.Buffer

	gitCmd := cmd.NewGitCommand("cat-file").WithGitDir(repoPath).
		WithOptions("-p").
		WithArgs(revision).
		WithStderr(&stderrBuf).
		WithStdout(&stdoutBuf)

	if err := gitCmd.Start(ctx); err != nil {
		return nil, fmt.Errorf("gitCmd start failed with %w", err)
	}

	if err := gitCmd.Wait(); err != nil {
		return nil, fmt.Errorf("git command failed with stderr:%v, error:%w", stderrBuf.String(), err)
	}
	return stdoutBuf.Bytes(), nil
}

func Show(db *model.DBEngine, store *storage.Storage) gin.HandlerFunc {
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
		blobPath := c.Param("blobpath")
		if strings.HasSuffix(blobPath, "/") {
			blobPath = strings.TrimSuffix(blobPath, "/")
		}
		if strings.HasPrefix(blobPath, "/") {
			blobPath = strings.TrimPrefix(blobPath, "/")
		}
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
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		revision := fmt.Sprintf("HEAD:%s", blobPath)
		blobContents, err := showBlob(c, repo.Path(), revision)
		if err != nil {
			c.HTML(http.StatusNotFound, "404.html", nil)
			return
		}
		// 判断文件类型
		contentType := http.DetectContentType(blobContents)
		//isBinary := false
		switch {
		case strings.HasPrefix(contentType, "text/"):
			// 文本类型文件，直接显示
		case strings.HasPrefix(contentType, "image/"):
			// 图片类型文件，返回图片
			break
		default:
			// 其他类型文件，显示"二进制文件"
			contentType = "text/plain"
			//isBinary = true
			blobContents = []byte("binary file")
		}
		c.Data(http.StatusOK, contentType, blobContents)
	}
}
