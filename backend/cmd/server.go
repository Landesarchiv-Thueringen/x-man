package main

import (
	"fmt"
	"lath/xdomea/internal/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("LATh xdomea server is running")
	db.Init()
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1"})
	router.GET("", getDefaultResponse)
	router.Run("localhost:3000")
}

func getDefaultResponse(context *gin.Context) {
	context.String(http.StatusOK, "LATh xdomea server is running")
}
