package dimag

import (
	"encoding/xml"
	"lath/xman/internal/db"
	"lath/xman/internal/xdomea"
	"log"
)

const DimagInformationObjectAbbrevation string = "IO"
const DimagRepresentationAbbrevation string = "R"
const DimagFileAbbrevation string = "F"

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
		ItemType: DimagFileAbbrevation,
		Title:    "xdomea XML-Datei",
		FilePath: message.GetRemoteXmlPath(importDir),
	}
	fileIndexItems = append(fileIndexItems, xmlIndexItem)
	repIndexItem := IndexItem{
		IndexID:    "",
		ItemType:   DimagRepresentationAbbrevation,
		Title:      fileRecordObject.GetTitle(),
		IndexItems: fileIndexItems,
	}
	repItems := []IndexItem{repIndexItem}
	ioIndexItem := IndexItem{
		IndexID:    "",
		ItemType:   DimagInformationObjectAbbrevation,
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
		ItemType: DimagFileAbbrevation,
		Title:    primaryDocument.GetFileName(),
		FilePath: primaryDocument.GetRemotePath(importDir),
	}
	return fileIndexItem
}
