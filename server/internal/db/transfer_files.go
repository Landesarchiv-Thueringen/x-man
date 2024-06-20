package db

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransferFile struct {
	AgencyID  primitive.ObjectID `bson:"agency_id"`
	ProcessID uuid.UUID          `bson:"process_id"`
	Path      string             `bson:"path"`
}

// FindProcessedTransferDirFile returns a map of processed paths for the given
// agency. A path mapped to true indicates that the file either
// - has already been processed, or
// - is currently being processed.
func FindTransferDirFilesForAgency(agencyID primitive.ObjectID) []TransferFile {
	coll := mongoDatabase.Collection("transfer_files")
	filter := bson.D{{"agency_id", agencyID}}
	var files []TransferFile
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	err = cursor.All(context.Background(), &files)
	if err != nil {
		panic(err)
	}
	return files
}

func FindTransferDirFilesForProcess(processID uuid.UUID) []TransferFile {
	coll := mongoDatabase.Collection("transfer_files")
	filter := bson.D{{"process_id", processID}}
	var files []TransferFile
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	err = cursor.All(context.Background(), &files)
	if err != nil {
		panic(err)
	}
	return files
}

// InsertTransferFile marks a file in a transfer directory as
// already processed. This file will not be processed again until the entry for
// the file is removed.
func InsertTransferFile(agencyID primitive.ObjectID, processID uuid.UUID, path string) {
	coll := mongoDatabase.Collection("transfer_files")
	_, err := coll.InsertOne(context.Background(), TransferFile{
		AgencyID:  agencyID,
		ProcessID: processID,
		Path:      path,
	})
	if err != nil {
		panic(err)
	}
}

func DeleteTransferFile(agencyID primitive.ObjectID, path string) (ok bool) {
	coll := mongoDatabase.Collection("transfer_files")
	filter := bson.D{{"agency_id", agencyID}, {"path", path}}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount == 0 {
		return false
	}
	return true
}
