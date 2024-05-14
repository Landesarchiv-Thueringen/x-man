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
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `json:"name"`
	Abbreviation string             `json:"abbreviation"`
	// Prefix is the agency prefix as by xdomea.
	Prefix string `json:"prefix"`
	// Code is the agency code as by xdomea.
	Code string `json:"code"`
	// ContactEmail is the e-mail address to use to contact the agency.
	ContactEmail string `bson:"contact_email" json:"contactEmail"`
	// TransferDirURL contains the protocol, host, username and password needed to access a file share.
	// Possible values for the protocol are file, webdav, webdavs.
	// The username and password are optional.
	TransferDirURL string `json:"transferDirURL"`
	// Users are users responsible for processes of this Agency.
	Users        []string            `json:"users"`
	CollectionID *primitive.ObjectID `bson:"collection_id" json:"collectionId"`
}

func FindAgency(ctx context.Context, id primitive.ObjectID) Agency {
	coll := mongoDatabase.Collection("agencies")
	filter := bson.D{{"_id", id}}
	var agency Agency
	err := coll.FindOne(ctx, filter).Decode(&agency)
	if err != nil {
		panic(err)
	}
	return agency
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
	if err != nil {
		panic(err)
	}
	err = cursor.All(ctx, &agencies)
	if err != nil {
		panic(err)
	}
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

func ReplaceAgency(agency Agency) {
	coll := mongoDatabase.Collection("agencies")
	filter := bson.D{{"_id", agency.ID}}
	_, err := coll.ReplaceOne(context.Background(), filter, agency)
	if err != nil {
		panic(err)
	}
	updateAgencyForProcesses(agency)
	updateAgencyForProcessingErrors(agency)
}

func DeleteAgency(id primitive.ObjectID) {
	coll := mongoDatabase.Collection("agencies")
	filter := bson.D{{"_id", id}}
	_, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}
