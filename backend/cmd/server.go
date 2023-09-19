package main

import (
	"flag"
	"lath/xdomea/internal/db"
	"lath/xdomea/internal/transferdir"
	"lath/xdomea/internal/xdomea"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var defaultResponse = "LATh xdomea server is running"

func main() {
	initServer()
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"http://127.0.0.1"})
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:4200"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type"}
	corsConfig.AllowMethods = []string{"GET", "PATCH"}
	// It's important that the cors configuration is used before declaring the routes
	router.Use(cors.New(corsConfig))
	router.GET("", getDefaultResponse)
	router.GET("messages/0501", get0501Messages)
	router.GET("messages/0503", get0503Messages)
	router.GET("message/:id", getMessageByID)
	router.GET("file-record-object/:id", getFileRecordObjectByID)
	router.GET("process-record-object/:id", getProcessRecordObjectByID)
	router.GET("document-record-object/:id", getDocumentRecordObjectByID)
	router.GET("record-object-appraisals", getRecordObjectAppraisal)
	router.PATCH("file-record-object-appraisal", setFileRecordObjectAppraisal)
	router.PATCH("process-record-object-appraisal", setProcessRecordObjectAppraisal)
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

func getMessageByID(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	message, err := db.GetMessageByID(uint(id))
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, message)
	}
}

func getFileRecordObjectByID(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	fileRecordObject, err := db.GetFileRecordObjectByID(uint(id))
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, fileRecordObject)
	}
}

func getProcessRecordObjectByID(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	processRecordObject, err := db.GetProcessRecordObjectByID(uint(id))
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, processRecordObject)
	}
}

func getDocumentRecordObjectByID(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	documentRecordObject, err := db.GetDocumentRecordObjectByID(uint(id))
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, documentRecordObject)
	}
}

func get0501Messages(context *gin.Context) {
	messages, err := db.GetMessagesByCode("0501")
	if err != nil {
		log.Fatal(err)
	}
	context.JSON(http.StatusOK, messages)
}

func get0503Messages(context *gin.Context) {
	messages, err := db.GetMessagesByCode("0503")
	if err != nil {
		log.Fatal(err)
	}
	context.JSON(http.StatusOK, messages)
}

func getRecordObjectAppraisal(context *gin.Context) {
	appraisals, err := db.GetRecordObjectAppraisals()
	if err != nil {
		log.Fatal(err)
	}
	context.JSON(http.StatusOK, appraisals)
}

func setFileRecordObjectAppraisal(context *gin.Context) {
	fileRecordObjectID := context.Query("id")
	id, err := strconv.ParseUint(fileRecordObjectID, 10, 32)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	appraisalCode := context.Query("appraisal")
	err = db.SetFileRecordObjectAppraisal(uint(id), appraisalCode)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
}

func setProcessRecordObjectAppraisal(context *gin.Context) {
	processRecordObjectID := context.Query("id")
	id, err := strconv.ParseUint(processRecordObjectID, 10, 32)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	appraisalCode := context.Query("appraisal")
	err = db.SetProcessRecordObjectAppraisal(uint(id), appraisalCode)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
}

func processFlags() {
	initFlag := flag.Bool("init", false, "initialize database")
	flag.Parse()
	if *initFlag {
		db.Migrate()
		xdomea.InitMessageTypes()
		xdomea.InitRecordObjectAppraisals()
	}
}
