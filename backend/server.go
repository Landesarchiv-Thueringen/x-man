package main

import (
  "fmt"
  "net/http"
  "github.com/gin-gonic/gin"
)

func main() {
  fmt.Println("LATh xdomea server is running")
  router := gin.Default()
  router.GET("", getDefaultResponse)
  router.Run("localhost:3000")
}

func getDefaultResponse(context *gin.Context) {
  context.String(http.StatusOK, "LATh xdomea server is running")
}