package dimag

import (
	"encoding/xml"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"
	"log"
)

const ControlFileName string = "controlfile.xml"
const InformationObjectAbbrevation string = "IO"
const RepresentationAbbrevation string = "R"
const FileAbbrevation string = "F"

func GenerateControlFile(
	message db.Message,
	fileRecordObject db.FileRecordObject,
	importDir string,
) string {
	primaryDocuments := fileRecordObject.GetPrimaryDocuments()
	fileIndexItems := []IndexItem{}
	for _, primaryDocument := range primaryDocuments {
		fileIndexItems = append(
			fileIndexItems,
			GetIndexItemForPrimaryDocument(primaryDocument, importDir),
		)
	}
	xmlIndexItem := IndexItem{
		IndexID:  "",
		ItemType: FileAbbrevation,
		Title:    "xdomea XML-Datei",
		FilePath: message.GetRemoteXmlPath(importDir),
	}
	fileIndexItems = append(fileIndexItems, xmlIndexItem)
	repIndexItem := IndexItem{
		IndexID:    "",
		ItemType:   RepresentationAbbrevation,
		Title:      fileRecordObject.GetTitle(),
		IndexItems: fileIndexItems,
	}
	repItems := []IndexItem{repIndexItem}
	ioIndexItem := IndexItem{
		IndexID:    "",
		ItemType:   InformationObjectAbbrevation,
		Title:      fileRecordObject.GetTitle(),
		IndexItems: repItems,
	}
	ioItems := []IndexItem{ioIndexItem}
	dimagControl := DimagControl{
		RootID:     "test-1",
		IndexItems: ioItems,
	}
	xmlBytes, err := xml.MarshalIndent(dimagControl, " ", " ")
	if err != nil {
		log.Fatal(err)
	}
	controlFileString := xdomea.XmlHeader + string(xmlBytes)
	return controlFileString
}

func GetIndexItemForPrimaryDocument(
	primaryDocument db.PrimaryDocument,
	importDir string,
) IndexItem {
	fileIndexItem := IndexItem{
		IndexID:  "",
		ItemType: FileAbbrevation,
		Title:    primaryDocument.GetFileName(),
		FilePath: primaryDocument.GetRemotePath(importDir),
	}
	return fileIndexItem
}
