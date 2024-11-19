package archive

import (
	"context"
	"fmt"
	"io"
	"lath/xman/internal/archive/dimag"
	"lath/xman/internal/archive/filesystem"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"lath/xman/internal/mail"
	"lath/xman/internal/report"
	"lath/xman/internal/tasks"
	"lath/xman/internal/xdomea"
	"os"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	tasks.RegisterTaskHandler(
		db.ProcessStepArchiving,
		initArchiveHandler,
		tasks.Options{
			ConcurrentTasks: 1,
			ConcurrentItems: 1,
			SafeRepeat:      false,
		},
	)
}

// ArchiveSubmission creates a new task for archiving the submission process and
// starts it.
func ArchiveSubmission(
	process db.SubmissionProcess,
	collection db.ArchiveCollection,
	userID string,
) {
	rootRecords := db.FindAllRootRecords(
		context.Background(), process.ProcessID, db.MessageType0503,
	)
	var items []db.TaskItem
	m, _, _ := xdomea.Packaging(process.ProcessID)
	items = append(items, taskItemsForFiles(m, []uuid.UUID{}, rootRecords.Files)...)
	items = append(items, taskItemsForProcesses(m, []uuid.UUID{}, rootRecords.Processes)...)
	items = append(items, taskItemsForDocuments(
		"Aussonderung "+process.ProcessID.String(),
		[]uuid.UUID{}, rootRecords.Documents)...,
	)
	task := db.InsertTask(db.Task{
		Type:      db.ProcessStepArchiving,
		ProcessID: process.ProcessID,
		UserID:    userID,
		Items:     items,
		Data: ArchiveTaskData{
			CollectionID: collection.ID,
		},
	})
	tasks.Run(&task)
}

func taskItemsForFiles(
	m map[uuid.UUID]xdomea.PackagingDecision,
	path []uuid.UUID,
	files []db.FileRecord,
) []db.TaskItem {
	var items []db.TaskItem
	for _, f := range files {
		title := xdomea.FileRecordTitle(f, len(path) > 0)
		switch m[f.RecordID] {
		case xdomea.PackagingDecisionSingle:
			items = append(items, db.TaskItem{
				Label: title,
				State: db.TaskStatePending,
				Data: ArchiveItemData{
					Title:      title,
					RecordType: db.RecordTypeFile,
					RecordPath: path,
					RecordID:   f.RecordID,
				},
			})
		case xdomea.PackagingDecisionSub:
			items = append(items,
				taskItemsForFiles(m, append(path, f.RecordID), f.Subfiles)...)
			items = append(items,
				taskItemsForProcesses(m, append(path, f.RecordID), f.Processes)...)
			items = append(items,
				taskItemsForDocuments(title, append(path, f.RecordID), f.Documents)...)
		default:
			panic("no packaging decision for file record " + f.RecordID.String())
		}
	}
	return items
}

func taskItemsForProcesses(
	m map[uuid.UUID]xdomea.PackagingDecision,
	path []uuid.UUID,
	processes []db.ProcessRecord,
) []db.TaskItem {
	var items []db.TaskItem
	for _, p := range processes {
		switch m[p.RecordID] {
		case xdomea.PackagingDecisionSingle:
			title := xdomea.ProcessRecordTitle(p, false)
			items = append(items, db.TaskItem{
				Label: title,
				State: db.TaskStatePending,
				Data: ArchiveItemData{
					Title:      title,
					RecordType: db.RecordTypeProcess,
					RecordPath: path,
					RecordID:   p.RecordID,
				},
			})
		case xdomea.PackagingDecisionSub:
			panic("unexpected packaging decision for process record " +
				p.RecordID.String() + ": \"sub\"",
			)
		default:
			panic("no packaging decision for process record " + p.RecordID.String())
		}
	}
	return items
}

