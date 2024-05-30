package dimag

import (
	"context"
	"encoding/xml"
	"lath/xman/internal/archive"
	"lath/xman/internal/db"
	"path/filepath"
)

const XmlHeader = "<?xml version='1.0' encoding='UTF-8'?>\n"
const ControlFileName string = "controlfile.xml"
const InformationObjectAbbreviation string = "O"
const RepresentationAbbreviation string = "R"
const FileAbbreviation string = "F"

func GenerateControlFile(
	message db.Message,
	archivePackageData db.ArchivePackage,
	importDir string,
) string {
	primaryDocuments := archivePackageData.PrimaryDocuments
	fileIndexItems := []IndexItem{}
	for _, primaryDocument := range primaryDocuments {
		fileIndexItems = append(
			fileIndexItems,
			GetIndexItemForPrimaryDocument(primaryDocument, importDir),
		)
	}
	xmlIndexItem := IndexItem{
		IndexID:  "",
		ItemType: FileAbbreviation,
		Title:    filepath.Base(message.MessagePath),
		FilePath: getRemoteXmlPath(message, importDir),
	}
	protocolIndexItem := GetIndexItemForProtocol(importDir)
	fileIndexItems = append(fileIndexItems, xmlIndexItem, protocolIndexItem)
	repIndexItem := IndexItem{
		IndexID:    "",
		ItemType:   RepresentationAbbreviation,
		Title:      archivePackageData.REPTitle,
		IndexItems: fileIndexItems,
	}
	repItems := []IndexItem{repIndexItem}
	ioIndexItem := IndexItem{
		IndexID:    "",
		Lifetime:   getCombinedLifetime(archivePackageData.IOLifetime),
		ItemType:   InformationObjectAbbreviation,
		Title:      archivePackageData.IOTitle,
		IndexItems: repItems,
	}
	ioItems := []IndexItem{ioIndexItem}
	archiveCollection, ok := db.FindArchiveCollection(context.Background(), archivePackageData.CollectionID)
	if !ok {
		panic("failed to find archive collection: " + archivePackageData.CollectionID.Hex())
	}
	dimagControl := DimagControl{
		RootID:     archiveCollection.DimagID,
		IndexItems: ioItems,
	}
	xmlBytes, err := xml.MarshalIndent(dimagControl, " ", " ")
	if err != nil {
		panic(err)
	}
	controlFileString := XmlHeader + string(xmlBytes)
	return controlFileString
}

func GetIndexItemForPrimaryDocument(
	primaryDocument db.PrimaryDocument,
	importDir string,
) IndexItem {
	fileIndexItem := IndexItem{
		IndexID:  "",
		ItemType: FileAbbreviation,
		Title:    getFileName(primaryDocument),
		FileName: getFileName(primaryDocument),
		FilePath: filepath.Join(importDir, primaryDocument.Filename),
	}
	return fileIndexItem
}

func GetIndexItemForProtocol(
	importDir string,
) IndexItem {
	fileIndexItem := IndexItem{
		IndexID:  "",
		ItemType: FileAbbreviation,
		Title:    archive.ProtocolFilename,
		FileName: archive.ProtocolFilename,
		FilePath: filepath.Join(importDir, archive.ProtocolFilename),
	}
	return fileIndexItem
}

func getFileName(primaryDocument db.PrimaryDocument) string {
	if primaryDocument.FilenameOriginal == "" {
		return primaryDocument.Filename
	}
	return primaryDocument.FilenameOriginal
}

func getRemoteXmlPath(message db.Message, importDir string) string {
	filename := filepath.Base(message.MessagePath)
	return filepath.Join(importDir, filename)
}

// getCombinedLifetime returns a string representation of lifetime start and end.
func getCombinedLifetime(l *db.Lifetime) string {
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
