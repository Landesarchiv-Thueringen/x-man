package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lath/xman/internal/archive"
	"lath/xman/internal/archive/dimag"
	"lath/xman/internal/auth"
	"lath/xman/internal/core"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"lath/xman/internal/mail"
	"lath/xman/internal/report"
	"lath/xman/internal/routines"
	"lath/xman/internal/tasks"
	"lath/xman/internal/verification"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var defaultResponse = fmt.Sprintf("x-man server %s is running", core.XMAN_VERSION)

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
	authorized.GET("api/primary-documents-info/:processId", getPrimaryDocumentsInfo)
	authorized.GET("api/primary-document-data/:processId/:filename", getPrimaryDocumentData)
	authorized.GET("api/report/appraisal/:processId", getAppraisalReport)
	authorized.GET("api/report/submission/:processId", getSubmissionReport)
	authorized.GET("api/archive-collections", getCollections)
	authorized.GET("api/user-info", getUserInformation)
	authorized.POST("api/user-preferences", setUserPreferences)
	authorized.GET("api/appraisals/:processId", getAppraisals)
	authorized.POST("api/appraisal-decision", setAppraisalDecision)
	authorized.POST("api/appraisal-note", setAppraisalNote)
	authorized.POST("api/appraisals", setAppraisals)
	authorized.PATCH("api/finalize-message-appraisal/:processId", finalizeMessageAppraisal)
	authorized.GET("api/packaging/:processId", getPackaging)
	authorized.POST("api/packaging", setPackagingChoice)
	authorized.POST("api/packaging-stats/:processId", getPackagingStatsForOptions)
	authorized.PATCH("api/archive-0503-message/:processId", archive0503Message)
	authorized.PATCH("api/process-note/:processId", setProcessNote)
	authorized.GET("api/task/:id", getTask)
	admin := router.Group("/")
	admin.Use(auth.AdminRequired())
	admin.GET("api/processes", getProcesses)
	admin.DELETE("api/process/:processId", deleteProcess)
	admin.POST("api/message/:processId/:messageType/reimport", reimportMessage)
	admin.DELETE("api/message/:processId/:messageType", deleteMessage)
	admin.GET("api/processing-errors", getProcessingErrors)
	admin.POST("api/processing-errors/resolve/:id", resolveProcessingError)
	admin.GET("api/admin-config", getAdminConfig)
	admin.GET("api/users", users)
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
	router.Run()
}

func initServer() {
	log.Println(defaultResponse)
	db.Init()
	auth.Init()
	migrateData()
	testConfiguration()
	go core.MonitorTransferDirs()
}

func migrateData() {
	s, _ := db.FindServerStateXman()
	if s.Version == "" {
		if os.Getenv("INIT_TEST_SETUP") == "true" {
			log.Println("Initializing database with test data...")
			core.InitTestSetup()
			log.Println("done")
		}
	} else {
		log.Printf("Database is up do date with X-Man version %s\n", core.XMAN_VERSION)
	}
	db.UpsertServerStateXmanVersion(core.XMAN_VERSION)
}

func testConfiguration() {
	log.Println("Testing connection to LDAP server...")
	auth.TestConnection()
	log.Println("Connection to LDAP server successful")
	if os.Getenv("SMTP_SERVER") != "" {
		log.Println("Testing connection to SMTP server...")
		err := mail.TestConnection()
		if err != nil {
			log.Fatal("Failed to connect to SMTP server: ", err)
		}
		log.Println("Connection to SMTP server successful")
	}
	if os.Getenv("BORG_URL") != "" {
		log.Println("Testing connection to BORG...")
		err := verification.TestConnection()
		if err != nil {
			log.Fatal("Failed to connect to BORG: ", err)
		}
		log.Println("Connection to BORG successful")
	}
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	if archiveTarget == "dimag" {
		log.Println("Testing connection to DIMAG...")
		err := dimag.TestConnection()
		if err != nil {
			log.Fatal("Failed to connect to DIMAG: ", err)
		}
		log.Println("Connection to DIMAG successful")
	}
}

func getDefaultResponse(c *gin.Context) {
	c.String(http.StatusOK, defaultResponse)
}

func getAbout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": core.XMAN_VERSION,
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
	userID := c.MustGet("userId").(string)
	userName := auth.GetDisplayName(userID)
	core.Resolve(processingError, db.ProcessingErrorResolution(body), userName)
	c.Status(http.StatusAccepted)
}

