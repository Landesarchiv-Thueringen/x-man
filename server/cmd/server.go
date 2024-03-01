package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"lath/xman/internal/agency"
	"lath/xman/internal/archive/dimag"
	"lath/xman/internal/archive/filesystem"
	"lath/xman/internal/auth"
	"lath/xman/internal/clearing"
	"lath/xman/internal/db"
	"lath/xman/internal/messagestore"
	"lath/xman/internal/report"
	"lath/xman/internal/routines"
	"lath/xman/internal/tasks"
	"lath/xman/internal/xdomea"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const XMAN_VERSION = 1

var defaultResponse = "LATh xdomea server is running"

func main() {
	initServer()
	routines.Init()
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{})
	router.GET("api", getDefaultResponse)
	router.GET("api/login", auth.Login)
	authorized := router.Group("/")
	authorized.Use(auth.AuthRequired())
	authorized.GET("api/config", getConfig)
	authorized.GET("api/processes/my", getMyProcesses)
	authorized.GET("api/process/:id", getProcess)
	authorized.GET("api/messages/0501", get0501Messages)
	authorized.GET("api/messages/0503", get0503Messages)
	authorized.GET("api/message/:id", getMessageByID)
	authorized.GET("api/file-record-object/:id", getFileRecordObjectByID)
	authorized.GET("api/process-record-object/:id", getProcessRecordObjectByID)
	authorized.GET("api/document-record-object/:id", getDocumentRecordObjectByID)
	authorized.GET("api/record-object-appraisals", getRecordObjectAppraisals)
	authorized.GET("api/confidentiality-level-codelist", getConfidentialityLevelCodelist)
	authorized.GET("api/message-appraisal-complete/:id", isMessageAppraisalComplete)
	authorized.GET("api/all-record-objects-appraised/:id", AreAllRecordObjectsAppraised)
	authorized.GET("api/message-type-code/:id", getMessageTypeCode)
	authorized.GET("api/primary-document", getPrimaryDocument)
	authorized.GET("api/primary-documents/:id", getPrimaryDocuments)
	authorized.GET("api/report/:processId", getReport)
	authorized.GET("api/agencies/my", getMyAgencies)
	authorized.PATCH("api/file-record-object-appraisal", setFileRecordObjectAppraisal)
	authorized.PATCH("api/file-record-object-appraisal-note", setFileRecordObjectAppraisalNote)
	authorized.PATCH("api/process-record-object-appraisal", setProcessRecordObjectAppraisal)
	authorized.PATCH("api/process-record-object-appraisal-note", setProcessRecordObjectAppraisalNote)
	authorized.PATCH("api/process-note/:processId", setProcessNote)
	authorized.PATCH("api/finalize-message-appraisal/:id", finalizeMessageAppraisal)
	authorized.PATCH("api/multi-appraisal", setAppraisalForMultipleRecordObjects)
	authorized.PATCH("api/archive-0503-message/:id", archive0503Message)
	admin := router.Group("/")
	admin.Use(auth.AdminRequired())
	admin.GET("api/processes", getProcesses)
	admin.DELETE("api/process/:id", deleteProcess)
	admin.GET("api/processing-errors", getProcessingErrors)
	admin.POST("api/processing-errors/resolve/:id", resolveProcessingError)
	admin.GET("api/users", auth.Users)
	admin.GET("api/agencies", getAgencies)
	admin.PUT("api/agency", putAgency)
	admin.POST("api/agency/:id", postAgency)
	admin.DELETE("api/agency/:id", deleteAgency)
	admin.GET("api/collections", getCollections)
	admin.PUT("api/collection", putCollection)
	admin.POST("api/collection/:id", postCollection)
	admin.DELETE("api/collection/:id", deleteCollection)
	admin.POST("api/test-transfer-dir", testTransferDir)
	admin.GET("api/tasks", getTasks)
	addr := "0.0.0.0:80"
	router.Run(addr)
}

func initServer() {
	log.Println(defaultResponse)
	db.Init()
	// It's important to the migrate after the database initialization.
	MigrateData()
	agency.MonitorTransferDirs()
}

func MigrateData() {
	xManVersion := db.GetXManVersion()
	if xManVersion == 0 {
		log.Printf("Migrating database from X-Man version %d to %d... ", xManVersion, XMAN_VERSION)
		db.Migrate()
		xdomea.InitMessageTypes()
		xdomea.InitXdomeaVersions()
		xdomea.InitRecordObjectAppraisals()
		xdomea.InitConfidentialityLevelCodelist()
		xdomea.InitMediumCodelist()
		agency.InitAgencies()
		db.SetXManVersion(XMAN_VERSION)
		log.Println("done")
	} else {
		log.Printf("Database is up do date with X-Man version %d\n", XMAN_VERSION)
	}
}

