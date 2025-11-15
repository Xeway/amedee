package middleware

import (
	"net/http"

	"github.com/Xeway/amedee/internal/session"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func IsConnectedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		isLoggedIn := sess.Get(session.SessionKey) != nil

		if !isLoggedIn {
			c.Redirect(http.StatusSeeOther, "/")
		} else {
			c.Next()
		}
	}
}
