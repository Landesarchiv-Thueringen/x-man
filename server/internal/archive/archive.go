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
	rootRecords := db.FindRootRecords(context.Background(), process.ProcessID, db.MessageType0503)
	var items []db.TaskItem
	for _, f := range rootRecords.Files {
		items = append(items, db.TaskItem{
			Label: fileRecordTitle(f),
			State: db.TaskStatePending,
			Data: ArchiveItemData{
				RecordType: db.RecordTypeFile,
				RecordID:   f.RecordID,
			},
		})
	}
	for _, p := range rootRecords.Processes {
		items = append(items, db.TaskItem{
			Label: processRecordTitle(p),
			State: db.TaskStatePending,
			Data: ArchiveItemData{
				RecordType: db.RecordTypeProcess,
				RecordID:   p.RecordID,
			},
		})
	}
	if len(rootRecords.Documents) > 0 {
		items = append(items, db.TaskItem{
			Label: "Nicht zugeordnete Dokumente",
			State: db.TaskStatePending,
			Data: ArchiveItemData{
				RecordType: db.RecordTypeDocument,
			},
		})
	}
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

type ArchiveTaskData struct {
	CollectionID primitive.ObjectID
}

type ArchiveItemData struct {
	RecordType db.RecordType
	RecordID   uuid.UUID
}

type rootRecordsMap struct {
	Files     map[uuid.UUID]db.FileRecord
	Processes map[uuid.UUID]db.ProcessRecord
	Documents []db.DocumentRecord
}

type DimagData struct {
	Connection dimag.Connection
}

type ArchiveHandler struct {
	process       db.SubmissionProcess
	message       db.Message
	rootRecords   rootRecordsMap
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
	rootRecords := db.FindRootRecords(context.Background(), process.ProcessID, db.MessageType0503)
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
		rootRecords:   makeRootRecordsMap(rootRecords),
		collection:    collection,
		archiveTarget: archiveTarget,
		targetData:    targetData,
		t:             t,
	}, nil
}

func (h *ArchiveHandler) HandleItem(ctx context.Context, itemData interface{}) error {
	d := db.UnmarshalData[ArchiveItemData](itemData)
	var aip db.ArchivePackage
	switch d.RecordType {
	case db.RecordTypeFile:
		aip = createAipFromFileRecord(h.process, h.rootRecords.Files[d.RecordID], h.collection.ID)
	case db.RecordTypeProcess:
		aip = createAipFromProcessRecord(h.process, h.rootRecords.Processes[d.RecordID], h.collection.ID)
	case db.RecordTypeDocument:
		aip = createAipFromDocumentRecords(h.process, h.rootRecords.Documents, h.collection.ID)
	}
	db.InsertArchivePackage(aip)
	var err error
	switch h.archiveTarget {
	case "filesystem":
		filesystem.StoreArchivePackage(h.process, h.message, aip)
	case "dimag":
		targetData := h.targetData.(DimagData)
		err = dimag.ImportArchivePackage(ctx, h.process, h.message, &aip, targetData.Connection)
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
	xdomea.Send0506Message(h.process, h.message)
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
			filename := fmt.Sprintf("Übernahmebericht %s %s.pdf", process.Agency.Abbreviation, process.CreatedAt)
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

func makeRootRecordsMap(r db.RootRecords) rootRecordsMap {
	m := rootRecordsMap{
		Files:     make(map[uuid.UUID]db.FileRecord),
		Processes: make(map[uuid.UUID]db.ProcessRecord),
		Documents: r.Documents,
	}
	for _, f := range r.Files {
		m.Files[f.RecordID] = f
	}
	for _, p := range r.Processes {
		m.Processes[p.RecordID] = p
	}
	return m
}

// createAipFromFileRecord creates the archive package metadata from a file record object.
func createAipFromFileRecord(
	process db.SubmissionProcess,
	f db.FileRecord,
	collectionID primitive.ObjectID,
) db.ArchivePackage {
	archivePackageData := db.ArchivePackage{
		ProcessID:        process.ProcessID,
		IOTitle:          fileRecordTitle(f),
		IOLifetime:       f.Lifetime,
		REPTitle:         "Original",
		PrimaryDocuments: db.GetPrimaryDocumentsForFile(&f),
		RootRecordIDs:    []uuid.UUID{f.RecordID},
		CollectionID:     collectionID,
	}
	return archivePackageData
}

// createAipFromProcessRecord creates the archive package metadata from a process record object.
func createAipFromProcessRecord(
	process db.SubmissionProcess,
	p db.ProcessRecord,
	collectionID primitive.ObjectID,
) db.ArchivePackage {
	archivePackageData := db.ArchivePackage{
		ProcessID:        process.ProcessID,
		IOTitle:          processRecordTitle(p),
		IOLifetime:       p.Lifetime,
		REPTitle:         "Original",
		PrimaryDocuments: db.GetPrimaryDocumentsForProcess(&p),
		RootRecordIDs:    []uuid.UUID{p.RecordID},
		CollectionID:     collectionID,
	}
	return archivePackageData
}

// createAipFromDocumentRecords creates the metadata for a shared archive package of multiple documents.
func createAipFromDocumentRecords(
	process db.SubmissionProcess,
	documentRecords []db.DocumentRecord,
	collectionID primitive.ObjectID,

) db.ArchivePackage {
	var primaryDocuments []db.PrimaryDocument
	for _, d := range documentRecords {
		primaryDocuments = append(primaryDocuments, db.GetPrimaryDocumentsForDocument(&d)...)
	}
	ioTitle := "Nicht zugeordnete Dokumente Behörde: " + process.Agency.Name +
		" Prozess-ID: " + process.ProcessID.String()
	repTitle := "Original"
	var rootRecordIDs []uuid.UUID
	for _, r := range documentRecords {
		rootRecordIDs = append(rootRecordIDs, r.RecordID)
	}
	aip := db.ArchivePackage{
		ProcessID:        process.ProcessID,
		IOTitle:          ioTitle,
		IOLifetime:       nil,
		REPTitle:         repTitle,
		PrimaryDocuments: primaryDocuments,
		RootRecordIDs:    rootRecordIDs,
		CollectionID:     collectionID,
	}
	return aip
}

func fileRecordTitle(f db.FileRecord) string {
	title := "Akte"
	if f.GeneralMetadata != nil {
		if f.GeneralMetadata.RecordNumber != "" {
			title += " " + f.GeneralMetadata.RecordNumber
		}
		if f.GeneralMetadata.Subject != "" {
			title += ": " + f.GeneralMetadata.Subject
		}
	}
	return title
}

func processRecordTitle(p db.ProcessRecord) string {
	title := "Vorgang"
	if p.GeneralMetadata != nil {
		if p.GeneralMetadata.RecordNumber != "" {
			title += " " + p.GeneralMetadata.RecordNumber
		}
		if p.GeneralMetadata.Subject != "" {
			title += ": " + p.GeneralMetadata.Subject
		}
	}
	return title
}