func taskItemsForDocuments(
	parentTitle string,
	path []uuid.UUID,
	documents []db.DocumentRecord,
) []db.TaskItem {
	var items []db.TaskItem
	if len(documents) > 0 {
		title := "Nicht zugeordnete Dokumente aus " + parentTitle
		items = append(items, db.TaskItem{
			Label: title,
			State: db.TaskStatePending,
			Data: ArchiveItemData{
				Title:      title,
				RecordType: db.RecordTypeDocument,
				RecordPath: path,
			},
		})
	}
	return items
}

type ArchiveTaskData struct {
	CollectionID primitive.ObjectID
}

type ArchiveItemData struct {
	Title      string
	RecordType db.RecordType
	RecordPath []uuid.UUID
	RecordID   uuid.UUID
	JobID      int
}

type recordsMap struct {
	Files         map[uuid.UUID]db.FileRecord
	Processes     map[uuid.UUID]db.ProcessRecord
	RootDocuments []db.DocumentRecord
}

type DimagData struct {
	Connection dimag.Connection
}

type ArchiveHandler struct {
	process       db.SubmissionProcess
	message       db.Message
	records       recordsMap
	archiveTarget string
	collection    db.ArchiveCollection
	targetData    interface{}
	t             *db.Task
}

// initArchiveHandler collects all necessary data for the given task from the
// database and populates a task-handler object that can process items of the
// task.
func initArchiveHandler(t *db.Task) (tasks.ItemHandler, error) {
	process, ok := db.FindProcess(context.Background(), t.ProcessID)
	if !ok {
		panic("failed to find process " + t.ProcessID.String())
	}
	message, ok := db.FindMessage(context.Background(), t.ProcessID, db.MessageType0503)
	if !ok {
		panic("failed to find 0503 message for process " + t.ProcessID.String())
	}
	d := db.UnmarshalData[ArchiveTaskData](t.Data)
	var collection db.ArchiveCollection
	if d.CollectionID != primitive.NilObjectID {
		collection, ok = db.FindArchiveCollection(context.Background(), d.CollectionID)
		if !ok {
			panic("failed to find archive collection  " + d.CollectionID.Hex())
		}
	}
	rootRecords := db.FindAllRootRecords(
		context.Background(), process.ProcessID, db.MessageType0503,
	)
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	var targetData interface{}
	if archiveTarget == "dimag" {
		c, err := dimag.InitConnection()
		if err != nil {
			return nil, err
		}
		targetData = interface{}(DimagData{Connection: c})
	}
	return &ArchiveHandler{
		process:       process,
		message:       message,
		records:       makeRecordsMap(rootRecords),
		collection:    collection,
		archiveTarget: archiveTarget,
		targetData:    targetData,
		t:             t,
	}, nil
}

func (h *ArchiveHandler) HandleItem(
	ctx context.Context,
	itemData interface{},
	updateItemData func(data interface{}),
) error {
	d := db.UnmarshalData[ArchiveItemData](itemData)
	var aip db.ArchivePackage
	switch d.RecordType {
	case db.RecordTypeFile:
		aip = createAipFromFileRecord(
			d.Title, h.process, d.RecordPath, h.records.Files[d.RecordID], h.collection.ID,
		)
	case db.RecordTypeProcess:
		aip = createAipFromProcessRecord(
			d.Title, h.process, d.RecordPath, h.records.Processes[d.RecordID], h.collection.ID,
		)
	case db.RecordTypeDocument:
		aip = createAipFromDocumentRecords(
			d.Title, h.process, d.RecordPath, h.records, h.collection.ID,
		)
	}
	// Check whether we already created the archive package. This can be the
	// case when we retry the task after an error.
	if existingAIP, found := db.FindArchivePackage(
		context.Background(), h.process.ProcessID, aip.RecordIDs,
	); found {
		aip = existingAIP
	} else {
		db.InsertArchivePackage(&aip)
	}
	var err error
	switch h.archiveTarget {
	case "filesystem":
		filesystem.StoreArchivePackage(h.process, h.message, aip)
	case "dimag":
		// We use DIMAG's asynchronous API. DIMAG creates a job with the import
		// data and runs it autonomously. Once the job is created, we save the
		// job ID. In case we encounter an error afterwards, we just continue
		// waiting for this job on retry.
		if d.JobID == 0 {
			targetData := h.targetData.(DimagData)
			jobID, err := dimag.StartImport(
				ctx, h.process, h.message, &aip, targetData.Connection,
			)
			if err != nil {
				return err
			}
			d.JobID = jobID
			updateItemData(d)
		}
		err = dimag.WaitForArchiveJob(ctx, d.JobID, &aip)
		// If the job failed, we want to create a new job on the next retry.
		if dimag.IsJobFailedError(err) {
			d.JobID = 0
			updateItemData(d)
		}
	default:
		panic("unknown archive target: " + h.archiveTarget)
	}
	return err
}

