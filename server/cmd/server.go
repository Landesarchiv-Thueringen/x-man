package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lath/xman/internal/archive"
	"lath/xman/internal/archive/dimag"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"lath/xman/internal/report"
	"lath/xman/internal/routines"
	"lath/xman/internal/tasks"
	"lath/xman/internal/xdomea"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const XMAN_MAJOR_VERSION = 0
const XMAN_MINOR_VERSION = 10
const XMAN_PATCH_VERSION = 0

var XMAN_VERSION = fmt.Sprintf("%d.%d.%d", XMAN_MAJOR_VERSION, XMAN_MINOR_VERSION, XMAN_PATCH_VERSION)

var defaultResponse = fmt.Sprintf("x-man server %s is running", XMAN_VERSION)

func main() {
	initServer()
	routines.Init()
	tasks.ResumeAfterAppRestart()
	router := gin.New()
	router.Use(gin.Logger(), gin.CustomRecovery(handleRecovery))
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{})
	router.GET("api", getDefaultResponse)
	router.GET("api/about", getAbout)
	router.GET("api/login", auth.Login)
	router.GET("api/updates", auth.AuthRequiredQueryParam(), getUpdates)
	authorized := router.Group("/")
	authorized.Use(auth.AuthRequired())
	authorized.GET("api/config", getConfig)
	authorized.GET("api/processes/my", getMyProcesses)
	authorized.GET("api/process/:processId", getProcessData)
	authorized.GET("api/message/:processId/:messageType", getMessage)
	authorized.GET("api/root-records/:processId/:messageType", getRootRecords)
	authorized.GET("api/all-record-objects-appraised/:processId", areAllRecordObjectsAppraised)
	authorized.GET("api/primary-document", getPrimaryDocument)
	authorized.GET("api/primary-documents-data/:processId", getPrimaryDocumentsData)
	authorized.GET("api/report/:processId", getReport)
	authorized.GET("api/archive-collections", getCollections)
	authorized.GET("api/user-info", getUserInformation)
	authorized.POST("api/user-preferences", setUserPreferences)
	authorized.GET("api/appraisals/:processId", getAppraisals)
	authorized.POST("api/appraisal-decision", setAppraisalDecision)
	authorized.POST("api/appraisal-note", setAppraisalNote)
	authorized.POST("api/appraisals", setAppraisals)
	authorized.PATCH("api/finalize-message-appraisal/:processId", finalizeMessageAppraisal)
	authorized.PATCH("api/archive-0503-message/:processId", archive0503Message)
	authorized.PATCH("api/process-note/:processId", setProcessNote)
	admin := router.Group("/")
	admin.Use(auth.AdminRequired())
	admin.GET("api/processes", getProcesses)
	admin.DELETE("api/process/:processId", deleteProcess)
	admin.GET("api/processing-errors", getProcessingErrors)
	admin.POST("api/processing-errors/resolve/:id", resolveProcessingError)
	admin.GET("api/users", Users)
	admin.GET("api/agencies", getAgencies)
	admin.PUT("api/agency", putAgency)
	admin.POST("api/agency", postAgency)
	admin.DELETE("api/agency/:id", deleteAgency)
	admin.PUT("api/archive-collection", putCollection)
	admin.POST("api/archive-collection", postCollection)
	admin.DELETE("api/archive-collection/:id", deleteCollection)
	admin.POST("api/test-transfer-dir", testTransferDir)
	admin.GET("api/tasks", getTasks)
	admin.POST("api/task/action/:id", taskAction)
	admin.GET("api/dimag-collection-ids", getCollectionDimagIDs)
	addr := "0.0.0.0:80"
	router.Run(addr)
}

func initServer() {
	log.Println(defaultResponse)
	db.Init()
	MigrateData()
	go xdomea.MonitorTransferDirs()
}

func MigrateData() {
	_, ok := db.FindServerStateXman()
	if !ok {
		if os.Getenv("INIT_TEST_SETUP") == "true" {
			log.Println("Initializing database with test data...")
			xdomea.InitTestSetup()
			log.Println("done")
		}
	} else {
		log.Printf("Database is up do date with X-Man version %s\n", XMAN_VERSION)
	}
	db.UpsertServerStateXmanVersion(XMAN_VERSION)
}

func getDefaultResponse(c *gin.Context) {
	c.String(http.StatusOK, defaultResponse)
}

