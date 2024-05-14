package db

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ArchivePackage struct {
	ProcessID uuid.UUID `bson:"process_id"`
	// CollectionID references the DIMAG collection (de: Bestand)
	CollectionID primitive.ObjectID `bson:"collection_id"`
	// IOTitle is the title of the information object in DIMAG.
	IOTitle string `bson:"io_title"`
	// IOLifetime is the combined lifetime begin and end of the information
	// object in DIMAG.
	IOLifetime string `bson:"io_lifetime"`
	// REPTitle is the title of the representation in DIMAG.
	REPTitle string `bson:"rep_title"`
	// RootRecordIDs are the RecordIDs of all root-level records contained in
	// the archive package.
	RootRecordIDs []uuid.UUID `bson:"root_record_ids"`
	// PrimaryDocuments are all primary documents contained in the archive
	// package.
	PrimaryDocuments []PrimaryDocument `bson:"primary_documents"`
	// PackageID is the ID assigned by DIMAG when importing the package.
	PackageID string `bson:"package_id"`
}

func InsertArchivePackage(aip ArchivePackage) {
	coll := mongoDatabase.Collection("archive_packages")
	_, err := coll.InsertOne(context.Background(), aip)
	if err != nil {
		panic(err)
	}
}

func FindArchivePackagesForProcess(ctx context.Context, processID uuid.UUID) []ArchivePackage {
	coll := mongoDatabase.Collection("archive_packages")
	filter := bson.D{{"process_id", processID}}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	var aips []ArchivePackage
	err = cursor.All(ctx, &aips)
	if err != nil {
		panic(err)
	}
	return aips
}

func DeleteArchivePackagesForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("archive_packages")
	filter := bson.D{{"process_id", processID}}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}
