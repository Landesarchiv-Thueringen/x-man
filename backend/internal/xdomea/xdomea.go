package xdomea

import (
	"errors"
	"lath/xdomea/internal/db"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

var uuidRegexString = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"
var Message0501MessageSuffix = "_Aussonderung.Anbieteverzeichnis.0501"
var Message0502MessageSuffix = "_Aussonderung.Bewertungsverzeichnis.0502"
var Message0503MessageSuffix = "_Aussonderung.Aussonderung.0503"
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
		{Code: "2.3.0", URI: "urn:xoev-de:xdomea:schema:2.3.0"},
		{Code: "2.4.0", URI: "urn:xoev-de:xdomea:schema:2.4.0"},
		{Code: "3.0.0", URI: "urn:xoev-de:xdomea:schema:3.0.0"},
		{Code: "3.1.0", URI: "urn:xoev-de:xdomea:schema:3.1.0"},
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
) {
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
	message := db.Message{
		MessageType:       messageType,
		TransferDir:       transferDir,
		StoreDir:          messageStoreDir,
		MessagePath:       messagePath,
		AppraisalComplete: appraisalComplete,
	}
	// TODO: xsd validation
	message = ParseMessage(message)
	_, err = db.AddMessage(xdomeaID, processStoreDir, message)
	if err != nil {
		log.Fatal(err)
	}
}
