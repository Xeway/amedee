package server

import (
	"os"

	"github.com/Xeway/amedee/internal/handler"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func Run() {
	r := gin.Default()

	// Do not trust any proxy by default (avoids "trust all proxies" warning)
	_ = r.SetTrustedProxies(nil)

	// Session middleware (cookie store). In production, set AMEDEE_SESSION_SECRET.
	secret := []byte(getSessionSecret())
	store := cookie.NewStore(secret)
	r.Use(sessions.Sessions("amedee_session", store))

	r.Static("/static", "./internal/static")
	r.LoadHTMLGlob("internal/template/*.html")

	r.GET("/", handler.Home)
	r.GET("/map", handler.Map)
	r.POST("/login", handler.Login)

	r.Run(":8080")
}

// getSessionSecret returns the session secret from env or a default dev value.
func getSessionSecret() string {
	if v := getenv("AMEDEE_SESSION_SECRET"); v != "" {
		return v
	}
	// NOTE: dev-only fallback; override in production with AMEDEE_SESSION_SECRET
	return "change-me-in-prod"
}

// small indirection to allow testing/mocking if needed
var getenv = func(key string) string {
	return os.Getenv(key)
}
