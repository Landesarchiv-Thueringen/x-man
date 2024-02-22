package xdomea

import (
	"encoding/xml"
	"errors"
	"lath/xman/internal/db"
	"log"
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
		appraisedObject, err := GenerateAppraisedObject(o, xdomeaVersion)
		if err != nil {
			panic("record object appraisal couldn't be retreived")
		}
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
	o db.AppraisableRecordObject,
	xdomeaVersion db.XdomeaVersion,
) (db.GeneratorAppraisedObject, error) {
	var appraisedObject db.GeneratorAppraisedObject
	appraisal, err := o.GetAppraisal()
	if err != nil {
		return appraisedObject, err
	}
	if appraisal == "B" {
		return appraisedObject, errors.New("appraisal B shouldn't be transmitted")
	}

	var objectAppraisal db.GeneratorObjectAppraisal
	if xdomeaVersion.IsVersionPriorTo300() {
		objectAppraisal = db.GeneratorObjectAppraisal{
			AppraisalCodePre300: &appraisal,
		}
	} else {
		appraisalCode := db.GeneratorCode{
			Code: appraisal,
		}
		objectAppraisal = db.GeneratorObjectAppraisal{
			AppraisalCode: &appraisalCode,
		}
	}
	appraisedObject = db.GeneratorAppraisedObject{
		XdomeaID:        o.GetID(),
		ObjectAppraisal: objectAppraisal,
	}
	return appraisedObject, nil
}

func Generate0504Message(message db.Message) string {
	xdomeaVersion, err := db.GetXdomeaVersionByCode(message.XdomeaVersion)
	if err != nil {
		panic(err)
	}
	messageHead := GenerateMessageHead0504(message.MessageHead.ProcessID, message.MessageHead.Sender)
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

func GenerateMessageHead0504(processID string, sender db.Contact) db.GeneratorMessageHead0504 {
	messageType := db.GeneratorCode{
		Code: "0504",
	}
	timeStamp := time.Now()
	lathContact := GetLAThContact()
	sendingSystem := GetSendingSystem()
	messageHead := db.GeneratorMessageHead0504{
		ProcessID:     processID,
		MessageType:   messageType,
		CreationTime:  timeStamp.Format("2006-01-02T15:04:05"),
		Sender:        lathContact,
		Receiver:      ConvertParserToGeneratorContact(sender),
		SendingSystem: sendingSystem,
	}
	return messageHead
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