func getAbout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": XMAN_VERSION,
	})
}

func getUpdates(c *gin.Context) {
	ch := db.RegisterUpdatesChannel()
	defer db.UnregisterUpdatesChannel(ch)
	// This connection should be kept open while a client is connected, i.e.,
	// the app is open in a browser. However, we might miss disconnects when not
	// properly propagated, e.g., by a misconfigured proxy. We add a generous
	// timeout to eventually unregister the channel in these cases. When in fact
	// still connected, the client will reconnect after being disconnected by
	// us.
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Hour*1)
	defer cancel()
	heartbeat := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				heartbeat <- struct{}{}
				time.Sleep(time.Second * 30)
			}
		}
	}()
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case <-heartbeat:
			c.SSEvent("heartbeat", struct{}{})
			return true
		case msg := <-ch:
			c.SSEvent("message", msg)
			return true
		}
	})
}

func getConfig(c *gin.Context) {
	deleteArchivedProcessesAfterDays, err := strconv.Atoi(os.Getenv("DELETE_ARCHIVED_SUBMISSIONS_AFTER_DAYS"))
	if err != nil {
		log.Fatal("failed to read environment variable: DELETE_ARCHIVED_SUBMISSIONS_AFTER_DAYS")
	}
	appraisalLevel := os.Getenv("APPRAISAL_LEVEL")
	if appraisalLevel == "" {
		log.Fatal("missing environment variable: APPRAISAL_LEVEL")
	}
	supportsEmailNotifications := os.Getenv("SMTP_SERVER") != ""
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	if archiveTarget == "" {
		log.Fatal("missing environment variable: ARCHIVE_TARGET")
	}
	borgSupport := os.Getenv("BORG_URL") != ""
	c.JSON(http.StatusOK, gin.H{
		"deleteArchivedProcessesAfterDays": deleteArchivedProcessesAfterDays,
		"appraisalLevel":                   appraisalLevel,
		"supportsEmailNotifications":       supportsEmailNotifications,
		"archiveTarget":                    archiveTarget,
		"borgSupport":                      borgSupport,
	})
}

func getProcessingErrors(c *gin.Context) {
	processingErrors := db.FindProcessingErrors(c)
	c.JSON(http.StatusOK, processingErrors)
}

func resolveProcessingError(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	processingError, found := db.FindProcessingError(c.Request.Context(), id)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	xdomea.Resolve(processingError, db.ProcessingErrorResolution(body))
	c.Status(http.StatusAccepted)
}

func getProcessData(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	process, found := db.FindProcess(c.Request.Context(), processID)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	processingErrors := db.FindProcessingErrorsForProcess(c.Request.Context(), processID)
	if processingErrors == nil {
		processingErrors = make([]db.ProcessingError, 0)
	}
	c.JSON(http.StatusOK, gin.H{
		"process":          process,
		"processingErrors": processingErrors,
	})
}

func getMyProcesses(c *gin.Context) {
	userID := c.MustGet("userId").(string)
	processes := db.FindProcessesForUser(c.Request.Context(), userID)
	c.JSON(http.StatusOK, processes)
}

func getProcesses(c *gin.Context) {
	processes := db.FindProcesses(c)
	c.JSON(http.StatusOK, processes)
}

