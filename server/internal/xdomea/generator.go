package xdomea

import (
	"context"
	"encoding/xml"
	"fmt"
	"lath/xman/internal/db"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/libxml2/xsd"
)

const XmlHeader = "<?xml version='1.0' encoding='UTF-8'?>\n"
const XsiXmlNs = "http://www.w3.org/2001/XMLSchema-instance"

type generatorMessage0502 struct {
	XMLName          xml.Name                   `xml:"xdomea:Aussonderung.Bewertungsverzeichnis.0502"`
	MessageHead      generatorMessageHead0502   `xml:"xdomea:Kopf"`
	AppraisedObjects []generatorAppraisedObject `xml:"xdomea:BewertetesObjekt"`
	XdomeaXmlNs      string                     `xml:"xmlns:xdomea,attr"`
	XsiXmlNs         string                     `xml:"xmlns:xsi,attr"`
}

type generatorMessage0504 struct {
	XMLName     xml.Name             `gorm:"-" xml:"xdomea:Aussonderung.AnbietungEmpfangBestaetigen.0504" json:"-"`
	MessageHead generatorMessageHead `xml:"xdomea:Kopf" json:"messageHead"`
	XdomeaXmlNs string               `xml:"xmlns:xdomea,attr"`
	XsiXmlNs    string               `xml:"xmlns:xsi,attr"`
}

type generatorMessage0506 struct {
	XMLName             xml.Name                      `gorm:"-" xml:"xdomea:Aussonderung.AussonderungImportBestaetigen.0506" json:"-"`
	MessageHead         generatorMessageHead          `xml:"xdomea:Kopf" json:"messageHead"`
	XdomeaXmlNs         string                        `xml:"xmlns:xdomea,attr"`
	XsiXmlNs            string                        `xml:"xmlns:xsi,attr"`
	ArchivingInfoPre300 *generatorArchivingInfoPre300 `xml:"xdomea:ErfolgOderMisserfolg"`
	ArchivedRecordInfo  []generatorArchivedRecordInfo `xml:"xdomea:AusgesondertesSGO"`
}

type generatorAppraisedObject struct {
	XMLName         xml.Name                 `xml:"xdomea:BewertetesObjekt"`
	RecordID        uuid.UUID                `xml:"xdomea:ID"`
	ObjectAppraisal generatorObjectAppraisal `xml:"xdomea:Aussonderungsart"`
}

type generatorObjectAppraisal struct {
	XMLName             xml.Name       `xml:"xdomea:Aussonderungsart"`
	AppraisalCode       *generatorCode `xml:"xdomea:Aussonderungsart"`
	AppraisalCodePre300 *string        `xml:"code"`
}

type generatorMessageHead0502 struct {
	ProcessID        uuid.UUID              `xml:"xdomea:ProzessID"`
	MessageType      generatorCode          `xml:"xdomea:Nachrichtentyp"`
	CreationTime     string                 `xml:"xdomea:Erstellungszeitpunkt"`
	Sender           generatorContact       `xml:"xdomea:Absender"`
	Receiver         generatorContact       `xml:"xdomea:Empfaenger"`
	SendingSystem    generatorSendingSystem `xml:"xdomea:SendendesSystem"`
	ReceiptRequested bool                   `xml:"xdomea:Empfangsbestaetigung"`
}

type generatorMessageHead struct {
	ProcessID     uuid.UUID              `xml:"xdomea:ProzessID"`
	MessageType   generatorCode          `xml:"xdomea:Nachrichtentyp"`
	CreationTime  string                 `xml:"xdomea:Erstellungszeitpunkt"`
	Sender        generatorContact       `xml:"xdomea:Absender"`
	Receiver      generatorContact       `xml:"xdomea:Empfaenger"`
	SendingSystem generatorSendingSystem `xml:"xdomea:SendendesSystem"`
}

type generatorSendingSystem struct {
	XMLName        xml.Name `xml:"xdomea:SendendesSystem"`
	ProductName    *string  `xml:"xdomea:Produktname"`
	ProductVersion *string  `xml:"xdomea:Version"`
}

type generatorContact struct {
	AgencyIdentification *generatorAgencyIdentification `xml:"xdomea:Behoerdenkennung"`
	Institution          *generatorInstitution          `xml:"xdomea:Institution"`
}

type generatorAgencyIdentification struct {
	Code   *generatorCode `xml:"xdomea:Behoerdenschluessel"`
	Prefix *generatorCode `xml:"xdomea:Praefix"`
}

