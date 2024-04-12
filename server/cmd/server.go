package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lath/xman/internal/archive/dimag"
	"lath/xman/internal/archive/filesystem"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/mail"
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

const XMAN_MAJOR_VERSION = 1
const XMAN_MINOR_VERSION = 0
const XMAN_PATCH_VERSION = 0

var XMAN_VERSION = fmt.Sprintf("%d.%d.%d", XMAN_MAJOR_VERSION, XMAN_MINOR_VERSION, XMAN_PATCH_VERSION)

var defaultResponse = fmt.Sprintf("x-man server %s is running", XMAN_VERSION)

func main() {
	initServer()
	routines.Init()
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{})
	router.GET("api", getDefaultResponse)
	router.GET("api/about", getAbout)
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
	authorized.GET("api/appraisal-codelist", getAppraisalCodelist)
	authorized.GET("api/confidentiality-level-codelist", getConfidentialityLevelCodelist)
	authorized.GET("api/all-record-objects-appraised/:id", AreAllRecordObjectsAppraised)
	authorized.GET("api/message-type-code/:id", getMessageTypeCode)
	authorized.GET("api/primary-document", getPrimaryDocument)
	authorized.GET("api/primary-documents/:id", getPrimaryDocuments)
	authorized.GET("api/report/:processId", getReport)
	authorized.GET("api/user-info/my", getMyUserInformation)
	authorized.POST("api/user-preferences", setUserPreferences)
	authorized.GET("api/appraisals/:processId", getAppraisals)
	authorized.POST("api/appraisal-decision", setAppraisalDecision)
	authorized.POST("api/appraisal-note", setAppraisalNote)
	authorized.POST("api/appraisals", setAppraisals)
	authorized.PATCH("api/finalize-message-appraisal/:id", finalizeMessageAppraisal)
	authorized.PATCH("api/archive-0503-message/:id", archive0503Message)
	authorized.PATCH("api/process-note/:processId", setProcessNote)
	admin := router.Group("/")
	admin.Use(auth.AdminRequired())
	admin.GET("api/processes", getProcesses)
	admin.DELETE("api/process/:id", deleteProcess)
	admin.GET("api/processing-errors", getProcessingErrors)
	admin.POST("api/processing-errors/resolve/:id", resolveProcessingError)
	admin.GET("api/users", Users)
	admin.GET("api/user-info", getUserInformation)
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
	admin.GET("api/collectionDimagIds", getCollectionDimagIDs)
	addr := "0.0.0.0:80"
	router.Run(addr)
}

func initServer() {
	log.Println(defaultResponse)
	db.Init()
	// It's important to the migrate after the database initialization.
	MigrateData()
	xdomea.MonitorTransferDirs()
}

func MigrateData() {
	major, minor, patch := db.GetXManVersion()
	if major == 0 {
		log.Printf("Migrating database from X-Man version %d.%d.%d to %s... ", major, minor, patch, XMAN_VERSION)
		db.Migrate()
		xdomea.InitMessageTypes()
		xdomea.InitXdomeaVersions()
		xdomea.InitRecordObjectAppraisals()
		xdomea.InitConfidentialityLevelCodelist()
		xdomea.InitMediumCodelist()
		xdomea.InitAgencies()
		db.SetXManVersion(XMAN_MAJOR_VERSION, XMAN_MINOR_VERSION, XMAN_PATCH_VERSION)
		log.Println("done")
	} else {
		log.Printf("Database is up do date with X-Man version %s\n", XMAN_VERSION)
	}
}

func getDefaultResponse(context *gin.Context) {
	context.String(http.StatusOK, defaultResponse)
}

func getAbout(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"version": XMAN_VERSION,
	})
}

