package db

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ArchivePackage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ProcessID uuid.UUID          `bson:"process_id"`
	// CollectionID references the DIMAG collection (de: Bestand)
	CollectionID primitive.ObjectID `bson:"collection_id"`
	// IOTitle is the title of the information object in DIMAG.
	IOTitle string `bson:"io_title"`
	// IOLifetime is the lifetime of the information object in DIMAG.
	IOLifetime *Lifetime `bson:"io_lifetime"`
	// REPTitle is the title of the representation in DIMAG.
	REPTitle string `bson:"rep_title"`
	// RecordPath represents the path from the message root to the records
	// contained in the archive package (given by RecordIDs).
	//
	// For root-level records, RecordPath is empty.
	//
	// For sub records, the first element of RecordPath is a root-level file
	// record, the second (if any) is a sub-file record of the record referenced
	// by the first element and so forth. Records referenced by RecordIDs are
	// sub records of the record referenced by the last element of RecordPath.
	//
	// All segments given by RecordPath must reference file or sub-file records.
	RecordPath []uuid.UUID `bson:"record_path"`
	// RecordIDs are the RecordIDs of all records contained in the archive
	// package.
	RecordIDs []uuid.UUID `bson:"record_ids"`
	// PrimaryDocuments are all primary documents contained in the archive
	// package.
	PrimaryDocuments []PrimaryDocumentContext `bson:"primary_documents"`
	// PackageID is the ID assigned by DIMAG when importing the package.
	PackageID string `bson:"package_id"`
}

func InsertArchivePackage(aip *ArchivePackage) {
	coll := mongoDatabase.Collection("archive_packages")
	result, err := coll.InsertOne(context.Background(), aip)
	if err != nil {
		panic(err)
	}
	aip.ID = result.InsertedID.(primitive.ObjectID)
}

func ReplaceArchivePackage(aip *ArchivePackage) (ok bool) {
	coll := mongoDatabase.Collection("archive_packages")
	filter := bson.D{{"_id", aip.ID}}
	result, err := coll.ReplaceOne(context.Background(), filter, aip)
	if err != nil {
		panic(err)
	}
	return result.MatchedCount == 1
}

func FindArchivePackagesForProcess(ctx context.Context, processID uuid.UUID) []ArchivePackage {
	coll := mongoDatabase.Collection("archive_packages")
	filter := bson.D{{"process_id", processID}}
	cursor, err := coll.Find(ctx, filter)
	handleError(ctx, err)
	var aips []ArchivePackage
	err = cursor.All(ctx, &aips)
	handleError(ctx, err)
	return aips
}

func FindArchivePackage(
	ctx context.Context,
	processID uuid.UUID,
	rootRecordIDs []uuid.UUID,
) (ArchivePackage, bool) {
	coll := mongoDatabase.Collection("archive_packages")
	filter := bson.D{
		{"process_id", processID},
		{"record_ids", bson.D{{"$all", rootRecordIDs}}},
	}
	var aip ArchivePackage
	err := coll.FindOne(ctx, filter).Decode(&aip)
	return aip, handleError(ctx, err)
}

func DeleteArchivePackagesForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("archive_packages")
	filter := bson.D{{"process_id", processID}}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}