type generatorInstitution struct {
	Name         *string `xml:"xdomea:Name"`
	Abbreviation *string `xml:"xdomea:Kurzbezeichnung"`
}

type generatorCode struct {
	Code string `xml:"code"`
}

type generatorArchivingInfoPre300 struct {
	Success              bool                            `xml:"xdomea:Erfolgreich"`
	RecordArchiveMapping []generatorRecordArchiveMapping `xml:"xdomea:Rueckgabeparameter"`
}

type generatorRecordArchiveMapping struct {
	RecordID  string `xml:"xdomea:ID"`
	ArchiveID string `xml:"xdomea:Archivkennung"`
}

type generatorArchivedRecordInfo struct {
	RecordID  string  `xml:"xdomea:IDSGO"`
	Success   bool    `xml:"xdomea:Erfolgreich"`
	ArchiveID *string `xml:"xdomea:Archivkennung"`
}

// Generate0502Message creates the XML code for the appraisal message (code: 0502).
// The generated message has the same xdomea version as the
func Generate0502Message(message db.Message) string {
	xdomeaVersion := XdomeaVersions[message.XdomeaVersion]
	messageHead := GenerateMessageHead0502(message.MessageHead.ProcessID, message.MessageHead.Sender)
	message0502 := generatorMessage0502{
		XdomeaXmlNs: xdomeaVersion.URI,
		XsiXmlNs:    XsiXmlNs,
		MessageHead: messageHead,
	}
	rootRecords := db.FindRootRecords(context.Background(), message.MessageHead.ProcessID, message.MessageType)
	m := appraisableRecords(&rootRecords)
	for id, _ := range m {
		appraisedObject := GenerateAppraisedObject(messageHead.ProcessID, id, xdomeaVersion)
		message0502.AppraisedObjects = append(message0502.AppraisedObjects, appraisedObject)
	}
	xmlBytes, err := xml.MarshalIndent(message0502, " ", " ")
	if err != nil {
		panic("0502 message couldn't be created")
	}
	messageXml := XmlHeader + string(xmlBytes)
	err = ValidateXdomeaXmlString(messageXml, xdomeaVersion)
	if err != nil {
		validationError, ok := err.(xsd.SchemaValidationError)
		if ok {
			for _, e := range validationError.Errors() {
				log.Printf("XML schema error: %s", e.Error())
			}
		}
		panic("generated 0502 message is invalid")
	}
	return messageXml
}

// GenerateAppraisedObject returns xdomea version dependent appraised object.
func GenerateAppraisedObject(
	processID uuid.UUID,
	recordID uuid.UUID,
	xdomeaVersion XdomeaVersion,
) generatorAppraisedObject {
	var appraisedObject generatorAppraisedObject
	a, _ := db.FindAppraisal(processID, recordID)
	if a.Decision != "A" && a.Decision != "V" {
		panic(fmt.Sprintf("called GenerateAppraisedObject with appraisal \"%s\": %v", a.Decision, recordID))
	}

	var objectAppraisal generatorObjectAppraisal
	if isVersionPriorTo300(xdomeaVersion.Code) {
		objectAppraisal = generatorObjectAppraisal{
			AppraisalCodePre300: (*string)(&a.Decision),
		}
	} else {
		appraisalCode := generatorCode{
			Code: string(a.Decision),
		}
		objectAppraisal = generatorObjectAppraisal{
			AppraisalCode: &appraisalCode,
		}
	}
	appraisedObject = generatorAppraisedObject{
		RecordID:        recordID,
		ObjectAppraisal: objectAppraisal,
	}
	return appraisedObject
}

func Generate0504Message(message db.Message) string {
	xdomeaVersion := XdomeaVersions[message.XdomeaVersion]
	messageHead := GenerateMessageHead(message.MessageHead.ProcessID, message.MessageHead.Sender, "0504")
	message0504 := generatorMessage0504{
		XdomeaXmlNs: xdomeaVersion.URI,
		XsiXmlNs:    XsiXmlNs,
		MessageHead: messageHead,
	}
	xmlBytes, err := xml.MarshalIndent(message0504, " ", " ")
	if err != nil {
		panic("0504 message couldn't be created")
	}
	messageXml := XmlHeader + string(xmlBytes)
	err = ValidateXdomeaXmlString(messageXml, xdomeaVersion)
	if err != nil {
		validationError, ok := err.(xsd.SchemaValidationError)
		if ok {
			for _, e := range validationError.Errors() {
				log.Printf("XML schema error: %s", e.Error())
			}
		}
		panic("generated 0504 message is invalid")
	}
	return messageXml
}

