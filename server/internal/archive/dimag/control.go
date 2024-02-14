package dimag

import (
	"encoding/xml"
	"lath/xman/internal/db"
	"log"
)

const XmlHeader = "<?xml version='1.0' encoding='UTF-8'?>\n"
const ControlFileName string = "controlfile.xml"
const InformationObjectAbbreviation string = "O"
const RepresentationAbbreviation string = "R"
const FileAbbreviation string = "F"

func GenerateControlFile(
	message db.Message,
	archivePackageData ArchivePackageData,
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
		Title:    "xdomea XML-Datei",
		FilePath: message.GetRemoteXmlPath(importDir),
	}
	fileIndexItems = append(fileIndexItems, xmlIndexItem)
	repIndexItem := IndexItem{
		IndexID:    "",
		ItemType:   RepresentationAbbreviation,
		Title:      archivePackageData.REPTitle,
		IndexItems: fileIndexItems,
	}
	repItems := []IndexItem{repIndexItem}
	ioIndexItem := IndexItem{
		IndexID:    "",
		Lifetime:   archivePackageData.IOLifetime,
		ItemType:   InformationObjectAbbreviation,
		Title:      archivePackageData.IOTitle,
		IndexItems: repItems,
	}
	ioItems := []IndexItem{ioIndexItem}
	dimagControl := DimagControl{
		RootID:     "test-985",
		IndexItems: ioItems,
	}
	xmlBytes, err := xml.MarshalIndent(dimagControl, " ", " ")
	if err != nil {
		log.Fatal(err)
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
		Title:    primaryDocument.GetFileName(),
		FilePath: primaryDocument.GetRemotePath(importDir),
	}
	return fileIndexItem
}
