package home

import (
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Page(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User

		// 设置缓存控制头，使浏览器不会缓存响应
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		session := sessions.Default(c)

		userName, ok1 := session.Get("username").(string)
		userID, ok2 := session.Get("user_id").(uint)

		if ok1 && ok2 && db.Where("name = ? AND id = ?", userName, userID).First(&user).Error == nil {
			c.Redirect(http.StatusFound, "/"+userName)
			return
		}

		c.HTML(http.StatusOK, "home.html", nil)
	}
}
