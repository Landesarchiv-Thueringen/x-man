package xdomea

import (
	"errors"
	"fmt"
	"lath/xman/internal/db"
	"lath/xman/internal/format"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
	"github.com/lestrrat-go/libxml2/xsd"
)

var uuidRegexString = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
var Message0501MessageSuffix = "_Aussonderung.Anbieteverzeichnis.0501"
var Message0502MessageSuffix = "_Aussonderung.Bewertungsverzeichnis.0502"
var Message0503MessageSuffix = "_Aussonderung.Aussonderung.0503"
var Message0504MessageSuffix = "_Aussonderung.AnbietungEmpfangBestaetigen.0504"
var Message0505MessageSuffix = "_Aussonderung.BewertungEmpfangBestaetigen.0505"
var message0501RegexString = uuidRegexString + Message0501MessageSuffix + ".zip"
var message0503RegexString = uuidRegexString + Message0503MessageSuffix + ".zip"
var message0505RegexString = uuidRegexString + Message0505MessageSuffix + ".zip"
var uuidRegex = regexp.MustCompile(uuidRegexString)
var message0501Regex = regexp.MustCompile(message0501RegexString)
var message0503Regex = regexp.MustCompile(message0503RegexString)
var message0505Regex = regexp.MustCompile(message0505RegexString)
var namespaceRegex = regexp.MustCompile("^urn:xoev-de:xdomea:schema:([0-9].[0-9].[0=9])$")

func InitMessageTypes() {
	messageTypes := []*db.MessageType{
		{Code: "0000"}, // unknown message type
		{Code: "0501"},
		{Code: "0502"},
		{Code: "0503"},
		{Code: "0504"},
		{Code: "0505"},
		{Code: "0506"},
		{Code: "0507"},
	}
	db.InitMessageTypes(messageTypes)
}

func InitXdomeaVersions() {
	versions := []*db.XdomeaVersion{
		{
			Code:    "2.3.0",
			URI:     "urn:xoev-de:xdomea:schema:2.3.0",
			XSDPath: "xsd/2.3.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
		},
		{
			Code:    "2.4.0",
			URI:     "urn:xoev-de:xdomea:schema:2.4.0",
			XSDPath: "xsd/2.4.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
		},
		{
			Code:    "3.0.0",
			URI:     "urn:xoev-de:xdomea:schema:3.0.0",
			XSDPath: "xsd/3.0.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
		},
		{
			Code:    "3.1.0",
			URI:     "urn:xoev-de:xdomea:schema:3.1.0",
			XSDPath: "xsd/3.1.0/xdomea-Nachrichten-AussonderungDurchfuehren.xsd",
		},
	}
	db.InitXdomeaVersions(versions)
}

func InitRecordObjectAppraisals() {
	appraisals := []*db.RecordObjectAppraisal{
		{Code: "A", ShortDesc: "Archivieren", Desc: "Das Schriftgutobjekt ist archivwürdig."},
		{Code: "B", ShortDesc: "Durchsicht", Desc: "Das Schriftgutobjekt ist zum Bewerten markiert."},
		{Code: "V", ShortDesc: "Vernichten", Desc: "Das Schriftgutobjekt ist zum Vernichten markiert."},
	}
	db.InitRecordObjectAppraisals(appraisals)
}

func InitConfidentialityLevelCodelist() {
	confidentialityLevelCodelist := []*db.ConfidentialityLevel{
		{ID: "001", ShortDesc: "Geheim", Desc: "Geheim: Das Schriftgutobjekt ist als geheim eingestuft."},
		{ID: "002", ShortDesc: "NfD", Desc: "NfD: Das Schriftgutobjekt ist als \"nur für den Dienstgebrauch (nfD)\" eingestuft."},
		{ID: "003", ShortDesc: "Offen", Desc: "Offen: Das Schriftgutobjekt ist nicht eingestuft."},
		{ID: "004", ShortDesc: "Streng geheim", Desc: "Streng geheim: Das Schriftgutobjekt ist als streng geheim eingestuft."},
		{ID: "005", ShortDesc: "Vertraulich", Desc: "Vertraulich: Das Schriftgutobjekt ist als vertraulich eingestuft."},
	}
	db.InitConfidentialityLevelCodelist(confidentialityLevelCodelist)
}