func getConfig(context *gin.Context) {
	deleteArchivedProcessesAfterDays, _ := strconv.Atoi(os.Getenv("DELETE_ARCHIVED_PROCESSES_AFTER_DAYS"))
	supportsEmailNotifications := os.Getenv("SMTP_SERVER") != ""
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	context.JSON(http.StatusOK, gin.H{
		"deleteArchivedProcessesAfterDays": deleteArchivedProcessesAfterDays,
		"supportsEmailNotifications":       supportsEmailNotifications,
		"archiveTarget":                    archiveTarget,
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
	xdomea.Resolve(processingError, db.ProcessingErrorResolution(body))
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
		context.AbortWithStatus(http.StatusNotFound)
		return
	}
	context.JSON(http.StatusOK, process)
}

func getMyProcesses(context *gin.Context) {
	userID := context.MustGet("userId").(string)
	processes := db.GetProcessesForUser(userID)
	context.JSON(http.StatusOK, processes)
}

func getProcesses(context *gin.Context) {
	processes := db.GetProcesses()
	context.JSON(http.StatusOK, processes)
}

func deleteProcess(context *gin.Context) {
	id := context.Param("id")
	if found := xdomea.DeleteProcess(id); found {
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
	message, found := db.GetMessageByID(id)
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
		return
	}
	context.Header("Content-Type", "application/json; charset=utf-8")
	context.String(http.StatusOK, message.MessageJSON)
}

func getFileRecordObjectByID(context *gin.Context) {
	id, err := uuid.Parse(context.Param("id"))
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	fileRecordObject, found := db.GetFileRecordObjectByID(id)
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
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
		context.AbortWithStatus(http.StatusNotFound)
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
	documentRecordObject, found := db.GetDocumentRecordObjectByID(id)
	if !found {
		context.AbortWithStatus(http.StatusNotFound)
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

func getAppraisalCodelist(context *gin.Context) {
	appraisals := db.GetAppraisalCodelist()
	context.JSON(http.StatusOK, appraisals)
}

func getConfidentialityLevelCodelist(context *gin.Context) {
	codelist := db.GetConfidentialityLevelCodelist()
	context.JSON(http.StatusOK, codelist)
}

func getAppraisals(context *gin.Context) {
	processID := context.Param("processId")
	appraisals := db.GetAppraisalsForProcess(processID)
	context.JSON(http.StatusOK, appraisals)
}

func setAppraisalDecision(context *gin.Context) {
	processID := context.Query("processId")
	if processID == "" {
		context.AbortWithStatus(http.StatusBadRequest)
		return
	}
	db.GetProcess(processID)
	recordObjectID := context.Query("recordObjectId")
	appraisalDecision, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	err = xdomea.SetAppraisalDecisionRecursive(processID,
		recordObjectID,
		db.AppraisalDecisionOption((appraisalDecision)))
	if err != nil {
		context.AbortWithError(http.StatusBadRequest, err)
		return
	}
	appraisals := db.GetAppraisalsForProcess(processID)
	context.JSON(http.StatusAccepted, appraisals)
}

func setAppraisalNote(context *gin.Context) {
	processID := context.Query("processId")
	if processID == "" {
		context.String(http.StatusBadRequest, "missing query parameter processId")
		return
	}
	db.GetProcess(processID)
	recordObjectID := context.Query("recordObjectId")
	appraisalInternalNote, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	err = xdomea.SetAppraisalInternalNote(processID,
		recordObjectID,
		string(appraisalInternalNote))
	if err != nil {
		context.AbortWithError(http.StatusBadRequest, err)
		return
	}
	appraisals := db.GetAppraisalsForProcess(processID)
	context.JSON(http.StatusAccepted, appraisals)
}

type MultiAppraisalBody struct {
	ProcessID       string                     `json:"processId"`
	RecordObjectIDs []string                   `json:"recordObjectIds"`
	Decision        db.AppraisalDecisionOption `json:"decision"`
	InternalNote    string                     `json:"internalNote"`
}

func setAppraisals(context *gin.Context) {
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
	err = xdomea.SetAppraisals(
		parsedBody.ProcessID,
		parsedBody.RecordObjectIDs,
		parsedBody.Decision,
		parsedBody.InternalNote,
	)
	if err != nil {
		context.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to set appraisals: %v", err))
		return
	}
	appraisals := db.GetAppraisalsForProcess(parsedBody.ProcessID)
	context.JSON(http.StatusAccepted, appraisals)
}

func finalizeMessageAppraisal(context *gin.Context) {
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
	process := db.GetProcessForMessage(message)
	if process.ProcessState.Appraisal.Complete {
		context.AbortWithStatus(http.StatusConflict)
		return
	}
	userID := context.MustGet("userId").(string)
	userName := auth.GetDisplayName(userID)
	message = xdomea.FinalizeMessageAppraisal(message, userName)
	messagePath := xdomea.Send0502Message(process.Agency, message)
	db.UpdateProcess(process.ID, db.Process{
		Message0502Path: &messagePath,
	})
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
	contentLength, contentType, body := report.GetReport(process)
	context.DataFromReader(http.StatusOK, contentLength, contentType, body, nil)
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
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	var collection db.Collection
	if archiveTarget == "dimag" {
		collectionIDString, hasCollectionID := context.GetQuery("collectionId")
		if !hasCollectionID || collectionIDString == "0" {
			context.String(http.StatusBadRequest, "missing query parameter \"collectionId\"")
			return
		}
		collectionID, err := strconv.ParseUint(collectionIDString, 10, 0)
		if err != nil {
			context.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}
		collection, found = db.GetCollection(uint(collectionID))
		if !found {
			context.String(http.StatusNotFound, fmt.Sprintf("collection not found: %d", collectionID))
			return
		}
	}
	userID := context.MustGet("userId").(string)
	userName := auth.GetDisplayName(userID)
	task := tasks.Start(db.TaskTypeArchiving, process, 0)
	go func() {
		switch archiveTarget {
		case "filesystem":
			err = filesystem.ArchiveMessage(process, message)
		case "dimag":
			err = dimag.ImportMessageSync(process, message, collection)
		default:
			panic("unknown archive target: " + archiveTarget)
		}
		if err != nil {
			processingError := tasks.MarkFailed(&task, err.Error())
			xdomea.HandleError(processingError)
		} else {
			tasks.MarkDone(&task, &userName)
			preferences := db.GetUserInformation(userID).Preferences
			if preferences.ReportByEmail {
				process, _ = db.GetProcess(message.MessageHead.ProcessID)
				_, contentType, reader := report.GetReport(process)
				body, err := io.ReadAll(reader)
				if err != nil {
					panic(err)
				}
				address := auth.GetMailAddress(userID)
				filename := fmt.Sprintf("Übernahmebericht %s %s.pdf", process.Agency.Abbreviation, process.CreatedAt)
				mail.SendMailReport(
					address, process,
					mail.Attachment{Filename: filename, ContentType: contentType, Body: body},
				)
			}
		}
	}()
}

func Users(c *gin.Context) {
	users := auth.ListUsers()
	c.JSON(http.StatusOK, users)
}

func getUserInformation(context *gin.Context) {
	userIDString, hasUserID := context.GetQuery("userId")
	if !hasUserID {
		context.AbortWithStatus(http.StatusBadRequest)
	}
	userInfo := db.GetUserInformation(userIDString)
	context.JSON(http.StatusOK, userInfo)
}

func getMyUserInformation(context *gin.Context) {
	userID := context.MustGet("userId").(string)
	userInfo := db.GetUserInformation(userID)
	context.JSON(http.StatusOK, userInfo)
}

func setUserPreferences(context *gin.Context) {
	userID := context.MustGet("userId").(string)
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		panic(err)
	}
	var userPreferences db.UserPreferences
	err = json.Unmarshal(body, &userPreferences)
	if err != nil {
		context.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	db.SaveUserPreferences(userID, userPreferences)
}

func getAgencies(context *gin.Context) {
	var agencies []db.Agency
	if collectionIDString, hasCollectionID := context.GetQuery("collectionId"); hasCollectionID {
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
	success := xdomea.TestTransferDir(string(body))
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

func getCollectionDimagIDs(context *gin.Context) {
	ids := dimag.GetCollectionIDs()
	context.JSON(http.StatusOK, ids)
}