func GenerateMessageHead0502(processID uuid.UUID, sender db.Contact) generatorMessageHead0502 {
	messageType := generatorCode{
		Code: "0502",
	}
	timeStamp := time.Now()
	lathContact := GetSenderContact()
	sendingSystem := GetSendingSystem()
	messageHead := generatorMessageHead0502{
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

func GenerateMessageHead(processID uuid.UUID, sender db.Contact, messageCode string) generatorMessageHead {
	messageType := generatorCode{
		Code: messageCode,
	}
	timeStamp := time.Now()
	lathContact := GetSenderContact()
	sendingSystem := GetSendingSystem()
	messageHead := generatorMessageHead{
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
	xdomeaVersion := XdomeaVersions[message.XdomeaVersion]
	messageHead := GenerateMessageHead(message.MessageHead.ProcessID, message.MessageHead.Sender, "0506")
	message0506 := generatorMessage0506{
		XdomeaXmlNs: xdomeaVersion.URI,
		XsiXmlNs:    XsiXmlNs,
		MessageHead: messageHead,
	}
	if isVersionPriorTo300(xdomeaVersion.Code) {
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
	err = ValidateXdomeaXmlString(messageXml, xdomeaVersion)
	if err != nil {
		validationError, ok := err.(xsd.SchemaValidationError)
		if ok {
			for _, e := range validationError.Errors() {
				log.Printf("XML schema error: %s", e.Error())
			}
		}
		panic("generated 0506 message is invalid")
	}
	return messageXml
}

func GetArchivingInfoPre300(archivePackages []db.ArchivePackage) generatorArchivingInfoPre300 {
	info := generatorArchivingInfoPre300{
		Success: true,
	}
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	if archiveTarget == "dimag" {
		var recordArchiveMapping []generatorRecordArchiveMapping
		for _, aip := range archivePackages {
			for _, recordID := range aip.RootRecordIDs {
				idMapping := generatorRecordArchiveMapping{
					RecordID:  recordID.String(),
					ArchiveID: aip.PackageID,
				}
				recordArchiveMapping = append(recordArchiveMapping, idMapping)
			}
		}
		info.RecordArchiveMapping = recordArchiveMapping
	}
	return info
}

func GetArchivedRecordInfo(archivePackages []db.ArchivePackage) []generatorArchivedRecordInfo {
	var info []generatorArchivedRecordInfo
	for _, aip := range archivePackages {
		for _, recordID := range aip.RootRecordIDs {
			info = append(info, GetArchivedRecordIDMapping(recordID.String(), aip))
		}
	}
	return info
}

func GetArchivedRecordIDMapping(recordID string, aip db.ArchivePackage) generatorArchivedRecordInfo {
	archiveTarget := os.Getenv("ARCHIVE_TARGET")
	idMapping := generatorArchivedRecordInfo{
		RecordID: recordID,
		Success:  true,
	}
	if archiveTarget == "dimag" {
		idMapping.ArchiveID = &aip.PackageID
	}
	return idMapping
}

func GetSenderContact() generatorContact {
	institutionName := os.Getenv("INSTITUTION_NAME")
	institutionAbbreviation := os.Getenv("INSTITUTION_ABBREVIATION")
	institution := generatorInstitution{
		Name:         &institutionName,
		Abbreviation: &institutionAbbreviation,
	}
	contact := generatorContact{
		Institution: &institution,
	}
	return contact
}

func ConvertParserToGeneratorContact(contact db.Contact) generatorContact {
	var generatorContact generatorContact
	if contact.Institution != nil {
		institution := generatorInstitution{
			Name:         contact.Institution.Name,
			Abbreviation: contact.Institution.Abbreviation,
		}
		generatorContact.Institution = &institution
	}
	return generatorContact
}

func GetSendingSystem() generatorSendingSystem {
	productName := "X-MAN"
	productVersion := "0.1"
	sendingSystem := generatorSendingSystem{
		ProductName:    &productName,
		ProductVersion: &productVersion,
	}
	return sendingSystem
}

func isVersionPriorTo300(v string) bool {
	return v == "2.3.0" || v == "2.4.0"
}
