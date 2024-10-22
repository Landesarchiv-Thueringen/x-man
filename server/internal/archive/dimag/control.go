package dimag

import (
	"context"
	"encoding/xml"
	"lath/xman/internal/db"
	"path/filepath"

	"github.com/google/uuid"
)

var xmlHeader = []byte("<?xml version='1.0' encoding='UTF-8'?>\n")

type itemType string

const (
	itemTypeInformationObject itemType = "O"
	itemTypeRepresentation    itemType = "R"
	itemTypeFile              itemType = "F"
	itemTypeDocumentation     itemType = "D"
)

type controlRoot struct {
	XMLName    xml.Name `xml:"verzeichnungseinheit"`
	RootID     string   `xml:"rootid"`
	IndexItems []indexItem
}

type indexItem struct {
	XMLName     xml.Name `xml:"verz-obj"`
	IndexID     string   `xml:"aid"`
	AlternateID string   `xml:"alternate-id,omitempty"`
	Lifetime    string   `xml:"entstehungs-zeitraum"`
	FilePath    string   `xml:"sftp-dateiname,omitempty"`
	FileName    string   `xml:"dateiname"`
	Title       string   `xml:"titel"`
	ItemType    itemType `xml:"typ"`
	IndexItems  []indexItem
}

func generateControlFile(
	message db.Message,
	archivePackage db.ArchivePackage,
	importDir string,
	verificationResultsFilename string,
) (ioAlternateID string, fileContent []byte) {
	primaryDocuments := archivePackage.PrimaryDocuments
	fileIndexItems := []indexItem{}
	for _, d := range primaryDocuments {
		item := indexItem{
			ItemType: itemTypeFile,
			Title:    fileName(d.PrimaryDocument),
			FileName: fileName(d.PrimaryDocument),
			FilePath: filepath.Join(importDir, d.Filename),
		}
		fileIndexItems = append(fileIndexItems, item)
	}
	messageIndexItem := indexItem{
		ItemType: itemTypeFile,
		Title:    filepath.Base(message.MessagePath),
		FilePath: filepath.Join(importDir, filepath.Base(message.MessagePath)),
	}
	fileIndexItems = append(fileIndexItems, messageIndexItem)
	if verificationResultsFilename != "" {
		verificationResultsIndexItem := indexItem{
			ItemType: itemTypeDocumentation,
			Title:    "Ergebnisse der Formatverifikation",
			FilePath: filepath.Join(importDir, verificationResultsFilename),
		}
		fileIndexItems = append(fileIndexItems, verificationResultsIndexItem)
	}
	repIndexItem := indexItem{
		ItemType:   itemTypeRepresentation,
		Title:      archivePackage.REPTitle,
		IndexItems: fileIndexItems,
	}
	repItems := []indexItem{repIndexItem}
	ioAlternateID = uuid.NewString()
	ioIndexItem := indexItem{
		AlternateID: ioAlternateID,
		Lifetime:    combinedLifetime(archivePackage.IOLifetime),
		ItemType:    itemTypeInformationObject,
		Title:       archivePackage.IOTitle,
		IndexItems:  repItems,
	}
	archiveCollection, ok := db.FindArchiveCollection(
		context.Background(),
		archivePackage.CollectionID,
	)
	if !ok {
		panic("failed to find archive collection: " + archivePackage.CollectionID.Hex())
	}
	root := controlRoot{
		RootID:     archiveCollection.DimagID,
		IndexItems: []indexItem{ioIndexItem},
	}
	xmlBytes, err := xml.MarshalIndent(root, " ", " ")
	if err != nil {
		panic(err)
	}
	return ioAlternateID, append(xmlHeader, xmlBytes...)
}

func fileName(primaryDocument db.PrimaryDocument) string {
	if primaryDocument.FilenameOriginal == "" {
		return primaryDocument.Filename
	}
	return primaryDocument.FilenameOriginal
}

// combinedLifetime returns a string representation of lifetime start and end.
func combinedLifetime(l *db.Lifetime) string {
	if l != nil {
		if l.Start != "" && l.End != "" {
			return "Von " + l.Start + " bis " + l.End
		} else if l.Start != "" {
			return "Von " + l.Start
		} else if l.End != "" {
			return "Bis " + l.End
		}
	}
	return "Unbekannt"
}
