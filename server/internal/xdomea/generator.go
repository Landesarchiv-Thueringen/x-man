package xdomea

import (
	"encoding/xml"
	"fmt"
	"lath/xman/internal/db"
	"log"
	"os"
	"time"

	"github.com/lestrrat-go/libxml2/xsd"
)

const XmlHeader = "<?xml version='1.0' encoding='UTF-8'?>\n"
const XsiXmlNs = "http://www.w3.org/2001/XMLSchema-instance"

// Generate0502Message creates the XML code for the appraisal message (code: 0502).
// The generated message has the same xdomea version as the
func Generate0502Message(message db.Message) string {
	xdomeaVersion, err := db.GetXdomeaVersionByCode(message.XdomeaVersion)
	if err != nil {
		panic(err)
	}
	messageHead := GenerateMessageHead0502(message.MessageHead.ProcessID, message.MessageHead.Sender)
	message0502 := db.GeneratorMessage0502{
		XdomeaXmlNs: xdomeaVersion.URI,
		XsiXmlNs:    XsiXmlNs,
		MessageHead: messageHead,
	}
	for _, o := range message.GetAppraisableObjects() {
		appraisedObject := GenerateAppraisedObject(messageHead.ProcessID, o, xdomeaVersion)
		message0502.AppraisedObjects = append(message0502.AppraisedObjects, appraisedObject)
	}
	xmlBytes, err := xml.MarshalIndent(message0502, " ", " ")
	if err != nil {
		panic("0502 message couldn't be created")
	}
	messageXml := XmlHeader + string(xmlBytes)
	valid, err := ValidateXdomeaXmlString(messageXml, xdomeaVersion)
	if err != nil {
		validationError, ok := err.(xsd.SchemaValidationError)
		if ok {
			for _, e := range validationError.Errors() {
				log.Printf("XML schema error: %s", e.Error())
			}
		}
		if !valid {
			panic("generated 0502 message is invalid")
		}
	}
	return messageXml
}

// GenerateAppraisedObject returns xdomea version dependent appraised object.
func GenerateAppraisedObject(
	processID string,
	o db.AppraisableRecordObject,
	xdomeaVersion db.XdomeaVersion,
) db.GeneratorAppraisedObject {
	var appraisedObject db.GeneratorAppraisedObject
	appraisal := db.GetAppraisal(processID, o.GetID())
	if appraisal.Decision != "A" && appraisal.Decision != "V" {
		panic(fmt.Sprintf("called GenerateAppraisedObject with appraisal \"%s\": %v", appraisal.Decision, o.GetID()))
	}

	var objectAppraisal db.GeneratorObjectAppraisal
	if xdomeaVersion.IsVersionPriorTo300() {
		objectAppraisal = db.GeneratorObjectAppraisal{
			AppraisalCodePre300: (*string)(&appraisal.Decision),
		}
	} else {
		appraisalCode := db.GeneratorCode{
			Code: string(appraisal.Decision),
		}
		objectAppraisal = db.GeneratorObjectAppraisal{
			AppraisalCode: &appraisalCode,
		}
	}
	appraisedObject = db.GeneratorAppraisedObject{
		XdomeaID:        o.GetID(),
		ObjectAppraisal: objectAppraisal,
	}
	return appraisedObject
}

func Generate0504Message(message db.Message) string {
	xdomeaVersion, err := db.GetXdomeaVersionByCode(message.XdomeaVersion)
	if err != nil {
		panic(err)
	}
	messageHead := GenerateMessageHead(message.MessageHead.ProcessID, message.MessageHead.Sender, "0504")
	message0504 := db.GeneratorMessage0504{
		XdomeaXmlNs: xdomeaVersion.URI,
		XsiXmlNs:    XsiXmlNs,
		MessageHead: messageHead,
	}
	xmlBytes, err := xml.MarshalIndent(message0504, " ", " ")
	if err != nil {
		panic("0504 message couldn't be created")
	}
	messageXml := XmlHeader + string(xmlBytes)
	valid, err := ValidateXdomeaXmlString(messageXml, xdomeaVersion)
	if err != nil {
		validationError, ok := err.(xsd.SchemaValidationError)
		if ok {
			for _, e := range validationError.Errors() {
				log.Printf("XML schema error: %s", e.Error())
			}
		}
		if !valid {
			panic("generated 0504 message is invalid")
		}
	}
	return messageXml
}

func GenerateMessageHead0502(processID string, sender db.Contact) db.GeneratorMessageHead0502 {
	messageType := db.GeneratorCode{
		Code: "0502",
	}
	timeStamp := time.Now()
	lathContact := GetLAThContact()
	sendingSystem := GetSendingSystem()
	messageHead := db.GeneratorMessageHead0502{
		ProcessID:        processID,
		MessageType:      messageType,
		CreationTime:     timeStamp.Format("2006-01-02T15:04:05"),
		Sender:           lathContact,
		Receiver:         ConvertParserToGeneratorContact(sender),
		SendingSystem:    sendingSystem,
		ReceiptRequested: true,
	}
	return messageHead
}