func InitMediumCodelist() {
	mediumCodelist := []*db.Medium{
		{ID: "001", ShortDesc: "Elektronisch", Desc: "Elektronisch: Das Schriftgutobjekt liegt ausschließlich in elektronischer Form vor."},
		{ID: "002", ShortDesc: "Hybrid", Desc: "Hybrid: Das Schriftgutobjekt liegt teilweise in	elektronischer Form und teilweise als Papier vor."},
		{ID: "003", ShortDesc: "Papier", Desc: "Papier: Das Schriftgutobjekt liegt ausschließlich als Papier vor."},
	}
	db.InitMediumCodelist(mediumCodelist)
}

func IsMessage(path string) bool {
	fileName := filepath.Base(path)
	return message0501Regex.MatchString(fileName) ||
		message0503Regex.MatchString(fileName) ||
		message0505Regex.MatchString(fileName)
}

func GetMessageTypeImpliedByPath(path string) (db.MessageType, error) {
	fileName := filepath.Base(path)
	if message0501Regex.MatchString(fileName) {
		return db.GetMessageTypeByCode("0501"), nil
	} else if message0503Regex.MatchString(fileName) {
		return db.GetMessageTypeByCode("0503"), nil
	} else if message0505Regex.MatchString(fileName) {
		return db.GetMessageTypeByCode("0505"), nil
	}
	return db.GetMessageTypeByCode("0000"), errors.New("unknown message: " + path)
}

func GetMessageID(path string) string {
	fileName := filepath.Base(path)
	return uuidRegex.FindString(fileName)
}

func GetMessageName(id string, messageType db.MessageType) string {
	messageSuffix := ""
	switch messageType.Code {
	case "0501":
		messageSuffix = Message0501MessageSuffix
	case "0503":
		messageSuffix = Message0503MessageSuffix
	case "0505":
		messageSuffix = Message0505MessageSuffix
	default:
		log.Fatal("not supported message type")
	}
	return id + messageSuffix + ".xml"
}

func AddMessage(
	agency db.Agency,
	xdomeaID string,
	messageType db.MessageType,
	processStoreDir string,
	messageStoreDir string,
	transferDir string,
	transferDirMessagePath string,
) (db.Process, db.Message, error) {
	var process db.Process
	var message db.Message
	messageName := GetMessageName(xdomeaID, messageType)
	messagePath := path.Join(messageStoreDir, messageName)
	_, err := os.Stat(messagePath)
	if err != nil {
		log.Fatal("message doesn't exist")
	}
	appraisalComplete := messageType.Code != "0501"
	message = db.Message{
		MessageType:            messageType,
		TransferDir:            transferDir,
		TransferDirMessagePath: transferDirMessagePath,
		StoreDir:               messageStoreDir,
		MessagePath:            messagePath,
		AppraisalComplete:      appraisalComplete,
	}
	// xsd schema validation
	err = IsMessageValid(agency, message)
	messageIsValid := err == nil
	message.SchemaValidation = messageIsValid
	if err != nil {
		log.Println(err)
		return process, message, err
	}
	// parse message
	message, err = ParseMessage(message)
	if err != nil {
		log.Fatal(err)
	}
	// store message metadata in database
	process, message, err = db.AddMessage(agency, xdomeaID, processStoreDir, message)
	if err != nil {
		log.Fatal(err)
	}
	compareAgencyFields(agency, message, process)
	if messageType.Code == "0503" {
		// get primary documents
		primaryDocuments, err := db.GetAllPrimaryDocuments(message.ID)
		// error while getting the primary documents should never happen, can't recover
		if err != nil {
			log.Fatal(err)
		}
		err = checkMessage0503Integrity(process, message, primaryDocuments)
		if err == nil {
			// if 0501 message exists, transfer the internal appraisal note from 0501 to 0503 message
			if process.Message0501 != nil {
				err = TransferAppraisalNoteFrom0501To0503(process)
				if err != nil {
					log.Println(err)
					return process, message, err
				}
			}
			// start format verification
			format.VerifyFileFormats(process, message)
		}
	}
	return process, message, nil
}

