package handler

import (
	"net/http"

	"github.com/Xeway/amedee/internal/session"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Home(c *gin.Context) {
	sess := sessions.Default(c)
	has := sess.Get(session.SessionKey) != nil
	c.HTML(http.StatusOK, "index.html", gin.H{"LoggedInToThirdParty": has})
}
