package db

import (
	"context"
	"encoding/xml"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type StoragePaths struct {
	// TransferDirPath is the path of the message's transfer file within the
	// transfer directory.
	TransferDirPath string `bson:"transfer_dir_path"`
	// StoreDir is root storage directory used for the storing files that
	// belong to the message.
	StoreDir string `bson:"store_dir"`
	// MessagePath is the path of the message's root file within the storage
	// directory.
	MessagePath string `bson:"message_path"`
}

type MessageType string

const (
	MessageType0501 MessageType = "0501"
	MessageType0502 MessageType = "0502"
	MessageType0503 MessageType = "0503"
	MessageType0504 MessageType = "0504"
	MessageType0505 MessageType = "0505"
	MessageType0506 MessageType = "0506"
	MessageType0507 MessageType = "0507"
)

type Message struct {
	StoragePaths   `bson:"storage_paths" json:"-"`
	XdomeaVersion  string      `bson:"xdomea_version" json:"xdomeaVersion"`
	MessageType    MessageType `bson:"message_type" json:"messageType"`
	MessageHead    MessageHead `bson:"message_head" json:"messageHead"`
	MaxRecordDepth int         `bson:"max_record_depth" json:"maxRecordDepth"`
}

type MessageHead struct {
	XMLName      xml.Name  `xml:"Kopf" bson:"-" json:"-"`
	ProcessID    uuid.UUID `xml:"ProzessID" bson:"process_id" json:"processID"`
	CreationTime string    `xml:"Erstellungszeitpunkt" bson:"creation_time" json:"creationTime"`
	Sender       Contact   `xml:"Absender" json:"sender"`
	Receiver     Contact   `xml:"Empfaenger" json:"receiver"`
}

type Contact struct {
	AgencyIdentification *AgencyIdentification `xml:"Behoerdenkennung" bson:"agency_identification" json:"agencyIdentification"`
	Institution          *Institution          `xml:"Institution" bson:"institution" json:"institution"`
}

type AgencyIdentification struct {
	Code   string `xml:"Behoerdenschluessel>code" json:"code"`
	Prefix string `xml:"Praefix>code" json:"prefix"`
}

type Institution struct {
	Name         string `xml:"Name"  json:"name"`
	Abbreviation string `xml:"Kurzbezeichnung" json:"abbreviation"`
}

// InsertMessage adds xdomea message to collection.
func InsertMessage(message Message) {
	coll := mongoDatabase.Collection("messages")
	_, err := coll.InsertOne(context.Background(), message)
	if err != nil {
		panic(err)
	}
}

// FindMessagesForProcess returns all messages for the given submission process.
// It returns an empty array, if there is no matching submission process.
func FindMessagesForProcess(ctx context.Context, processID uuid.UUID) []Message {
	coll := mongoDatabase.Collection("messages")
	filter := bson.D{{"message_head.process_id", processID}}
	var messages []Message
	cursor, err := coll.Find(ctx, filter)
	handleError(ctx, err)
	err = cursor.All(ctx, &messages)
	handleError(ctx, err)
	return messages
}

// FindMessage returns the message of the given type for the given process.
func FindMessage(
	ctx context.Context,
	processID uuid.UUID,
	messageType MessageType,
) (message Message, found bool) {
	message, err := findMessage(ctx, processID, messageType)
	return message, handleError(ctx, err)
}

// TryFindMessage is like FindMessage, but doesn't panic.
func TryFindMessage(
	ctx context.Context,
	processID uuid.UUID,
	messageType MessageType,
) (message Message, found bool) {
	message, err := findMessage(ctx, processID, messageType)
	return message, err == nil
}

func findMessage(
	ctx context.Context,
	processID uuid.UUID,
	messageType MessageType,
) (message Message, err error) {
	coll := mongoDatabase.Collection("messages")
	filter := bson.D{
		{"message_head.process_id", processID},
		{"message_type", messageType},
	}
	err = coll.FindOne(ctx, filter).Decode(&message)
	return message, err
}

// DeleteMessage deletes the given message.
//
// Do not use directly, instead use `xdomea.DeleteMessage`.
func DeleteMessage(message Message) (ok bool) {
	coll := mongoDatabase.Collection("messages")
	filter := bson.D{
		{"message_head.process_id", message.MessageHead.ProcessID},
		{"message_type", message.MessageType},
	}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount == 0 {
		return false
	}
	return true
}

func DeleteMessagesForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("messages")
	filter := bson.D{
		{"message_head.process_id", processID},
	}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}
