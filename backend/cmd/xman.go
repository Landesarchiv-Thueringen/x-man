package main

import (
	"lath/xman/internal/app"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	err := app.Init()
	if err != nil {
		log.Fatal(err)
	}
	router := gin.Default()
	router.GET("api", getDefaultResponse)
	router.Run()
}

func getDefaultResponse(c *gin.Context) {
	c.String(http.StatusOK, app.DefaultResponse)
}