func getDefaultResponse(context *gin.Context) {
	context.String(http.StatusOK, defaultResponse)
}

func getConfig(context *gin.Context) {
	deleteArchivedProcessesAfterDays, _ := strconv.Atoi(os.Getenv("DELETE_ARCHIVED_PROCESSES_AFTER_DAYS"))
	context.JSON(http.StatusOK, gin.H{
		"deleteArchivedProcessesAfterDays": deleteArchivedProcessesAfterDays,
	})
}

func getProcessingErrors(context *gin.Context) {
	processingErrors := db.GetProcessingErrors()
	context.JSON(http.StatusOK, processingErrors)
}

func resolveProcessingError(context *gin.Context) {
	id, err := strconv.ParseUint(context.Param("id"), 10, 32)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	processingError, found := db.GetProcessingError(uint(id))
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	clearing.Resolve(processingError, db.ProcessingErrorResolution(body))
	context.Status(http.StatusAccepted)
}

func getProcess(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	process, found := db.GetProcess(id.String())
	if !found {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	context.JSON(http.StatusOK, process)
}

func getMyProcesses(context *gin.Context) {
	userID := context.MustGet("userId").([]byte)
	processes := db.GetProcessesForUser(userID)
	context.JSON(http.StatusOK, processes)
}

func getProcesses(context *gin.Context) {
	processes := db.GetProcesses()
	context.JSON(http.StatusOK, processes)
}

func deleteProcess(context *gin.Context) {
	id := context.Param("id")
	if found := messagestore.DeleteProcess(id); found {
		context.Status(http.StatusAccepted)
	} else {
		context.Status(http.StatusNotFound)
	}
}

func getMessageByID(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	message, found := db.GetCompleteMessageByID(id)
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
		return
	}
	context.JSON(http.StatusOK, message)
}

func getFileRecordObjectByID(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	fileRecordObject, err := db.GetFileRecordObjectByID(id)
	if err != nil {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	context.JSON(http.StatusOK, fileRecordObject)
}

func getProcessRecordObjectByID(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	processRecordObject, found := db.GetProcessRecordObjectByID(id)
	if !found {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	context.JSON(http.StatusOK, processRecordObject)
}

func getDocumentRecordObjectByID(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	documentRecordObject, err := db.GetDocumentRecordObjectByID(id)
	if err != nil {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	context.JSON(http.StatusOK, documentRecordObject)
}

func get0501Messages(context *gin.Context) {
	messages := db.GetMessagesByCode("0501")
	context.JSON(http.StatusOK, messages)
}

func get0503Messages(context *gin.Context) {
	messages := db.GetMessagesByCode("0503")
	context.JSON(http.StatusOK, messages)
}

func getRecordObjectAppraisals(context *gin.Context) {
	appraisals := db.GetRecordObjectAppraisals()
	context.JSON(http.StatusOK, appraisals)
}

func getConfidentialityLevelCodelist(context *gin.Context) {
	codelist := db.GetConfidentialityLevelCodelist()
	context.JSON(http.StatusOK, codelist)
}

func setFileRecordObjectAppraisal(context *gin.Context) {
	fileRecordObjectID := context.Query("id")
	id, err := uuid.Parse(fileRecordObjectID)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	appraisalCode := context.Query("appraisal")
	fileRecordObject, err := db.SetFileRecordObjectAppraisal(id, appraisalCode, true)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	context.JSON(http.StatusOK, fileRecordObject)
}

func setFileRecordObjectAppraisalNote(context *gin.Context) {
	fileRecordObjectID := context.Query("id")
	id, err := uuid.Parse(fileRecordObjectID)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	note := context.Query("note")
	fileRecordObject, err := db.SetFileRecordObjectAppraisalNote(id, note)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	context.JSON(http.StatusOK, fileRecordObject)
}

func setProcessRecordObjectAppraisal(context *gin.Context) {
	processRecordObjectID := context.Query("id")
	id, err := uuid.Parse(processRecordObjectID)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	appraisalCode := context.Query("appraisal")
	processRecordObject, found := db.GetProcessRecordObjectByID(id)
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
		return
	}
	err = db.SetProcessRecordObjectAppraisal(&processRecordObject, appraisalCode)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	context.JSON(http.StatusOK, processRecordObject)
}

func setProcessRecordObjectAppraisalNote(context *gin.Context) {
	processRecordObjectID := context.Query("id")
	id, err := uuid.Parse(processRecordObjectID)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	note := context.Query("note")
	processRecordObject, found := db.GetProcessRecordObjectByID(id)
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
		return
	}
	db.SetProcessRecordObjectAppraisalNote(&processRecordObject, note)
	context.JSON(http.StatusOK, processRecordObject)
}

func setProcessNote(context *gin.Context) {
	processId := context.Param("processId")
	note, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	process, found := db.GetProcess(processId)
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
	}
	db.SetProcessNote(process, string(note))
}