func deleteProcess(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	if found := xdomea.DeleteProcess(processID); found {
		c.Status(http.StatusAccepted)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func getMessage(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	messageType := c.Param("messageType")
	message, found := db.FindMessage(c.Request.Context(), processID, db.MessageType(messageType))
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, message)
}

func getRootRecords(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	messageType := c.Param("messageType")
	rootRecords := db.FindRootRecords(c.Request.Context(), processID, db.MessageType(messageType))
	c.JSON(http.StatusOK, rootRecords)
}

func getAppraisals(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	appraisals := db.FindAppraisalsForProcess(c.Request.Context(), processID)
	c.JSON(http.StatusOK, appraisals)
}

func setAppraisalDecision(c *gin.Context) {
	processID, err := uuid.Parse(c.Query("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	recordID, err := uuid.Parse(c.Query("recordId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	appraisalDecision, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	err = xdomea.SetAppraisalDecisionRecursive(processID,
		recordID,
		db.AppraisalDecisionOption((appraisalDecision)))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	appraisals := db.FindAppraisalsForProcess(c.Request.Context(), processID)
	c.JSON(http.StatusAccepted, appraisals)
}

func setAppraisalNote(c *gin.Context) {
	processID, err := uuid.Parse(c.Query("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	recordID, err := uuid.Parse(c.Query("recordId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	appraisalNote, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	err = xdomea.SetAppraisalInternalNote(processID, recordID, string(appraisalNote))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	appraisals := db.FindAppraisalsForProcess(c.Request.Context(), processID)
	c.JSON(http.StatusAccepted, appraisals)
}

type MultiAppraisalBody struct {
	ProcessID       uuid.UUID                  `json:"processId"`
	RecordObjectIDs []uuid.UUID                `json:"recordObjectIds"`
	Decision        db.AppraisalDecisionOption `json:"decision"`
	InternalNote    string                     `json:"internalNote"`
}

func setAppraisals(c *gin.Context) {
	jsonBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	var parsedBody MultiAppraisalBody
	err = json.Unmarshal(jsonBody, &parsedBody)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	err = xdomea.SetAppraisals(
		parsedBody.ProcessID,
		parsedBody.RecordObjectIDs,
		parsedBody.Decision,
		parsedBody.InternalNote,
	)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("failed to set appraisals: %v", err))
		return
	}
	appraisals := db.FindAppraisalsForProcess(c.Request.Context(), parsedBody.ProcessID)
	c.JSON(http.StatusAccepted, appraisals)
}

func finalizeMessageAppraisal(ctx *gin.Context) {
	processID, err := uuid.Parse(ctx.Param("processId"))
	if err != nil {
		ctx.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	message, found := db.FindMessage(ctx, processID, db.MessageType0501)
	if !found {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	process, found := db.FindProcess(context.Background(), message.MessageHead.ProcessID)
	if !found {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	if process.ProcessState.Appraisal.Complete {
		ctx.AbortWithStatus(http.StatusConflict)
		return
	}
	userID := ctx.MustGet("userId").(string)
	userName := auth.GetDisplayName(userID)
	message = xdomea.FinalizeMessageAppraisal(message, userName)
	messagePath := xdomea.Send0502Message(process.Agency, message)
	db.MustUpdateProcessMessagePath(process.ProcessID, db.MessageType0502, messagePath)
}

func areAllRecordObjectsAppraised(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	appraisalComplete := xdomea.AreAllRecordObjectsAppraised(c.Request.Context(), processID)
	c.JSON(http.StatusOK, appraisalComplete)
}

func setProcessNote(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	note, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	ok := db.UpdateProcessNote(processID, string(note))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func getPrimaryDocument(c *gin.Context) {
	processID, err := uuid.Parse(c.Query("processID"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	filename := c.Query("filename")
	message, ok := db.FindMessage(c.Request.Context(), processID, db.MessageType0503)
	if !ok {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	path := filepath.Join(message.StoreDir, filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.FileAttachment(path, filename)
}

func getPrimaryDocumentsData(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	primaryDocuments := db.FindPrimaryDocumentsDataForProcess(c.Request.Context(), processID)
	c.JSON(http.StatusOK, primaryDocuments)
}

func getReport(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	process, found := db.FindProcess(c.Request.Context(), processID)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	contentLength, contentType, body := report.GetReport(c.Request.Context(), process)
	c.DataFromReader(http.StatusOK, contentLength, contentType, body, nil)
}

// archive0503Message archives all metadata and primary files in the digital archive.
func archive0503Message(c *gin.Context) {
	processID, err := uuid.Parse(c.Param("processId"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	process, found := db.FindProcess(context.Background(), processID)
	if !found {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	var isArchivable bool
	state := process.ProcessState
	if os.Getenv("BORG_URL") != "" {
		isArchivable = state.FormatVerification.Complete && !state.Archiving.Complete
	} else {
		isArchivable = state.Receive0503.Complete && !state.Archiving.Complete
	}
	if !isArchivable {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("message can't be archived"))
		return
	}
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	var collection db.ArchiveCollection
	if archiveTarget == "dimag" {
		collectionIDString := c.Query("collectionId")
		if collectionIDString == "" {
			c.String(http.StatusBadRequest, "missing query parameter \"collectionId\"")
			return
		}
		collectionID, err := primitive.ObjectIDFromHex(collectionIDString)
		if err != nil {
			c.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}
		collection, found = db.FindArchiveCollection(context.Background(), collectionID)
		if !found {
			c.String(http.StatusNotFound, fmt.Sprintf("collection not found: %v", collectionID))
			return
		}
	}
	userID := c.MustGet("userId").(string)
	archive.ArchiveSubmission(process, collection, userID)
}

func Users(c *gin.Context) {
	users := auth.ListUsers()
	c.JSON(http.StatusOK, users)
}

func getUserInformation(c *gin.Context) {
	userID := c.MustGet("userId").(string)
	agencies := db.FindAgenciesForUser(c.Request.Context(), userID)
	preferences := db.TryFindUserPreferences(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{
		"agencies":    agencies,
		"preferences": preferences,
	})
}

func setUserPreferences(c *gin.Context) {
	userID := c.MustGet("userId").(string)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	var userPreferences db.UserPreferences
	err = json.Unmarshal(body, &userPreferences)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	userPreferences.UserID = userID
	db.UpsertUserPreferences(userPreferences)
}

func getAgencies(c *gin.Context) {
	var agencies []db.Agency
	if collectionIDString, hasCollectionID := c.GetQuery("collectionId"); hasCollectionID {
		collectionID, err := primitive.ObjectIDFromHex(collectionIDString)
		if err != nil {
			c.AbortWithError(http.StatusUnprocessableEntity, err)
			return
		}
		agencies = db.FindAgenciesForCollection(c.Request.Context(), collectionID)
	} else {
		agencies = db.FindAgencies(c)
	}
	c.JSON(http.StatusOK, agencies)
}

func putAgency(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	var agency db.Agency
	err = json.Unmarshal(body, &agency)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	id := db.InsertAgency(agency)
	c.String(http.StatusAccepted, id.String())
}

func postAgency(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	var agency db.Agency
	err = json.Unmarshal(body, &agency)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	ok := db.ReplaceAgency(agency)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func deleteAgency(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	ok := db.DeleteAgency(id)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func getCollections(c *gin.Context) {
	Collections := db.FindArchiveCollections(c)
	c.JSON(http.StatusOK, Collections)
}

func putCollection(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	var Collection db.ArchiveCollection
	err = json.Unmarshal(body, &Collection)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	id := db.InsertArchiveCollection(Collection)
	c.JSON(http.StatusAccepted, gin.H{
		"id": id.Hex(),
	})
}

func postCollection(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	var collection db.ArchiveCollection
	err = json.Unmarshal(body, &collection)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	ok := db.ReplaceArchiveCollection(collection)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func deleteCollection(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
	}
	found := db.DeleteArchiveCollection(id)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Status(http.StatusAccepted)
}

func testTransferDir(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	success := xdomea.TestTransferDir(string(body))
	if success {
		c.JSON(http.StatusOK, gin.H{"result": "success"})
	} else {
		c.JSON(http.StatusOK, gin.H{"result": "failed"})
	}
}

func getTasks(c *gin.Context) {
	tasks := db.FindTasks(c)
	c.JSON(http.StatusOK, tasks)
}

func taskAction(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	action := db.TaskAction(body)
	if action != db.TaskActionPause && action != db.TaskActionRetry && action != db.TaskActionRun {
		c.Status(http.StatusBadRequest)
		return
	}
	err = tasks.Action(id, action)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusAccepted)
}

func getCollectionDimagIDs(c *gin.Context) {
	ids := dimag.GetCollectionIDs()
	c.JSON(http.StatusOK, ids)
}

func handleRecovery(c *gin.Context, err any) {
	if os.Getenv("DEBUG_MODE") == "debug" {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError,
			gin.H{"error": err},
		)
	} else {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	e := db.ProcessingError{
		Title:     "Anwendungsfehler",
		ErrorType: "application-error",
		Info:      fmt.Sprintf("%s %s\n\n%v", c.Request.Method, c.Request.URL, err),
	}
	if processID, err := uuid.Parse(c.Param("processId")); err == nil {
		e.ProcessID = processID
	}
	if messageType := c.Param("messageType"); messageType != "" {
		e.MessageType = db.MessageType(messageType)
	}
	if userID := c.GetString("userId"); userID != "" {
		e.Info += "\n\nNutzer: " + auth.GetDisplayName(userID)
	}
	errors.AddProcessingError(e)
}
