package xdomea

import (
	"errors"
	"lath/xman/internal/db"
	"log"
	"slices"

	"github.com/beevik/etree"
)

var idPathXdomea = etree.MustCompilePath("./Identifikation/ID")

// PruneMessage removes all records from message which are no part of the archive package.
func PruneMessage(message db.Message, aip db.ArchivePackage) (string, error) {
	rootIDs := aip.GetRootIDs()
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(message.MessagePath); err != nil {
		return "", err
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
				log.Println(rootIDs)
				log.Println(idEl.Text())
				log.Println(slices.Contains(rootIDs, idEl.Text()))
				if !slices.Contains(rootIDs, idEl.Text()) {
					removedChild := root.RemoveChild(genericRecord)
					// Should never happen unless the xdomea specification changes.
					if removedChild == nil {
						return "", errors.New("")
					}
				}
			}
		}
	}
	return doc.WriteToString()
}
