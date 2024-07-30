package dimag

import (
	"context"
	"encoding/xml"
	"lath/xman/internal/db"
	"path/filepath"
)

var XmlHeader = []byte("<?xml version='1.0' encoding='UTF-8'?>\n")

type itemType string

const (
	itemTypeInformationObject itemType = "O"
	itemTypeRepresentation    itemType = "R"
	itemTypeFile              itemType = "F"
)

type dimagControl struct {
	XMLName    xml.Name `xml:"verzeichnungseinheit"`
	RootID     string   `xml:"rootid"`
	IndexItems []indexItem
}

type indexItem struct {
	XMLName    xml.Name `xml:"verz-obj"`
	IndexID    string   `xml:"aid"`
	Lifetime   string   `xml:"entstehungs-zeitraum"`
	FilePath   string   `xml:"sftp-dateiname,omitempty"`
	FileName   string   `xml:"dateiname"`
	Title      string   `xml:"titel"`
	ItemType   itemType `xml:"typ"`
	IndexItems []indexItem
}

func generateControlFile(
	message db.Message,
	archivePackage db.ArchivePackage,
	importDir string,
) []byte {
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
	// protocolIndexItem := indexItem{
	// 	ItemType: itemTypeFile,
	// 	Title:    shared.ProtocolFilename,
	// 	FileName: shared.ProtocolFilename,
	// 	FilePath: filepath.Join(importDir, shared.ProtocolFilename),
	// }
	// fileIndexItems = append(fileIndexItems, protocolIndexItem)
	repIndexItem := indexItem{
		ItemType:   itemTypeRepresentation,
		Title:      archivePackage.REPTitle,
		IndexItems: fileIndexItems,
	}
	repItems := []indexItem{repIndexItem}
	ioIndexItem := indexItem{
		Lifetime:   combinedLifetime(archivePackage.IOLifetime),
		ItemType:   itemTypeInformationObject,
		Title:      archivePackage.IOTitle,
		IndexItems: repItems,
	}
	ioItems := []indexItem{ioIndexItem}
	archiveCollection, ok := db.FindArchiveCollection(
		context.Background(),
		archivePackage.CollectionID,
	)
	if !ok {
		panic("failed to find archive collection: " + archivePackage.CollectionID.Hex())
	}
	dimagControl := dimagControl{
		RootID:     archiveCollection.DimagID,
		IndexItems: ioItems,
	}
	xmlBytes, err := xml.MarshalIndent(dimagControl, " ", " ")
	if err != nil {
		panic(err)
	}
	return append(XmlHeader, xmlBytes...)
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
			return l.Start + " - " + l.End
		} else if l.Start != "" {
			return l.Start + " - "
		} else if l.End != "" {
			return " - " + l.End
		}
	}
	return ""
}
