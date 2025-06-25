package shared

import (
	"context"
	"encoding/json"
	"lath/xman/internal/db"
	"slices"

	"github.com/beevik/etree"
)

const ProtocolFilename = "xman_protocol.json"

var idPathXdomea = etree.MustCompilePath("./Identifikation/ID")

func GenerateProtocol(process db.SubmissionProcess) []byte {
	processingErrors := db.FindProcessingErrorsForProcess(context.Background(), process.ProcessID)
	protocol := map[string]any{
		"processState":     process.ProcessState,
		"processingErrors": processingErrors,
	}
	result, err := json.MarshalIndent(protocol, "", "  ")
	if err != nil {
		panic(err)
	}
	return result
}

// PruneMessage removes all records from the message which are no part of the
// archive package.
func PruneMessage(message db.Message, aip db.ArchivePackage) []byte {
	recordIDs := make([]string, len(aip.RecordIDs))
	copy(recordIDs, aip.RecordIDs)
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(message.MessagePath); err != nil {
		panic(err)
	}
	root := doc.Root()
	var file *etree.Element
	for i, id := range aip.RecordPath {
		if i == 0 {
			file = pruneRootRecords(root, []string{id})
		} else {
			file = pruneSubRecords(file, []string{id})
		}
	}
	if len(aip.RecordPath) == 0 {
		pruneRootRecords(root, recordIDs)
	} else {
		pruneSubRecords(file, recordIDs)
	}
	result, err := doc.WriteToBytes()
	if err != nil {
		panic(err)
	}
	return result
}

// pruneRootRecords removes all root records from the document that are not in
// `keepIDs`.
//
// Returns the last element that has been kept.
func pruneRootRecords(root *etree.Element, keepIDs []string) *etree.Element {
	genericRecords := root.SelectElements("Schriftgutobjekt")
	var kept *etree.Element
	for _, genericRecord := range genericRecords {
		recordEl := genericRecord.SelectElement("Akte")
		if recordEl == nil {
			recordEl = genericRecord.SelectElement("Vorgang")
		}
		if recordEl == nil {
			recordEl = genericRecord.SelectElement("Dokument")
		}
		if recordEl != nil {
			idEl := recordEl.FindElementPath(idPathXdomea)
			if idEl != nil {
				if slices.Contains(keepIDs, idEl.Text()) {
					kept = recordEl
				} else {
					removedChild := root.RemoveChild(genericRecord)
					// Should never happen unless the xdomea specification changes.
					if removedChild == nil {
						panic("failed to remove element")
					}
				}
			}
		}
	}
	if kept == nil {
		panic("no records left after pruning")
	}
	return kept
}

// pruneSubRecords removes all sub records from the given file record that are
// not in `keepIDs`.
//
// Returns the last element that has been kept.
func pruneSubRecords(file *etree.Element, keepIDs []string) *etree.Element {
	var kept *etree.Element
	// xdomea version < 3.0
	for _, recordEl := range file.SelectElements("Teilakte") {
		idEl := recordEl.FindElementPath(idPathXdomea)
		if idEl != nil {
			if slices.Contains(keepIDs, idEl.Text()) {
				kept = recordEl
			} else {
				removedEl := file.RemoveChild(recordEl)
				if removedEl == nil {
					panic("failed to remove element")
				}
			}
		}
	}
	// all xdomea versions
	fileContent := file.SelectElement("Akteninhalt")
	var subRecords []*etree.Element
	subRecords = append(subRecords, fileContent.SelectElements("Teilakte")...)
	subRecords = append(subRecords, fileContent.SelectElements("Vorgang")...)
	subRecords = append(subRecords, fileContent.SelectElements("Dokument")...)
	for _, recordEl := range subRecords {
		idEl := recordEl.FindElementPath(idPathXdomea)
		if idEl != nil {
			if slices.Contains(keepIDs, idEl.Text()) {
				kept = recordEl
			} else {
				removedEl := fileContent.RemoveChild(recordEl)
				if removedEl == nil {
					panic("failed to remove element")
				}
			}
		}
	}
	if kept == nil {
		panic("no records left after pruning")
	}
	return kept
}