type MultiAppraisalBody struct {
	FileRecordObjectIDs    []string `json:"fileRecordObjectIDs"`
	ProcessRecordObjectIDs []string `json:"processRecordObjectIDs"`
	AppraisalCode          string   `json:"appraisalCode"`
	AppraisalNote          *string  `json:"appraisalNote"`
}

type MultiAppraisalResponse struct {
	UpdatedFileRecordObjects    []db.FileRecordObject    `json:"updatedFileRecordObjects"`
	UpdatedProcessRecordObjects []db.ProcessRecordObject `json:"updatedProcessRecordObjects"`
}

func setAppraisalForMultipleRecordObjects(context *gin.Context) {
	jsonBody, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	var parsedBody MultiAppraisalBody
	err = json.Unmarshal(jsonBody, &parsedBody)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	updatedFileRecordObjects, err := xdomea.SetAppraisalForFileRecordObjects(
		parsedBody.FileRecordObjectIDs,
		parsedBody.AppraisalCode,
		parsedBody.AppraisalNote,
	)
	if err != nil {
		context.AbortWithError(http.StatusBadRequest, err)
		return
	}
	updatedProcessRecordObjects, err := xdomea.SetAppraisalForProcessRecordObjects(
		parsedBody.ProcessRecordObjectIDs,
		parsedBody.AppraisalCode,
		parsedBody.AppraisalNote,
	)
	if err != nil {
		context.AbortWithError(http.StatusBadRequest, err)
		return
	}
	response := MultiAppraisalResponse{
		UpdatedFileRecordObjects:    updatedFileRecordObjects,
		UpdatedProcessRecordObjects: updatedProcessRecordObjects,
	}
	context.JSON(http.StatusOK, response)
}

