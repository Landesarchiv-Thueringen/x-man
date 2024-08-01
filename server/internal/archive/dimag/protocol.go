package dimag

import (
	"context"
	"encoding/xml"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type protocolFileRoot struct {
	XMLName xml.Name `xml:"protokollierung"`
	Objects []protocolObject
}

type protocolObject struct {
	XMLName           xml.Name `xml:"protokollobjekt"`
	ParentAlternateID string   `xml:"parent-alternate-id"`
	Entries           []protocolEntry
}

type protocolEntry struct {
	XMLName     xml.Name `xml:"eintrag"`
	CompletedAt string   `xml:"prozess-ende"`          // 2006-01-02 15:04:05
	CompletedBy string   `xml:"prozess-ausfuehrender"` // John Doe
	Action      string   `xml:"prozess"`               // Do thing
	Reference   string   `xml:"bezug"`                 // file.txt
	Info        string   `xml:"naehere-angaben"`       // Do thing with file.txt: ok, more things
}

func generateProtocolFile(
	process db.SubmissionProcess,
	archivePackage db.ArchivePackage,
	ioAlternateID string,
) []byte {
	entries := append(
		entriesForCompletedSteps(process),
		entryForArchiving(process),
	)
	entries = append(entries,
		entriesForProcessingErrors(process)...,
	)
	protocol := protocolObject{ParentAlternateID: ioAlternateID, Entries: entries}
	root := protocolFileRoot{Objects: []protocolObject{protocol}}
	xmlBytes, err := xml.MarshalIndent(root, " ", " ")
	if err != nil {
		panic(err)
	}
	return append(xmlHeader, xmlBytes...)
}

// entriesForCompletedSteps creates a protocol entry for each completed
// processing step of the submission process.
func entriesForCompletedSteps(process db.SubmissionProcess) []protocolEntry {
	var entries []protocolEntry
	s := process.ProcessState
	if s.Receive0501.Complete {
		entries = append(entries, protocolEntry{
			CompletedAt: s.Receive0501.CompletedAt.Local().Format("2006-01-02 15:04:05"),
			Action:      "Empfangen",
			Reference:   process.ProcessID.String() + xdomea.Message0501MessageSuffix + ".zip",
			Info:        "Anbietung empfangen",
		})
	}
	if s.Appraisal.Complete {
		entries = append(entries, protocolEntry{
			CompletedAt: s.Appraisal.CompletedAt.Local().Format("2006-01-02 15:04:05"),
			CompletedBy: s.Appraisal.CompletedBy,
			Action:      "Bewertung",
			Reference:   process.ProcessID.String() + xdomea.Message0501MessageSuffix + ".zip",
			Info:        "Anbietung bewerten",
		})
	}
	if s.Receive0505.Complete {
		entries = append(entries, protocolEntry{
			CompletedAt: s.Receive0505.CompletedAt.Local().Format("2006-01-02 15:04:05"),
			Action:      "Empfangen",
			Reference:   process.ProcessID.String() + xdomea.Message0505MessageSuffix + ".zip",
			Info:        "Empfangsbestätigung für Bewertung erhalten",
		})
	}
	if s.Receive0503.Complete {
		entries = append(entries, protocolEntry{
			CompletedAt: s.Receive0503.CompletedAt.Local().Format("2006-01-02 15:04:05"),
			Action:      "Empfangen",
			Reference:   process.ProcessID.String() + xdomea.Message0503MessageSuffix + ".zip",
			Info:        "Abgabe empfangen",
		})
	}
	if s.FormatVerification.Complete {
		// We check the corresponding task for errors, since the process step
		// will have been marked as completed without errors at this point.
		if s.FormatVerification.TaskID == primitive.NilObjectID {
			panic("format-verification task ID not set")
		}
		task, ok := db.FindTask(context.Background(), s.FormatVerification.TaskID)
		var status string
		if task.Error == "" {
			status = "ok"
		} else {
			status = task.Error
		}
		if !ok {
			panic("failed to find task: " + s.FormatVerification.TaskID.Hex())
		}
		entries = append(entries, protocolEntry{
			CompletedAt: s.FormatVerification.CompletedAt.Local().Format("2006-01-02 15:04:05"),
			Action:      "Formatverifikation",
			Reference:   "Aussonderung " + process.ProcessID.String(),
			Info:        "Formatverifikation aller Primärdateien: " + status,
		})
	}
	return entries
}

// entryForArchiving creates a protocol entry for the archiving process.
func entryForArchiving(process db.SubmissionProcess) protocolEntry {
	// At the time this function is called, the archiving process is still
	// running.
	s := process.ProcessState
	if s.Archiving.TaskID == primitive.NilObjectID {
		panic("archiving task ID not set")
	}
	task, ok := db.FindTask(context.Background(), s.Archiving.TaskID)
	if !ok {
		panic("failed to find task: " + s.Archiving.TaskID.Hex())
	}
	return protocolEntry{
		CompletedAt: s.Archiving.UpdatedAt.Local().Format("2006-01-02 15:04:05"),
		CompletedBy: auth.GetDisplayName(task.UserID),
		Action:      "Ingest",
		Reference:   "Aussonderung " + process.ProcessID.String(),
		Info:        "Ingest starten",
	}
}

// entryForArchiving creates a protocol entry for each processing error related
// to the submission process.
func entriesForProcessingErrors(process db.SubmissionProcess) []protocolEntry {
	processingErrors := db.FindProcessingErrorsForProcess(context.Background(), process.ProcessID)
	var entries []protocolEntry
	for _, e := range processingErrors {
		entries = append(entries, protocolEntry{
			CompletedAt: e.CreatedAt.Local().Format("2006-01-02 15:04:05"),
			Action:      "Fehler",
			Reference:   "Aussonderung " + process.ProcessID.String(),
			Info:        e.Title + "\nLösung: " + string(e.Resolution),
		})
	}
	return entries
}
