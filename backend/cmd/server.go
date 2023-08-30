package main

import (
	"flag"
	"lath/xdomea/internal/db"
	"lath/xdomea/internal/transferdir"
	"lath/xdomea/internal/xdomea"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var defaultResponse = "LATh xdomea server is running"

func main() {
	initServer()
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1"})
	router.GET("", getDefaultResponse)
	router.Run("localhost:3000")
}

func initServer() {
	log.Println(defaultResponse)
	db.Init()
	// It's important to process the flags after the database initialization.
	processFlags()
	go transferdir.Watch("transfer/lpd", "transfer/aaj")
}

func getDefaultResponse(context *gin.Context) {
	context.String(http.StatusOK, defaultResponse)
}

func processFlags() {
	initFlag := flag.Bool("init", false, "initialize database")
	flag.Parse()
	if *initFlag {
		db.Migrate()
		xdomea.InitMessageTypes()
	}
}
