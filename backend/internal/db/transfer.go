package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TransferMode string

const (
	Local     TransferMode = "local"
	WebDAV    TransferMode = "dav"
	WebDAVSec TransferMode = "davs"
)

type TransferDir struct {
	TransferMode TransferMode `bson:"transfer_mode" json:"transferMode"`
	Host         *string      `bson:"host" json:"host"`
	Path         string       `bson:"path" json:"path"`
}

type Sender struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string        `bson:"name" json:"name"`
	Abbreviation string        `bson:"abbreviation" json:"abbreviation"`
	TransferDir  TransferDir   `bson:"transfer_dir" json:"transferDir"`
}

func InsertSender(sender Sender) (bson.ObjectID, error) {
	coll := db.Collection("senders")
	result, err := coll.InsertOne(context.Background(), sender)
	if err != nil {
		log.Println(err)
		return bson.ObjectID{}, ErrDbInsert
	}
	return result.InsertedID.(bson.ObjectID), nil
}

func FindSenders(ctx context.Context) ([]Sender, error) {
	return findSenders(ctx, bson.D{{}})
}

func findSenders(ctx context.Context, filter bson.D) ([]Sender, error) {
	coll := db.Collection("senders")
	var senders []Sender
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		log.Println(err)
		return senders, ErrDbFind
	}
	err = cursor.All(ctx, &senders)
	if err != nil {
		log.Println(err)
		return senders, ErrDbFind
	}
	return senders, nil
}
