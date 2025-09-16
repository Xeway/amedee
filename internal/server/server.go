package server

import (
	"github.com/Xeway/amedee/internal/handler"
	"github.com/gin-gonic/gin"
)

func Run() {
	r := gin.Default()

	r.Static("/static", "./internal/static")
	r.LoadHTMLGlob("internal/template/*.html")

	r.GET("/", handler.Home)
	r.GET("/map", handler.Map)

	r.Run(":8080")
}
