package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProcessedTransferDirFile struct {
	AgencyID        primitive.ObjectID `bson:"agency_id"`
	TransferDirPath string             `bson:"transfer_dir_path"`
}

// FindProcessedTransferDirFile returns a map of processed paths for the given
// agency. A path mapped to true indicates that the file either
// - has already been processed, or
// - is currently being processed.
func FindProcessedTransferDirFiles(agencyID primitive.ObjectID) map[string]bool {
	coll := mongoDatabase.Collection("transfer_dir_files")
	filter := bson.D{{"agency_id", agencyID}}
	var files []ProcessedTransferDirFile
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	cursor.All(context.Background(), &files)
	m := make(map[string]bool)
	for _, file := range files {
		m[file.TransferDirPath] = true
	}
	return m
}

// InsertProcessedTransferDirFile marks a file in a transfer directory as
// already processed. This file will not be processed again until the entry for
// the file is removed.
func InsertProcessedTransferDirFile(agencyID primitive.ObjectID, transferDirPath string) {
	coll := mongoDatabase.Collection("transfer_dir_files")
	_, err := coll.InsertOne(context.Background(), ProcessedTransferDirFile{
		AgencyID: agencyID, TransferDirPath: transferDirPath,
	})
	if err != nil {
		panic(err)
	}
}

func DeleteProcessedTransferDirFile(agencyID primitive.ObjectID, transferDirPath string) (ok bool) {
	coll := mongoDatabase.Collection("transfer_dir_files")
	filter := bson.D{{"agency_id", agencyID}, {"transfer_dir_path", transferDirPath}}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount == 0 {
		return false
	}
	return true
}
