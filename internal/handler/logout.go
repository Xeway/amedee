package handler

import (
	"net/http"

	"github.com/Xeway/amedee/internal/global"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Logout(c *gin.Context) {

	sess := sessions.Default(c)
	sess.Delete(global.SessionKey)
	sess.Save()
	c.Redirect(http.StatusSeeOther, "/")
}
