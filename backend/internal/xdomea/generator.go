package xdomea

import (
	"encoding/xml"
	"errors"
	"lath/xdomea/internal/db"
	"log"
)

func Generate0502Message(message db.Message) {
	lathContact := GetLAThContact()
	messageHead := db.GeneratorMessageHead{
		ProcessID: message.MessageHead.ProcessID,
		Sender:    lathContact,
		Receiver:  ConvertParserToGeneratorContact(message.MessageHead.Sender),
	}
	message0502 := db.GeneratorMessage0502{
		XdomeaXmlNs: "urn:xoev-de:xdomea:schema:3.0.0",
		MessageHead: messageHead,
	}
	for _, r := range message.RecordObjects {
		if r.FileRecordObject != nil {
			appraisedObject, err := GenerateAppraisedObject(*r.FileRecordObject)
			if err == nil {
				message0502.AppraisedObjects = append(message0502.AppraisedObjects, appraisedObject)
			}
		}
	}
	out, err := xml.MarshalIndent(message0502, " ", " ")
	if err != nil {
		log.Fatal("0502 message couldn't be created")
	}
	log.Println(string(out))
}

func GenerateAppraisedObject(fileRecordObject db.FileRecordObject) (db.GeneratorAppraisedObject, error) {
	var appraisedObject db.GeneratorAppraisedObject
	if fileRecordObject.ArchiveMetadata != nil &&
		fileRecordObject.ArchiveMetadata.AppraisalCode != nil {
		appraisalCode := db.GeneratorAppraisalCode{
			Code: *fileRecordObject.ArchiveMetadata.AppraisalCode,
		}
		objectAppraisal := db.GeneratorObjectAppraisal{
			AppraisalCode: appraisalCode,
		}
		appraisedObject = db.GeneratorAppraisedObject{
			XdomeaID:        fileRecordObject.ID,
			ObjectAppraisal: objectAppraisal,
		}
		return appraisedObject, nil
	}
	return appraisedObject, errors.New("no appraisal existing")
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