func (h *ArchiveHandler) Finish() {
	if h.archiveTarget == "dimag" {
		targetData := h.targetData.(DimagData)
		dimag.CloseConnection(targetData.Connection)
	}

}
func (h *ArchiveHandler) AfterDone() {
	if h.t.State != db.TaskStateDone {
		return
	}
	err := xdomea.Send0506Message(h.process, h.message)
	if err != nil {
		errorData := db.ProcessingError{
			Title:     "Fehler beim Senden der 0506-Nachricht",
			ProcessID: h.process.ProcessID,
		}
		errors.AddProcessingErrorWithData(err, errorData)
	}
	preferences := db.FindUserPreferencesWithDefault(context.Background(), h.t.UserID)
	if preferences.ReportByEmail {
		defer errors.HandlePanic("generate report for e-mail", &db.ProcessingError{
			ProcessID: h.process.ProcessID,
		})
		process, ok := db.FindProcess(context.Background(), h.process.ProcessID)
		if !ok {
			panic("failed to find process:" + process.ProcessID.String())
		}
		_, contentType, reader := report.GetReport(context.Background(), process)
		body, err := io.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		errorData := db.ProcessingError{
			Title:     "Fehler beim Versenden einer E-Mail-Benachrichtigung",
			ProcessID: h.process.ProcessID,
		}
		address, err := auth.GetMailAddress(h.t.UserID)
		if err != nil {
			errors.AddProcessingErrorWithData(err, errorData)
		} else {
			filename := fmt.Sprintf(
				"Übernahmebericht %s %s.pdf",
				process.Agency.Abbreviation, process.CreatedAt,
			)
			err = mail.SendMailReport(
				address, process,
				mail.Attachment{Filename: filename, ContentType: contentType, Body: body},
			)
			if err != nil {
				errors.AddProcessingErrorWithData(err, errorData)
			}
		}
	}
}

func makeRecordsMap(r db.RootRecords) recordsMap {
	m := recordsMap{
		Files:         make(map[uuid.UUID]db.FileRecord),
		Processes:     make(map[uuid.UUID]db.ProcessRecord),
		RootDocuments: r.Documents,
	}
	var processFiles func(files []db.FileRecord)
	var processProcesses func(processes []db.ProcessRecord)
	processFiles = func(files []db.FileRecord) {
		for _, f := range files {
			m.Files[f.RecordID] = f
			processFiles(f.Subfiles)
			processProcesses(f.Processes)
		}
	}
	processProcesses = func(processes []db.ProcessRecord) {
		for _, p := range processes {
			m.Processes[p.RecordID] = p
		}
	}
	processFiles(r.Files)
	processProcesses(r.Processes)
	return m
}

// createAipFromFileRecord creates the archive package metadata from a file
// record.
func createAipFromFileRecord(
	title string,
	process db.SubmissionProcess,
	path []uuid.UUID,
	f db.FileRecord,
	collectionID primitive.ObjectID,
) db.ArchivePackage {
	primaryDocuments := xdomea.GetPrimaryDocumentsForFile(&f)
	primaryDocuments, _ = xdomea.FilterMissingPrimaryDocuments(
		process.ProcessID, primaryDocuments,
	)
	archivePackageData := db.ArchivePackage{
		ProcessID:        process.ProcessID,
		IOTitle:          title,
		IOLifetime:       f.Lifetime,
		REPTitle:         "Original",
		PrimaryDocuments: primaryDocuments,
		RecordIDs:        []uuid.UUID{f.RecordID},
		RecordPath:       path,
		CollectionID:     collectionID,
	}
	return archivePackageData
}

