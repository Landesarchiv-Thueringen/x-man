package main

import (
	"flag"
	"lath/xman/internal/db"
	"lath/xman/internal/messagestore"
	"lath/xman/internal/transferdir"
	"lath/xman/internal/xdomea"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var defaultResponse = "LATh xdomea server is running"

func main() {
	initServer()
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"*"})
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type"}
	corsConfig.AllowMethods = []string{"GET", "PATCH"}
	// It's important that the cors configuration is used before declaring the routes
	router.Use(cors.New(corsConfig))
	router.GET("", getDefaultResponse)
	router.GET("processing-errors", getProcessingErrors)
	router.GET("processes", getProcesses)
	router.GET("messages/0501", get0501Messages)
	router.GET("messages/0503", get0503Messages)
	router.GET("message/:id", getMessageByID)
	router.GET("file-record-object/:id", getFileRecordObjectByID)
	router.GET("process-record-object/:id", getProcessRecordObjectByID)
	router.GET("document-record-object/:id", getDocumentRecordObjectByID)
	router.GET("record-object-appraisals", getRecordObjectAppraisals)
	router.GET("record-object-confidentialities", getRecordObjectConfidentialities)
	router.GET("message-appraisal-complete/:id", isMessageAppraisalComplete)
	router.GET("message-type-code/:id", getMessageTypeCode)
	router.GET("primary-document", getPrimaryDocument)
	router.GET("primary-documents/:id", getPrimaryDocuments)
	router.PATCH("file-record-object-appraisal", setFileRecordObjectAppraisal)
	router.PATCH("process-record-object-appraisal", setProcessRecordObjectAppraisal)
	router.PATCH("finalize-message-appraisal/:id", finalizeMessageAppraisal)
	router.Run("localhost:3000")
}

func initServer() {
	log.Println(defaultResponse)
	db.Init()
	// It's important to process the flags after the database initialization.
	processFlags()
	go transferdir.Watch("transfer/tmik")
}

func getDefaultResponse(context *gin.Context) {
	context.String(http.StatusOK, defaultResponse)
}

func getProcessingErrors(context *gin.Context) {
	processingErrors := db.GetProcessingErrors()
	context.JSON(http.StatusOK, processingErrors)
}

func getProcesses(context *gin.Context) {
	processes, err := db.GetProcesses()
	if err != nil {
		context.JSON(http.StatusInternalServerError, processes)
	} else {
		context.JSON(http.StatusOK, processes)
	}
}

func getMessageByID(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	message, err := db.GetCompleteMessageByID(id)
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, message)
	}
}

func getFileRecordObjectByID(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	fileRecordObject, err := db.GetFileRecordObjectByID(id)
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, fileRecordObject)
	}
}

func getProcessRecordObjectByID(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	processRecordObject, err := db.GetProcessRecordObjectByID(id)
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, processRecordObject)
	}
}

func getDocumentRecordObjectByID(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	documentRecordObject, err := db.GetDocumentRecordObjectByID(id)
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

func getRecordObjectAppraisals(context *gin.Context) {
	appraisals, err := db.GetRecordObjectAppraisals()
	if err != nil {
		log.Fatal(err)
	}
	context.JSON(http.StatusOK, appraisals)
}

func getRecordObjectConfidentialities(context *gin.Context) {
	appraisals, err := db.GetRecordObjectConfidentialities()
	if err != nil {
		log.Fatal(err)
	}
	context.JSON(http.StatusOK, appraisals)
}

func setFileRecordObjectAppraisal(context *gin.Context) {
	fileRecordObjectID := context.Query("id")
	id, err := uuid.Parse(fileRecordObjectID)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	appraisalCode := context.Query("appraisal")
	fileRecordObject, err := db.SetFileRecordObjectAppraisal(id, appraisalCode)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	context.JSON(http.StatusOK, fileRecordObject)
}

func setProcessRecordObjectAppraisal(context *gin.Context) {
	processRecordObjectID := context.Query("id")
	id, err := uuid.Parse(processRecordObjectID)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	appraisalCode := context.Query("appraisal")
	processRecordObject, err := db.SetProcessRecordObjectAppraisal(id, appraisalCode)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	context.JSON(http.StatusOK, processRecordObject)
}

func finalizeMessageAppraisal(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	message, err := db.GetCompleteMessageByID(id)
	if err != nil {
		// message couldn't be found
		context.JSON(http.StatusNotFound, err)
	} else if message.AppraisalComplete {
		// appraisal for message is already complete
		context.AbortWithStatus(http.StatusBadRequest)
	} else {
		message.AppraisalComplete = true
		err = db.UpdateMessage(message)
		if err != nil {
			context.JSON(http.StatusInternalServerError, err)
		}
		messagestore.Store0502Message(message)
	}
}

func isMessageAppraisalComplete(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	appraisalComplete, err := db.IsMessageAppraisalComplete(id)
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, appraisalComplete)
	}
}

func getMessageTypeCode(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	messageTypeCode, err := db.GetMessageTypeCode(id)
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, messageTypeCode)
	}
}

func getPrimaryDocument(context *gin.Context) {
	messageIDParam := context.Query("messageID")
	messageID, err := uuid.Parse(messageIDParam)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	primaryDocumentIDParam := context.Query("primaryDocumentID")
	primaryDocumentID, err := strconv.ParseUint(primaryDocumentIDParam, 10, 32)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	path, err := db.GetPrimaryFileStorePath(messageID, uint(primaryDocumentID))
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	}
	fileName := filepath.Base(path)
	// context.Header("Content-Description", "File Transfer")
	context.Header("Content-Transfer-Encoding", "binary")
	context.Header("Content-Disposition", "attachment; filename="+fileName)
	context.Header("Content-Type", "application/octet-stream")
	context.FileAttachment(path, fileName)
}

func getPrimaryDocuments(context *gin.Context) {
	messageID, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, err)
	}
	primaryDocuments, err := db.GetAllPrimaryDocuments(messageID)
	if err != nil {
		context.JSON(http.StatusNotFound, err)
	} else {
		context.JSON(http.StatusOK, primaryDocuments)
	}
}

func processFlags() {
	initFlag := flag.Bool("init", false, "initialize database")
	flag.Parse()
	if *initFlag {
		db.Migrate()
		xdomea.InitMessageTypes()
		xdomea.InitXdomeaVersions()
		xdomea.InitRecordObjectAppraisals()
		xdomea.InitRecordObjectConfidentialities()
	}
}