// IsMessageValid performs a xsd schema validation against the xdomea version of the message.
func IsMessageValid(agency db.Agency, message db.Message) error {
	xdomeaVersion, err := ExtractVersionFromMessage(message)
	if err != nil {
		return err
	}
	err = ValidateXdomeaXmlFile(message.MessagePath, xdomeaVersion)
	if err != nil {
		// Print all schema errors and add error for clearing if a schema validation error occurred.
		validationError, ok := err.(xsd.SchemaValidationError)
		if ok {
			for _, e := range validationError.Errors() {
				log.Printf("error: %s", e.Error())
			}
			processingErr := db.ProcessingError{
				Agency:           agency,
				Description:      "Schema-Validierung ungültig",
				MessageID:        &message.ID,
				TransferDirPath:  &message.TransferDirMessagePath,
				MessageStorePath: &message.MessagePath,
			}
			db.AddProcessingError(processingErr)
		}
		return err
	}
	return nil
}

func checkMessage0503Integrity(
	process db.Process,
	message0503 db.Message,
	primaryDocuments []db.PrimaryDocument,
) error {
	// check if all primary document files exist
	for _, primaryDocument := range primaryDocuments {
		filePath := path.Join(message0503.StoreDir, primaryDocument.FileName)
		_, err := os.Stat(filePath)
		if err != nil {
			log.Println(err.Error())
			processingErr := db.ProcessingError{
				Agency:           process.Agency,
				Description:      "Primärdatei fehlt in Abgabe",
				MessageID:        &message0503.ID,
				TransferDirPath:  &message0503.TransferDirMessagePath,
				MessageStorePath: &message0503.StoreDir,
			}
			db.AddProcessingErrorToProcess(process, processingErr)
			return err
		}
	}
	// check if 0501 message exists
	message0501, err := db.GetMessageOfProcessByCode(process, "0501")
	// 0501 Message doesn't exist. No further message validation necessary.
	if err != nil {
		return nil
	}
	// check if appraisal of 0501 message is already complete
	if !process.ProcessState.Appraisal.Complete {
		errorMessage := "Die Abgabe wurde erhalten, bevor die Bewertung der Anbietung abgeschlossen wurde"
		processingErr := db.ProcessingError{
			Agency:           process.Agency,
			Description:      errorMessage,
			MessageID:        &message0503.ID,
			TransferDirPath:  &message0503.TransferDirMessagePath,
			MessageStorePath: &message0503.StoreDir,
		}
		db.AddProcessingErrorToProcess(process, processingErr)
		return errors.New(errorMessage)
	} else {
		return checkRecordObjectsOfMessage0503(process, message0501, message0503)
	}
}

func checkRecordObjectsOfMessage0503(
	process db.Process,
	message0501 db.Message,
	message0503 db.Message,
) error {
	message0503Incomplete := false
	additionalInfo := ""
	err := checkFileRecordObjectsOfMessage0503(
		message0501.ID,
		message0503.ID,
		&additionalInfo,
	)
	if err != nil {
		message0503Incomplete = true
	}
	err = checkProcessRecordObjectsOfMessage0503(
		message0501.ID,
		message0503.ID,
		&additionalInfo,
	)
	if err != nil {
		message0503Incomplete = true
	}
	if message0503Incomplete {
		errorMessage := "Die Abgabe ist nicht vollständig"
		processingErr := db.ProcessingError{
			Agency:           process.Agency,
			Description:      errorMessage,
			AdditionalInfo:   &additionalInfo,
			MessageID:        &message0503.ID,
			TransferDirPath:  &message0503.TransferDirMessagePath,
			MessageStorePath: &message0503.StoreDir,
		}
		db.AddProcessingErrorToProcess(process, processingErr)
		return errors.New(errorMessage)
	}
	return nil
}

