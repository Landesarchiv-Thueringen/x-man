package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"lath/xman/internal/db"
	"slices"

	"github.com/beevik/etree"
)

const ProtocolFilename = "xman_protocol.json"

var idPathXdomea = etree.MustCompilePath("./Identifikation/ID")

func GenerateProtocol(process db.SubmissionProcess) []byte {
	processStateBytes, err := json.MarshalIndent(process.ProcessState, "", " ")
	if err != nil {
		panic(err)
	}
	protocol := append(processStateBytes, '\n')
	processingErrors := db.FindProcessingErrorsForProcess(context.Background(), process.ProcessID)
	if len(processingErrors) > 0 {
		errorsBytes, err := json.MarshalIndent(processingErrors, "", " ")
		if err != nil {
			panic(err)
		}
		protocol = fmt.Appendln(protocol, errorsBytes)
	}
	return protocol
}

// PruneMessage removes all records from the message which are no part of the
// archive package.
func PruneMessage(message db.Message, aip db.ArchivePackage) []byte {
	rootRecordIDs := make([]string, len(aip.RootRecordIDs))
	for i, id := range aip.RootRecordIDs {
		rootRecordIDs[i] = id.String()
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(message.MessagePath); err != nil {
		panic(err)
	}
	root := doc.Root()
	genericRecords := root.SelectElements("Schriftgutobjekt")
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
				if !slices.Contains(rootRecordIDs, idEl.Text()) {
					removedChild := root.RemoveChild(genericRecord)
					// Should never happen unless the xdomea specification changes.
					if removedChild == nil {
						panic("removedChild == nil")
					}
				}
			}
		}
	}
	result, err := doc.WriteToBytes()
	if err != nil {
		panic(err)
	}
	return result
}
