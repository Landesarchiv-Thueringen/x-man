package xdomea

import (
	"errors"
	"io/ioutil"
	"lath/xman/internal/db"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/xsd"
)

var uuidRegexString = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
var Message0501MessageSuffix = "_Aussonderung.Anbieteverzeichnis.0501"
var Message0502MessageSuffix = "_Aussonderung.Bewertungsverzeichnis.0502"
var Message0503MessageSuffix = "_Aussonderung.Aussonderung.0503"
var Message0504MessageSuffix = "_Aussonderung.AnbietungEmpfangBestaetigen.0504"
var message0501RegexString = uuidRegexString + Message0501MessageSuffix + ".zip"
var message0503RegexString = uuidRegexString + Message0503MessageSuffix + ".zip"
var uuidRegex = regexp.MustCompile(uuidRegexString)
var message0501Regex = regexp.MustCompile(message0501RegexString)
var message0503Regex = regexp.MustCompile(message0503RegexString)
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

func InitRecordObjectConfidentialities() {
	confidentialities := []*db.RecordObjectConfidentiality{
		{Code: "001", Desc: "Geheim: Das Schriftgutobjekt ist als geheim eingestuft."},
		{Code: "002", Desc: "NfD: Das Schriftgutobjekt ist als \"nur für den Dienstgebrauch (nfD)\" eingestuft."},
		{Code: "003", Desc: "Offen: Das Schriftgutobjekt ist nicht eingestuft."},
		{Code: "004", Desc: "Streng geheim: Das Schriftgutobjekt ist als streng geheim eingestuft."},
		{Code: "005", Desc: "Vertraulich: Das Schriftgutobjekt ist als vertraulich eingestuft."},
	}
	db.InitRecordObjectConfidentialities(confidentialities)
}

func IsMessage(path string) bool {
	fileName := filepath.Base(path)
	return message0501Regex.MatchString(fileName) || message0503Regex.MatchString(fileName)
}

func GetMessageTypeImpliedByPath(path string) (db.MessageType, error) {
	fileName := filepath.Base(path)
	if message0501Regex.MatchString(fileName) {
		return db.GetMessageTypeByCode("0501"), nil
	} else if message0503Regex.MatchString(fileName) {
		return db.GetMessageTypeByCode("0503"), nil
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
	default:
		log.Fatal("not supported message type")
	}
	return id + messageSuffix + ".xml"
}

func AddMessage(
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
	appraisalComplete := false
	if messageType.Code == "0503" {
		appraisalComplete = true
	}
	message = db.Message{
		MessageType:            messageType,
		TransferDir:            transferDir,
		TransferDirMessagePath: transferDirMessagePath,
		StoreDir:               messageStoreDir,
		MessagePath:            messagePath,
		AppraisalComplete:      appraisalComplete,
	}
	// xsd schema validation
	err = IsMessageValid(message)
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
	process, message, err = db.AddMessage(xdomeaID, processStoreDir, message)
	if err != nil {
		log.Fatal(err)
	}
	if messageType.Code == "0503" {
		checkMessage0503Integrity(process, message)
	}
	return process, message, nil
}

// performs xsd schema validation
func IsMessageValid(message db.Message) error {
	xdomeaVersion, err := ExtractVersionFromMessage(message)
	if err != nil {
		return err
	}
	schema, err := xsd.ParseFromFile(xdomeaVersion.XSDPath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer schema.Free()
	messageFile, err := os.Open(message.MessagePath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer messageFile.Close()
	buffer, err := ioutil.ReadAll(messageFile)
	if err != nil {
		log.Println(err)
		return err
	}
	messageXML, err := libxml2.Parse(buffer)
	if err != nil {
		return err
	}
	defer messageXML.Free()
	err = schema.Validate(messageXML)
	if err != nil {
		for _, e := range err.(xsd.SchemaValidationError).Errors() {
			log.Printf("error: %s", e.Error())
		}
		processingErr := db.ProcessingError{
			Description:      "Schema-Validierung ungültig",
			MessageID:        &message.ID,
			TransferDirPath:  &message.TransferDirMessagePath,
			MessageStorePath: &message.MessagePath,
		}
		db.AddProcessingError(processingErr)
		return err
	}
	return nil
}

func checkMessage0503Integrity(process db.Process, message0503 db.Message) {
	primaryDocuments, err := db.GetAllPrimaryDocuments(message0503.ID)
	// error while getting the primary documents should never happen, can't recover
	if err != nil {
		log.Fatal(err)
	}
	// check if all primary document files exist
	for _, primaryDocument := range primaryDocuments {
		filePath := path.Join(message0503.StoreDir, primaryDocument.FileName)
		_, err := os.Stat(filePath)
		if err != nil {
			log.Println(err.Error())
			processingErr := db.ProcessingError{
				Description:      "Primärdatei fehlt in Abgabe",
				MessageID:        &message0503.ID,
				TransferDirPath:  &message0503.TransferDirMessagePath,
				MessageStorePath: &message0503.StoreDir,
			}
			db.AddProcessingErrorToProcess(process, processingErr)
			return
		}
	}
	// check if 0501 message exists
	message0501, err := db.GetMessageOfProcessByCode(process, "0501")
	if err != nil {
		processingErr := db.ProcessingError{
			Description:      "es existiert keine Anbietung für die Abgabe",
			MessageID:        &message0503.ID,
			TransferDirPath:  &message0503.TransferDirMessagePath,
			MessageStorePath: &message0503.StoreDir,
		}
		db.AddProcessingErrorToProcess(process, processingErr)
		return
	}
	// check if appraisal of 0501 message is already complete
	if !message0501.AppraisalComplete {
		processingErr := db.ProcessingError{
			Description:      "Abgabe erhalten, bevor die Bewertung der Anbietung abgeschlossen wurde",
			MessageID:        &message0503.ID,
			TransferDirPath:  &message0503.TransferDirMessagePath,
			MessageStorePath: &message0503.StoreDir,
		}
		db.AddProcessingErrorToProcess(process, processingErr)
		return
	} else {
		checkRecordObjetcsOfMessage0503(process, message0501, message0503)
	}
}

func checkRecordObjetcsOfMessage0503(
	process db.Process,
	message0501 db.Message,
	message0503 db.Message,
) {
	message0503Incomplete := false
	additionalInfo := ""
	err := checkFileRecordObjetcsOfMessage0503(
		message0501.ID,
		message0503.ID,
		&additionalInfo,
	)
	if err != nil {
		message0503Incomplete = true
	}
	err = checkProcessRecordObjetcsOfMessage0503(
		message0501.ID,
		message0503.ID,
		&additionalInfo,
	)
	if err != nil {
		message0503Incomplete = true
	}
	if message0503Incomplete {
		processingErr := db.ProcessingError{
			Description:      "Abgabe ist nicht vollständig",
			AdditionalInfo:   &additionalInfo,
			MessageID:        &message0503.ID,
			TransferDirPath:  &message0503.TransferDirMessagePath,
			MessageStorePath: &message0503.StoreDir,
		}
		db.AddProcessingErrorToProcess(process, processingErr)
		return
	}
}

func checkFileRecordObjetcsOfMessage0503(
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

func checkProcessRecordObjetcsOfMessage0503(
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
