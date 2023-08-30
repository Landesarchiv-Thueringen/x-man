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
var message0501MessageSuffix = "_Aussonderung.Anbieteverzeichnis.0501"
var message0503MessageSuffix = "_Aussonderung.Aussonderung.0503"
var message0501RegexString = uuidRegexString + message0501MessageSuffix + ".zip"
var message0503RegexString = uuidRegexString + message0503MessageSuffix + ".zip"
var uuidRegex = regexp.MustCompile(uuidRegexString)
var message0501Regex = regexp.MustCompile(message0501RegexString)
var message0503Regex = regexp.MustCompile(message0503RegexString)

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

func IsMessage(path string) bool {
	fileName := filepath.Base(path)
	return message0501Regex.MatchString(fileName) || message0503Regex.MatchString(fileName)
}

func GetMessageTypeImpliedByPath(path string) (db.MessageType, error) {
	fileName := filepath.Base(path)
	log.Println(path)
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
		messageSuffix = message0501MessageSuffix
	case "0503":
		messageSuffix = message0503MessageSuffix
	default:
		log.Fatal("not supported message type")
	}
	return id + messageSuffix + ".xml"
}

func AddMessage(xdomeaID string, messageType db.MessageType, storePath string) {
	log.Println(xdomeaID)
	log.Println(messageType.Code)
	log.Println(storePath)
	messageName := GetMessageName(xdomeaID, messageType)
	log.Println(messageName)
	messagePath := path.Join(storePath, messageName)
	_, err := os.Stat(messagePath)
	if err != nil {
		log.Fatal("message doesn't exist")
	}
	message := db.Message{MessageType: messageType, StoreDir: storePath, MessagePath: messagePath}
	_, err = db.AddMessage(xdomeaID, message)
	if err != nil {
		log.Fatal(err)
	}
}
