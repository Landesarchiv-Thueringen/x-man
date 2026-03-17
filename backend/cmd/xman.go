package main

import (
	"lath/xman/internal/app"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("api", getDefaultResponse)
	router.Run()
}

func getDefaultResponse(c *gin.Context) {
	c.String(http.StatusOK, app.DefaultResponse)
}
