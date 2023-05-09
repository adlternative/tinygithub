package blob

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg/git/blob"
	gitRepo "github.com/adlternative/tinygithub/pkg/git/repo"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func Show(db *model.DBEngine, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {

		userName := c.Param("username")
		repoName := strings.TrimSuffix(c.Param("reponame"), ".git")
		blobPath := c.Param("blobpath")
		if strings.HasSuffix(blobPath, "/") {
			blobPath = strings.TrimSuffix(blobPath, "/")
		}
		if strings.HasPrefix(blobPath, "/") {
			blobPath = strings.TrimPrefix(blobPath, "/")
		}

		if blobPath == "" {
			c.HTML(http.StatusNotFound, "404.html", nil)
			return
		}
		var user model.User
		user.Name = userName

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
		blobContents, err := blob.ShowBlob(c, repo.Path(), revision)
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

func ShowV2(db *model.DBEngine, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		userName := c.Param("username")
		repoName := strings.TrimSuffix(c.Param("reponame"), ".git")
		revision := c.Query("revision")
		blobPath := c.Query("path")

		if strings.HasSuffix(blobPath, "/") {
			blobPath = strings.TrimSuffix(blobPath, "/")
		}
		if strings.HasPrefix(blobPath, "/") {
			blobPath = strings.TrimPrefix(blobPath, "/")
		}

		if blobPath == "" {
			c.JSON(http.StatusNotFound, nil)
			return
		}
		if revision == "" {
			revision = "HEAD"
		}

		var user model.User
		user.Name = userName

		if err := db.Where("name = ?", userName).Preload("Repositories", "name = ?", repoName).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": fmt.Sprintf("user %s repo %s not found", userName, repoName),
				})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		if len(user.Repositories) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("user %s repo %s not found", userName, repoName),
			})
			return
		}

		repo, err := store.GetRepository(userName, repoName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("user %s repo %s not found", userName, repoName),
			})
		}

		// git rev-parse
		isEmpty, _ := gitRepo.IsRepositoryEmpty(c, repo.Path())
		var blobContents []byte

		if !isEmpty {
			blobRevision := fmt.Sprintf("%s:%s", revision, blobPath)
			blobContents, err = blob.ShowBlob(c, repo.Path(), blobRevision)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": fmt.Sprintf("blob %s not found, error: %v\n", blobContents, err),
				})
				return
			}
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