func getProcessData(c *gin.Context) {
	processID := c.Param("processId")
	process, found := db.FindProcess(c.Request.Context(), processID)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	warnings := db.FindWarningsForProcess(c.Request.Context(), processID)
	if warnings == nil {
		warnings = make([]db.Warning, 0)
	}
	processingErrors := db.FindProcessingErrorsForProcess(c.Request.Context(), processID)
	if processingErrors == nil {
		processingErrors = make([]db.ProcessingError, 0)
	}
	c.JSON(http.StatusOK, gin.H{
		"process":          process,
		"warnings":         warnings,
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
	processID := c.Param("processId")
	if found := core.DeleteProcess(processID); found {
		c.Status(http.StatusAccepted)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func reimportMessage(c *gin.Context) {
	processID := c.Param("processId")
	messageType := c.Param("messageType")
	err := core.DeleteMessage(processID, db.MessageType(messageType), true)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

func deleteMessage(c *gin.Context) {
	processID := c.Param("processId")
	messageType := c.Param("messageType")
	err := core.DeleteMessage(processID, db.MessageType(messageType), false)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
}

func getMessage(c *gin.Context) {
	processID := c.Param("processId")
	messageType := c.Param("messageType")
	message, found := db.FindMessage(c.Request.Context(), processID, db.MessageType(messageType))
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, message)
}

func getRootRecords(c *gin.Context) {
	processID := c.Param("processId")
	messageType := c.Param("messageType")
	rootRecords := db.FindAllRootRecords(c.Request.Context(), processID, db.MessageType(messageType))
	c.JSON(http.StatusOK, rootRecords)
}

func getAppraisals(c *gin.Context) {
	processID := c.Param("processId")
	appraisals := db.FindAppraisalsForProcess(c.Request.Context(), processID)
	c.JSON(http.StatusOK, appraisals)
}

func setAppraisalDecision(c *gin.Context) {
	processID := c.Query("processId")
	recordID := c.Query("recordId")
	appraisalDecision, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	err = core.SetAppraisalDecisionRecursive(processID,
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
	processID := c.Query("processId")
	recordID := c.Query("recordId")
	appraisalNote, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}
	err = core.SetAppraisalInternalNote(processID, recordID, string(appraisalNote))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	appraisals := db.FindAppraisalsForProcess(c.Request.Context(), processID)
	c.JSON(http.StatusAccepted, appraisals)
}

type MultiAppraisalBody struct {
	ProcessID       string                     `json:"processId"`
	RecordObjectIDs []string                   `json:"recordObjectIds"`
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
	err = core.SetAppraisals(
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

func finalizeMessageAppraisal(c *gin.Context) {
	processID := c.Param("processId")
	message, found := db.FindMessage(c, processID, db.MessageType0501)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	process, found := db.FindProcess(context.Background(), message.MessageHead.ProcessID)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if process.ProcessState.Appraisal.Complete {
		c.AbortWithStatus(http.StatusConflict)
		return
	}
	userID := c.MustGet("userId").(string)
	userName := auth.GetDisplayName(userID)
	message = core.FinalizeMessageAppraisal(message, userName)
	err := core.Send0502Message(process.Agency, message)
	if err != nil {
		errorData := db.ProcessingError{
			Title:     "Fehler beim Senden der 0502-Nachricht",
			ProcessID: &processID,
		}
		errors.AddProcessingErrorWithData(err, errorData)
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		preferences := db.FindUserPreferencesWithDefault(context.Background(), userID)
		if preferences.ReportByEmail {
			defer errors.HandlePanic("generate report for e-mail", &db.ProcessingError{
				ProcessID: &message.MessageHead.ProcessID,
			})
			process, ok := db.FindProcess(context.Background(), message.MessageHead.ProcessID)
			if !ok {
				panic("failed to find process:" + process.ProcessID)
			}
			_, contentType, reader := report.GetAppraisalReport(context.Background(), process)
			body, err := io.ReadAll(reader)
			if err != nil {
				panic(err)
			}
			errorData := db.ProcessingError{
				Title:     "Fehler beim Versenden einer E-Mail-Benachrichtigung",
				ProcessID: &message.MessageHead.ProcessID,
			}
			address, err := auth.GetMailAddress(userID)
			if err != nil {
				errors.AddProcessingErrorWithData(err, errorData)
			} else {
				filename := fmt.Sprintf(
					"Bewertungsbericht %s %s.pdf",
					process.Agency.Abbreviation, process.CreatedAt,
				)
				err = mail.SendMailAppraisalReport(
					address, process,
					mail.Attachment{Filename: filename, ContentType: contentType, Body: body},
				)
				if err != nil {
					errors.AddProcessingErrorWithData(err, errorData)
				}
			}
		}
	}
}

func areAllRecordObjectsAppraised(c *gin.Context) {
	processID := c.Param("processId")
	appraisalComplete := core.AreAllRecordObjectsAppraised(c.Request.Context(), processID)
	c.JSON(http.StatusOK, appraisalComplete)
}

func getPackaging(c *gin.Context) {
	processID := c.Param("processId")
	decisions, stats, choices := core.Packaging(processID)
	c.JSON(http.StatusOK, gin.H{
		"decisions": decisions,
		"stats":     stats,
		"choices":   choices,
	})
}

// getPackagingStatsForOptions returns a map with packaging stats for each
// available packaging option, if applied to all given root records.
//
// The given root record have to be file records.
//
// The methods is invoked via a POST request to be able to retrieve a
// potentially long list of record IDs as request body.
func getPackagingStatsForOptions(c *gin.Context) {
	processID := c.Param("processId")
	jsonBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	var recordIDs []string
	err = json.Unmarshal(jsonBody, &recordIDs)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	rootRecords := db.FindRootRecords(c.Request.Context(), processID, db.MessageType0503, recordIDs)
	statsMap := core.PackagingStatsForChoices(rootRecords.Files)
	c.JSON(http.StatusOK, statsMap)
}

func setPackagingChoice(c *gin.Context) {
	jsonBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	var data struct {
		ProcessID string             `json:"processId"`
		RecordIDs []string           `json:"recordIds"`
		Packaging db.PackagingChoice `json:"packagingChoice"`
	}
	err = json.Unmarshal(jsonBody, &data)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	for _, id := range data.RecordIDs {
		db.UpsertPackagingChoice(data.ProcessID, id, data.Packaging)
	}
	decisions, stats, choices := core.Packaging(data.ProcessID)
	c.JSON(http.StatusOK, gin.H{
		"decisions": decisions,
		"stats":     stats,
		"choices":   choices,
	})
}

func setProcessNote(c *gin.Context) {
	processID := c.Param("processId")
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
	processID := c.Query("processID")
	filename := c.Query("filename")
	message, ok := db.FindMessage(c.Request.Context(), processID, db.MessageType0503)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	path := filepath.Join(message.StoreDir, filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.FileAttachment(path, filename)
}

func getPrimaryDocumentsInfo(c *gin.Context) {
	processID := c.Param("processId")
	primaryDocuments := db.FindPrimaryDocumentsDataForProcess(c.Request.Context(), processID)
	data := make([]gin.H, len(primaryDocuments))
	for i, d := range primaryDocuments {
		data[i] = gin.H{
			"filename":         d.Filename,
			"filenameOriginal": d.FilenameOriginal,
			"recordId":         d.RecordID,
		}
		if d.FormatVerification != nil {
			data[i]["formatVerificationSummary"] = d.FormatVerification.Summary
		}
	}
	c.JSON(http.StatusOK, data)
}

func getPrimaryDocumentData(c *gin.Context) {
	processID := c.Param("processId")
	filename := c.Param("filename")
	data, found := db.FindPrimaryDocumentData(c.Request.Context(), processID, filename)
	if !found {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, data)
}

func getAppraisalReport(c *gin.Context) {
	processID := c.Param("processId")
	process, found := db.FindProcess(c.Request.Context(), processID)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	contentLength, contentType, body := report.GetAppraisalReport(c.Request.Context(), process)
	c.DataFromReader(http.StatusOK, contentLength, contentType, body, nil)
}

func getSubmissionReport(c *gin.Context) {
	processID := c.Param("processId")
	process, found := db.FindProcess(c.Request.Context(), processID)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	contentLength, contentType, body := report.GetSubmissionReport(c.Request.Context(), process)
	c.DataFromReader(http.StatusOK, contentLength, contentType, body, nil)
}

// archive0503Message archives all metadata and primary files in the digital archive.
func archive0503Message(c *gin.Context) {
	processID := c.Param("processId")
	process, found := db.FindProcess(context.Background(), processID)
	if !found {
		c.AbortWithStatus(http.StatusNotFound)
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

func getAdminConfig(c *gin.Context) {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpTlsMode := os.Getenv("SMTP_TLS_MODE")

	c.JSON(http.StatusOK, gin.H{
		"smtpServer":  smtpServer,
		"smtpTlsMode": smtpTlsMode,
	})
}

func users(c *gin.Context) {
	users := auth.ListUsers()
	c.JSON(http.StatusOK, users)
}

func getUserInformation(c *gin.Context) {
	userID := c.MustGet("userId").(string)
	agencies := db.FindAgenciesForUser(c.Request.Context(), userID)
	preferences := db.FindUserPreferencesWithDefault(c.Request.Context(), userID)
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
	if agencies == nil {
		agencies = make([]db.Agency, 0)
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
	c.JSON(http.StatusAccepted, gin.H{"id": id.Hex()})
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
	success := core.TestTransferDir(string(body))
	if success {
		c.JSON(http.StatusOK, gin.H{"result": "success"})
	} else {
		c.JSON(http.StatusOK, gin.H{"result": "failed"})
	}
}

func getTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
		return
	}
	task, ok := db.FindTask(c.Request.Context(), id)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, task)
}

func getTasks(c *gin.Context) {
	tasks := db.FindTasksMetadata(c.Request.Context())
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
	err = tasks.Action(id, action)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusAccepted)
}

func getCollectionDimagIDs(c *gin.Context) {
	ids, err := dimag.GetCollectionIDs()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, ids)
}

func handleRecovery(c *gin.Context, err any) {
	if os.Getenv("DEBUG_MODE") == "true" {
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
	if processID := c.Param("processId"); processID != "" {
		e.ProcessID = &processID
	}
	if messageType := c.Param("messageType"); messageType != "" {
		e.MessageType = db.MessageType(messageType)
	}
	if userID := c.GetString("userId"); userID != "" {
		e.Info += "\n\nNutzer: " + auth.GetDisplayName(userID)
	}
	errors.AddProcessingError(e)
}
