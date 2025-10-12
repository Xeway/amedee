package handler

import (
	"net/http"

	"github.com/Xeway/amedee/internal/session"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func LoggedIn(c *gin.Context) {
	sess := sessions.Default(c)
	isLoggedIn := sess.Get(session.SessionKey) != nil

	if isLoggedIn {
		c.HTML(http.StatusOK, "links.html", gin.H{"LoggedIn": true})
	} else {
		c.HTML(http.StatusOK, "login_modal.html", nil)
	}
}