func checkFileRecordObjectsOfMessage0503(
	message0501ID uuid.UUID,
	message0503ID uuid.UUID,
	additionalInfo *string,
) error {
	message0503Incomplete := false
	fileIndex0501, err := db.GetAllFileRecordObjects(message0501ID)
	if err != nil {
		log.Fatal(err)
	}
	fileIndex0503, err := db.GetAllFileRecordObjects(message0503ID)
	if err != nil {
		log.Fatal(err)
	}
	for id0501, file0501 := range fileIndex0501 {
		// missing appraisal metadata for 0501 message, should not happen
		if file0501.ArchiveMetadata == nil || file0501.ArchiveMetadata.AppraisalCode == nil {
			continue
		}
		_, file0503Exists := fileIndex0503[id0501]
		if *file0501.ArchiveMetadata.AppraisalCode == "A" && !file0503Exists {
			errorMessage :=
				"0503 integrity check failed: missing file record object [" + file0501.ID.String() + "]"
			*additionalInfo += "Akte [" + file0501.ID.String() + "] fehlt in Abgabe"
			log.Println(errorMessage)
			message0503Incomplete = true
		}
	}
	if message0503Incomplete {
		return errors.New("0503 message incomplete: file record objects missing")
	}
	return nil
}

func checkProcessRecordObjectsOfMessage0503(
	message0501ID uuid.UUID,
	message0503ID uuid.UUID,
	additionalInfo *string,
) error {
	message0503Incomplete := false
	processIndex0501, err := db.GetAllProcessRecordObjects(message0501ID)
	if err != nil {
		log.Fatal(err)
	}
	processIndex0503, err := db.GetAllProcessRecordObjects(message0503ID)
	if err != nil {
		log.Fatal(err)
	}
	for id0501, process0501 := range processIndex0501 {
		// missing appraisal metadata for 0501 message, should not happen
		if process0501.ArchiveMetadata == nil || process0501.ArchiveMetadata.AppraisalCode == nil {
			continue
		}
		_, process0503Exists := processIndex0503[id0501]
		if *process0501.ArchiveMetadata.AppraisalCode == "A" && !process0503Exists {
			errorMessage := "0503 integrity check failed: missing process record object [" +
				process0501.ID.String() + "]"
			*additionalInfo += "Vorgang [" + process0501.ID.String() + "] fehlt in Abgabe"
			log.Println(errorMessage)
			message0503Incomplete = true
		}
	}
	if message0503Incomplete {
		return errors.New("0503 message incomplete: process record objects missing")
	}
	return nil
}

// compareAgencyFields checks whether the message's metadata match the agency
// and creates a processing error if not.
//
// Only values that are set in `agency` are checked.
func compareAgencyFields(agency db.Agency, message db.Message, process db.Process) {
	a := message.MessageHead.Sender.AgencyIdentification
	if a == nil ||
		(agency.Prefix != "" && a.Prefix == nil) ||
		(agency.Code != "" && a.Code == nil) ||
		(a.Prefix != nil && agency.Prefix != *a.Prefix) ||
		(a.Code != nil && agency.Code != *a.Code) {
		info := ""
		if a != nil && a.Prefix != nil {
			info += fmt.Sprintf("Präfix der Nachricht: %s\n", *a.Prefix)
		} else {
			info += fmt.Sprintf("Präfix der Nachricht: (kein Wert)\n")
		}
		if a != nil && a.Code != nil {
			info += fmt.Sprintf("Behördenschlüssel der Nachricht: %s\n\n", *a.Code)
		} else {
			info += fmt.Sprintf("Behördenschlüssel der Nachricht: (kein Wert)\n\n")
		}
		if agency.Prefix != "" {
			info += fmt.Sprintf("Präfix der konfigurierten abgebenden Stelle: %s\n", agency.Prefix)
		} else {
			info += fmt.Sprintf("Präfix der konfigurierten abgebenden Stelle: (kein Wert)\n")
		}
		if agency.Code != "" {
			info += fmt.Sprintf("Behördenschlüssel der konfigurierten abgebenden Stelle: %s", agency.Code)
		} else {
			info += fmt.Sprintf("Behördenschlüssel der konfigurierten abgebenden Stelle: (kein Wert)")
		}
		processingErr := db.ProcessingError{
			Agency:           process.Agency,
			Description:      "Behördenkennung der Nachricht stimmt nicht mit der konfigurierten abgebenden Stelle überein",
			MessageID:        &message.ID,
			TransferDirPath:  &message.TransferDirMessagePath,
			MessageStorePath: &message.StoreDir,
			AdditionalInfo:   &info,
		}
		db.AddProcessingErrorToProcess(process, processingErr)
	}
}