func GenerateMessageHead(processID string, sender db.Contact, messageCode string) db.GeneratorMessageHead {
	messageType := db.GeneratorCode{
		Code: messageCode,
	}
	timeStamp := time.Now()
	lathContact := GetLAThContact()
	sendingSystem := GetSendingSystem()
	messageHead := db.GeneratorMessageHead{
		ProcessID:     processID,
		MessageType:   messageType,
		CreationTime:  timeStamp.Format("2006-01-02T15:04:05"),
		Sender:        lathContact,
		Receiver:      ConvertParserToGeneratorContact(sender),
		SendingSystem: sendingSystem,
	}
	return messageHead
}

func Generate0506Message(message db.Message, archivePackages []db.ArchivePackage) string {
	xdomeaVersion, err := db.GetXdomeaVersionByCode(message.XdomeaVersion)
	if err != nil {
		panic(err)
	}
	messageHead := GenerateMessageHead(message.MessageHead.ProcessID, message.MessageHead.Sender, "0506")
	message0506 := db.GeneratorMessage0506{
		XdomeaXmlNs: xdomeaVersion.URI,
		XsiXmlNs:    XsiXmlNs,
		MessageHead: messageHead,
	}
	if xdomeaVersion.IsVersionPriorTo300() {
		info := GetArchivingInfoPre300(archivePackages)
		message0506.ArchivingInfoPre300 = &info
	} else {
		message0506.ArchivedRecordInfo = GetArchivedRecordInfo(archivePackages)
	}
	xmlBytes, err := xml.MarshalIndent(message0506, " ", " ")
	messageXml := XmlHeader + string(xmlBytes)
	if err != nil {
		panic("0506 message couldn't be created")
	}
	valid, err := ValidateXdomeaXmlString(messageXml, xdomeaVersion)
	if err != nil {
		validationError, ok := err.(xsd.SchemaValidationError)
		if ok {
			for _, e := range validationError.Errors() {
				log.Printf("XML schema error: %s", e.Error())
			}
		}
		if !valid {
			panic("generated 0506 message is invalid")
		}
	}
	return messageXml
}

func GetArchivingInfoPre300(archivePackages []db.ArchivePackage) db.GeneratorArchivingInfoPre300 {
	info := db.GeneratorArchivingInfoPre300{
		Success: true,
	}
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	if archiveTarget == "dimag" {
		var recordArchiveMapping []db.GeneratorRecordArchiveMapping
		for _, aip := range archivePackages {
			for _, fileRecord := range aip.FileRecordObjects {
				idMapping := db.GeneratorRecordArchiveMapping{
					RecordID:  fileRecord.XdomeaID.String(),
					ArchiveID: aip.PackageID,
				}
				recordArchiveMapping = append(recordArchiveMapping, idMapping)
			}
			for _, processRecord := range aip.ProcessRecordObjects {
				idMapping := db.GeneratorRecordArchiveMapping{
					RecordID:  processRecord.XdomeaID.String(),
					ArchiveID: aip.PackageID,
				}
				recordArchiveMapping = append(recordArchiveMapping, idMapping)
			}
			for _, documentRecord := range aip.DocumentRecordObjects {
				idMapping := db.GeneratorRecordArchiveMapping{
					RecordID:  documentRecord.XdomeaID.String(),
					ArchiveID: aip.PackageID,
				}
				recordArchiveMapping = append(recordArchiveMapping, idMapping)
			}
		}
		info.RecordArchiveMapping = recordArchiveMapping
	}
	return info
}

func GetArchivedRecordInfo(archivePackages []db.ArchivePackage) []db.GeneratorArchivedRecordInfo {
	var info []db.GeneratorArchivedRecordInfo
	for _, aip := range archivePackages {
		for _, fileRecord := range aip.FileRecordObjects {
			info = append(info, GetArchivedRecordIDMapping(fileRecord.XdomeaID.String(), aip))
		}
		for _, processRecord := range aip.ProcessRecordObjects {
			info = append(info, GetArchivedRecordIDMapping(processRecord.XdomeaID.String(), aip))
		}
		for _, documentRecord := range aip.DocumentRecordObjects {
			info = append(info, GetArchivedRecordIDMapping(documentRecord.XdomeaID.String(), aip))
		}
	}
	return info
}

func GetArchivedRecordIDMapping(recordID string, aip db.ArchivePackage) db.GeneratorArchivedRecordInfo {
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	idMapping := db.GeneratorArchivedRecordInfo{
		RecordID: recordID,
		Success:  true,
	}
	if archiveTarget == "dimag" {
		idMapping.ArchiveID = &aip.PackageID
	}
	return idMapping
}

func GetLAThContact() db.GeneratorContact {
	institutionName := "Landesarchiv Thüringen"
	institutionAbbreviation := "LATh"
	institution := db.GeneratorInstitution{
		Name:         &institutionName,
		Abbreviation: &institutionAbbreviation,
	}
	contact := db.GeneratorContact{
		Institution: &institution,
	}
	return contact
}

func ConvertParserToGeneratorContact(contact db.Contact) db.GeneratorContact {
	var generatorContact db.GeneratorContact
	if contact.Institution != nil {
		institution := db.GeneratorInstitution{
			Name:         contact.Institution.Name,
			Abbreviation: contact.Institution.Abbreviation,
		}
		generatorContact.Institution = &institution
	}
	return generatorContact
}

func GetSendingSystem() db.GeneratorSendingSystem {
	productName := "X-MAN"
	productVersion := "0.1"
	sendingSystem := db.GeneratorSendingSystem{
		ProductName:    &productName,
		ProductVersion: &productVersion,
	}
	return sendingSystem
}
