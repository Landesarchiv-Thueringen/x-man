package xdomea

import (
	"encoding/xml"
	"errors"
	"lath/xman/internal/db"
	"log"
	"time"
)

const XmlHeader = "<?xml version='1.0' encoding='UTF-8'?>\n"
const XsiXmlNs = "http://www.w3.org/2001/XMLSchema-instance"

func Generate0502Message(message db.Message) string {
	xdomeaVersion, err := db.GetXdomeaVersionByCode(message.XdomeaVersion)
	if err != nil {
		log.Fatal(err)
	}
	messageHead := GenerateMessageHeadLATh(message.MessageHead.ProcessID, message.MessageHead.Sender)
	message0502 := db.GeneratorMessage0502{
		XdomeaXmlNs: xdomeaVersion.URI,
		XsiXmlNs:    XsiXmlNs,
		MessageHead: messageHead,
	}
	// TODO: think about root level process and document record objects
	for _, fileRecordObject := range message.FileRecordObjects {
		for _, o := range fileRecordObject.GetAppraisableObjects() {
			appraisedObject, err := GenerateAppraisedObject(o)
			if err == nil {
				message0502.AppraisedObjects = append(message0502.AppraisedObjects, appraisedObject)
			}
		}
	}
	xmlBytes, err := xml.MarshalIndent(message0502, " ", " ")
	if err != nil {
		log.Fatal("0502 message couldn't be created")
	}
	messageXml := XmlHeader + string(xmlBytes)
	return messageXml
}

func GenerateAppraisedObject(o db.AppraisableRecordObject) (db.GeneratorAppraisedObject, error) {
	var appraisedObject db.GeneratorAppraisedObject
	appraisal, err := o.GetAppraisal()
	if err != nil {
		return appraisedObject, err
	}
	if appraisal == "B" {
		return appraisedObject, errors.New("appraisal B shouldn't be transmitted")
	}
	appraisalCode := db.GeneratorAppraisalCode{
		Code: appraisal,
	}
	objectAppraisal := db.GeneratorObjectAppraisal{
		AppraisalCode: appraisalCode,
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
		log.Fatal(err)
	}
	messageHead := GenerateMessageHeadLATh(message.MessageHead.ProcessID, message.MessageHead.Sender)
	message0504 := db.GeneratorMessage0504{
		XdomeaXmlNs: xdomeaVersion.URI,
		XsiXmlNs:    XsiXmlNs,
		MessageHead: messageHead,
	}
	xmlBytes, err := xml.MarshalIndent(message0504, " ", " ")
	if err != nil {
		log.Fatal("0504 message couldn't be created")
	}
	messageXml := XmlHeader + string(xmlBytes)
	return messageXml
}

func GenerateMessageHeadLATh(processID string, sender db.Contact) db.GeneratorMessageHead {
	timeStamp := time.Now()
	lathContact := GetLAThContact()
	messageHead := db.GeneratorMessageHead{
		ProcessID:    processID,
		CreationTime: timeStamp.Format("01-02-2006 15:04"),
		Sender:       lathContact,
		Receiver:     ConvertParserToGeneratorContact(sender),
	}
	return messageHead
}

func GetLAThContact() db.GeneratorContact {
	institutionName := "Landesarchiv Thüringen"
	institutionAbbrevation := "LATh"
	institution := db.GeneratorInstitution{
		Name:         &institutionName,
		Abbreviation: &institutionAbbrevation,
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
