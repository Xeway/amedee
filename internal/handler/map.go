package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Map(c *gin.Context) {
	c.HTML(http.StatusOK, "map.html", nil)
}
