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

	// Session middleware (cookie store)
	secret := []byte(getSessionSecret())
	store := cookie.NewStore(secret)
	r.Use(sessions.Sessions("amedee_session", store))

	r.Static("/static", "internal/static")
	r.LoadHTMLGlob("internal/template/*.html")

	r.GET("/", handler.Home)
	r.GET("/map", handler.Map)
	r.POST("/login", handler.Login)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(port)
}

func getSessionSecret() string {
	if v := os.Getenv("AMEDEE_SESSION_SECRET"); v != "" {
		return v
	}
	return "jaimelescookies"
}
