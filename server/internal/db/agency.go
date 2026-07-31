package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Agency represents an institution as configured in the administration panel.
//
// It maps a transfer directory to assigned users and an archive collection.
//
// All messages that are received via the configured transfer directory are
// considered to belong the the configured institution, ignoring the content of
// the "sender" field.
type Agency struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SchemaVersion int                `bson:"schema_version" json:"_"`
	Name          string             `json:"name"`
	Abbreviation  string             `json:"abbreviation"`
	// Prefix is the agency prefix as by xdomea.
	Prefix string `json:"prefix"`
	// Code is the agency code as by xdomea.
	Code string `json:"code"`
	// ContactEmail is the e-mail address to use to contact the agency.
	ContactEmail string `bson:"contact_email" json:"contactEmail"`
	// Users are users responsible for processes of this Agency.
	Users        []string            `json:"users"`
	CollectionID *primitive.ObjectID `bson:"collection_id" json:"collectionId"`
	TransferDir  TransferDir         `json:"transferDir"`
}

type TransferProtocol string

const (
	ProtocolFile         TransferProtocol = "file"
	ProtocolWebDAV       TransferProtocol = "dav"
	ProtocolWebDAVSecure TransferProtocol = "davs"
)

// TransferDir contains all information to receive and send messages to the transfer directory.
// All paths should be stored without trailing or leading slashes.
// File systems and webDAVs have different requirements for the paths.
type TransferDir struct {
	Protocol         TransferProtocol `json:"protocol"`
	Host             string           `json:"host"`
	Path             string           `json:"path"`
	User             string           `json:"user"`
	Password         string           `json:"password"`
	Path0502         string           `json:"path0502"`
	Path0504         string           `json:"path0504"`
	Path0506         string           `json:"path0506"`
	Path0507         string           `json:"path0507"`
	AllowInsecureTLS bool             `json:"allowInsecureTLS"`
}

func FindAgencies(ctx context.Context) []Agency {
	return findAgencies(ctx, bson.D{{}})
}

func FindAgenciesForUser(ctx context.Context, userID string) []Agency {
	return findAgencies(ctx, bson.D{{"users", bson.D{{"$all", bson.A{userID}}}}})
}

func FindAgenciesForCollection(ctx context.Context, collectionID primitive.ObjectID) []Agency {
	return findAgencies(ctx, bson.D{{"collection_id", collectionID}})
}

func findAgencies(ctx context.Context, filter bson.D) []Agency {
	coll := mongoDatabase.Collection("agencies")
	var agencies []Agency
	cursor, err := coll.Find(ctx, filter)
	handleError(ctx, err)
	err = cursor.All(ctx, &agencies)
	handleError(ctx, err)
	return agencies
}

func InsertAgency(agency Agency) (id primitive.ObjectID) {
	coll := mongoDatabase.Collection("agencies")
	result, err := coll.InsertOne(context.Background(), agency)
	if err != nil {
		panic(err)
	}
	return result.InsertedID.(primitive.ObjectID)
}

func ReplaceAgency(agency Agency) (ok bool) {
	coll := mongoDatabase.Collection("agencies")
	filter := bson.D{{"_id", agency.ID}}
	result, err := coll.ReplaceOne(context.Background(), filter, agency)
	if err != nil {
		panic(err)
	}
	if result.MatchedCount == 0 {
		return false
	}
	updateAgencyForProcesses(agency)
	updateAgencyForProcessingErrors(agency)
	return true
}

func DeleteAgency(id primitive.ObjectID) (ok bool) {
	coll := mongoDatabase.Collection("agencies")
	filter := bson.D{{"_id", id}}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount == 0 {
		return false
	}
	return true
}
