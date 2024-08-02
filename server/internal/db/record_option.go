package db

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PackagingOption is the user-selectable instruction of how to package a root
// record and its sub records. A packaging option always refers to a root
// record.
type PackagingOption string

const (
	// Create a single package for the root record.
	PackagingOptionRoot PackagingOption = "root"
	// Create a package for each subfile and process of the root record.
	// Remaining documents will be packaged as a single package.
	PackagingOptionLevel1 PackagingOption = "level-1"
	// For each subfile of the root record, create a package for each contained
	// subfile and process. Also create a package for each process of the root
	// record. Remaining documents will be packaged as a single package per
	// (sub)file.
	PackagingOptionLevel2 PackagingOption = "level-2"
)

// TODO: Consider renaming the table to "packaging" if there are no additional
// uses or merge with appraisals.

type RecordOption struct {
	ProcessID uuid.UUID `bson:"process_id" json:"-"`
	RecordID  uuid.UUID `bson:"record_id" json:"recordId"`
	// Packaging is the packaging option selected by the user. It affects
	// packaging of the given record and its sub records. Its value will be
	// ignored, if the given record is already part of another package.
	Packaging PackagingOption `bson:"packaging" json:"packaging"`
}

func FindRecordOptionsForProcess(ctx context.Context, processID uuid.UUID) []RecordOption {
	coll := mongoDatabase.Collection("record_options")
	filter := bson.D{{"process_id", processID}}
	cursor, err := coll.Find(ctx, filter)
	handleError(ctx, err)
	var o []RecordOption
	err = cursor.All(ctx, &o)
	handleError(ctx, err)
	return o
}

func UpsertPackaging(
	processID uuid.UUID,
	recordID uuid.UUID,
	packaging PackagingOption,
) {
	upsertRecordOption(processID, recordID, bson.D{
		{"packaging", packaging},
	})
}

func upsertRecordOption(
	processID uuid.UUID,
	recordID uuid.UUID,
	setItems bson.D,
) {
	coll := mongoDatabase.Collection("record_options")
	filter := bson.D{
		{"process_id", processID},
		{"record_id", recordID},
	}
	update := bson.D{{"$set", append(bson.D{
		{"process_id", processID},
		{"record_id", recordID},
	}, setItems...)}}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		panic(err)
	}
}

func DeleteRecordOptionsForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("record_options")
	filter := bson.D{{"process_id", processID}}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}