// createAipFromProcessRecord creates the archive package metadata from a
// process record.
func createAipFromProcessRecord(
	title string,
	process db.SubmissionProcess,
	path []uuid.UUID,
	p db.ProcessRecord,
	collectionID primitive.ObjectID,
) db.ArchivePackage {
	primaryDocuments := xdomea.GetPrimaryDocumentsForProcess(&p)
	primaryDocuments, _ = xdomea.FilterMissingPrimaryDocuments(
		process.ProcessID, primaryDocuments,
	)
	archivePackageData := db.ArchivePackage{
		ProcessID:        process.ProcessID,
		IOTitle:          title,
		IOLifetime:       p.Lifetime,
		REPTitle:         "Original",
		PrimaryDocuments: primaryDocuments,
		RecordIDs:        []uuid.UUID{p.RecordID},
		RecordPath:       path,
		CollectionID:     collectionID,
	}
	return archivePackageData
}

// createAipFromDocumentRecords creates the metadata for a shared archive
// package of multiple documents.
func createAipFromDocumentRecords(
	title string,
	process db.SubmissionProcess,
	path []uuid.UUID,
	records recordsMap,
	collectionID primitive.ObjectID,
) db.ArchivePackage {
	// Find all documents for the record path.
	var documents []db.DocumentRecord
	var lifetime *db.Lifetime
	if len(path) == 0 {
		documents = records.RootDocuments
		lifetime = lifetimeFromDocuments(documents)
	} else {
		parentRecordID := path[len(path)-1]
		if parent, ok := records.Files[parentRecordID]; ok {
			documents = parent.Documents
			lifetime = parent.Lifetime
		} else if parent, ok := records.Processes[parentRecordID]; ok {
			documents = parent.Documents
			lifetime = parent.Lifetime
		} else {
			panic("could not find parent record: " + parentRecordID.String())
		}
	}
	var primaryDocuments []db.PrimaryDocumentContext
	for _, d := range documents {
		primaryDocuments = append(primaryDocuments,
			xdomea.GetPrimaryDocumentsForDocument(&d)...)
	}
	primaryDocuments, _ = xdomea.FilterMissingPrimaryDocuments(
		process.ProcessID, primaryDocuments,
	)
	var recordIDs []uuid.UUID
	for _, r := range documents {
		recordIDs = append(recordIDs, r.RecordID)
	}
	aip := db.ArchivePackage{
		ProcessID:        process.ProcessID,
		IOTitle:          title,
		IOLifetime:       lifetime,
		REPTitle:         "Original",
		PrimaryDocuments: primaryDocuments,
		RecordIDs:        recordIDs,
		RecordPath:       path,
		CollectionID:     collectionID,
	}
	return aip
}

// lifetimeFromDocuments reads the document date from all given documents and
// returns the lifetime as the time from the earliest to the latest document
// encountered.
func lifetimeFromDocuments(documents []db.DocumentRecord) *db.Lifetime {
	var start time.Time
	var end time.Time
	processDate := func(d string) {
		if d == "" {
			return
		}
		date, err := time.Parse(time.DateOnly, d)
		if err != nil {
			panic(err)
		}
		if start.IsZero() || date.Before(start) {
			start = date
		}
		if end.IsZero() || date.After(end) {
			end = date
		}
	}
	for _, d := range documents {
		processDate(d.DocumentDate)
	}
	if start.IsZero() {
		return nil
	} else {
		return &db.Lifetime{
			Start: start.Format(time.DateOnly),
			End:   end.Format(time.DateOnly),
		}
	}
}