func finalizeMessageAppraisal(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	message, found := db.GetCompleteMessageByID(id)
	if !found {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	if message.AppraisalComplete {
		// appraisal for message is already complete
		context.AbortWithStatus(http.StatusBadRequest)
		return
	}
	userID := context.MustGet("userId").([]byte)
	userName := auth.GetDisplayName(userID)
	message = xdomea.FinalizeMessageAppraisal(message, userName)
	messagePath := messagestore.Store0502Message(message)
	process := db.GetProcessForMessage(message)
	process.Message0502Path = &messagePath
	db.UpdateProcess(process)
}

func isMessageAppraisalComplete(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	appraisalComplete, err := db.IsMessageAppraisalComplete(id)
	if err != nil {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	context.JSON(http.StatusOK, appraisalComplete)
}

func AreAllRecordObjectsAppraised(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	message, found := db.GetCompleteMessageByID(id)
	if !found {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	appraisalComplete := xdomea.AreAllRecordObjectsAppraised(message)
	context.JSON(http.StatusOK, appraisalComplete)
}

func getMessageTypeCode(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	messageTypeCode, err := db.GetMessageTypeCode(id)
	if err != nil {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	context.JSON(http.StatusOK, messageTypeCode)
}

func getPrimaryDocument(context *gin.Context) {
	messageIDParam := context.Query("messageID")
	messageID, err := uuid.Parse(messageIDParam)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	primaryDocumentIDParam := context.Query("primaryDocumentID")
	primaryDocumentID, err := strconv.ParseUint(primaryDocumentIDParam, 10, 32)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	path, err := db.GetPrimaryFileStorePath(messageID, uint(primaryDocumentID))
	if err != nil {
		context.AbortWithError(http.StatusNotFound, err)
		return
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
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	primaryDocuments := db.GetAllPrimaryDocumentsWithFormatVerification(messageID)
	context.JSON(http.StatusOK, primaryDocuments)
}

func getReport(context *gin.Context) {
	processID := context.Param("processId")
	process, found := db.GetProcess(processID)
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
		return
	}
	values, err := report.GetReportData(process)
	if err != nil {
		context.AbortWithError(http.StatusBadRequest, err)
		return
	}
	jsonValue, _ := json.Marshal(values)
	resp, err := http.Post("http://report/render", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		panic(err)
	} else if resp.StatusCode != http.StatusOK {
		context.String(http.StatusInternalServerError, "Internal server error")
		body, _ := io.ReadAll(resp.Body)
		println(string(body))
		return
	}
	context.DataFromReader(http.StatusOK, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

// archive0503Message archives all metadata and primary files in the digital archive.
func archive0503Message(context *gin.Context) {
	messageID, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	message, found := db.GetCompleteMessageByID(messageID)
	if !found {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	process, found := db.GetProcess(message.MessageHead.ProcessID)
	if !found {
		context.AbortWithError(http.StatusNotFound, err)
		return
	}
	if !process.IsArchivable() {
		context.AbortWithError(http.StatusBadRequest, errors.New("message can't be archived"))
		return
	}
	userID := context.MustGet("userId").([]byte)
	userName := auth.GetDisplayName(userID)
	task := tasks.Start(db.TaskTypeArchiving, process, 0)
	go func() {
		switch archiveTarget := os.Getenv("ARCHIVE_TARGET"); archiveTarget {
		case "filesystem":
			err = filesystem.ArchiveMessage(process, message)
		case "dimag":
			err = dimag.ImportMessageSync(process, message)
		default:
			panic("unknown archive target: " + archiveTarget)
		}
		if err != nil {
			tasks.MarkFailed(&task, err.Error(), true)
		} else {
			tasks.MarkDone(&task, &userName)
		}
	}()
}

func getMyAgencies(context *gin.Context) {
	userID := context.MustGet("userId").([]byte)
	agencies := db.GetAgenciesForUser(userID)
	context.JSON(http.StatusOK, agencies)
}

func getAgencies(context *gin.Context) {
	var agencies []db.Agency
	if userIDString, hasUserID := context.GetQuery("userId"); hasUserID {
		userID, err := base64.StdEncoding.DecodeString(userIDString)
		if err != nil {
			context.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}
		agencies = db.GetAgenciesForUser(userID)
	} else if collectionIDString, hasCollectionID := context.GetQuery("collectionId"); hasCollectionID {
		collectionID, err := strconv.ParseUint(collectionIDString, 10, 32)
		if err != nil {
			context.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}
		agencies = db.GetAgenciesForCollection(uint(collectionID))
	} else {
		agencies = db.GetAgencies()
	}
	context.JSON(http.StatusOK, agencies)
}

func putAgency(context *gin.Context) {
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	var agency db.Agency
	err = json.Unmarshal(body, &agency)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	id, err := db.CreateAgency(agency)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	context.String(http.StatusAccepted, strconv.FormatUint(uint64(id), 10))
}

func postAgency(context *gin.Context) {
	idParam := context.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	var agency db.Agency
	err = json.Unmarshal(body, &agency)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	err = db.UpdateAgency(uint(id), agency)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	context.Status(http.StatusAccepted)
}

func deleteAgency(context *gin.Context) {
	idParam := context.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	found := db.DeleteAgency(uint(id))
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
		return
	}
	context.Status(http.StatusAccepted)
}

func getCollections(context *gin.Context) {
	Collections := db.GetCollections()
	context.JSON(http.StatusOK, Collections)
}

func putCollection(context *gin.Context) {
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	var Collection db.Collection
	err = json.Unmarshal(body, &Collection)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	id, err := db.CreateCollection(Collection)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	context.String(http.StatusAccepted, strconv.FormatUint(uint64(id), 10))
}

func postCollection(context *gin.Context) {
	idParam := context.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	var Collection db.Collection
	err = json.Unmarshal(body, &Collection)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	err = db.UpdateCollection(uint(id), Collection)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	context.Status(http.StatusAccepted)
}

func deleteCollection(context *gin.Context) {
	idParam := context.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	found := db.DeleteCollection(uint(id))
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
		return
	}
	context.Status(http.StatusAccepted)
}

func testTransferDir(context *gin.Context) {
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	success := agency.TestTransferDir(string(body))
	if success {
		context.JSON(http.StatusOK, gin.H{"result": "success"})
	} else {
		context.JSON(http.StatusOK, gin.H{"result": "failed"})
	}
}

func getTasks(context *gin.Context) {
	tasks := db.GetTasks()
	context.JSON(http.StatusOK, tasks)
}
