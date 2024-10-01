package db

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PackagingChoice is the user-selectable instruction of how to package a root
// record and its sub records. A PackagingChoice always refers to a root record.
type PackagingChoice string

const (
	// Create a single package for the root record.
	PackagingChoiceRoot PackagingChoice = "root"
	// Create a package for each subfile and process of the root record.
	// Remaining documents will be packaged as a single package.
	PackagingChoiceLevel1 PackagingChoice = "level-1"
	// For each subfile of the root record, create a package for each contained
	// subfile and process. Also create a package for each process of the root
	// record. Remaining documents will be packaged as a single package per
	// (sub)file.
	PackagingChoiceLevel2 PackagingChoice = "level-2"
)

// PackagingRecord is a database entry that describes a user's packaging choice
// for a given root record.
type PackagingRecord struct {
	ProcessID uuid.UUID `bson:"process_id" json:"-"`
	RecordID  uuid.UUID `bson:"record_id" json:"recordId"`
	// PackagingChoice is the packaging option selected by the user. It affects
	// packaging of the given record and its sub records.
	PackagingChoice PackagingChoice `bson:"packaging_choice" json:"packagingChoice"`
}

func FindPackagingChoicesForProcess(ctx context.Context, processID uuid.UUID) []PackagingRecord {
	coll := mongoDatabase.Collection("packaging_choices")
	filter := bson.D{{"process_id", processID}}
	cursor, err := coll.Find(ctx, filter)
	handleError(ctx, err)
	var o []PackagingRecord
	err = cursor.All(ctx, &o)
	handleError(ctx, err)
	return o
}

func UpsertPackagingChoice(
	processID uuid.UUID,
	recordID uuid.UUID,
	packagingChoice PackagingChoice,
) {
	coll := mongoDatabase.Collection("packaging_choices")
	filter := bson.D{
		{"process_id", processID},
		{"record_id", recordID},
	}
	update := bson.D{{"$set", bson.D{
		{"process_id", processID},
		{"record_id", recordID},
		{"packaging_choice", packagingChoice},
	}}}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		panic(err)
	}
}

func DeletePackagingChoicesForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("packaging_choices")
	filter := bson.D{{"process_id", processID}}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}
