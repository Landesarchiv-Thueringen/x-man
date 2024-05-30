package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProcessingErrorResolution string

const (
	ErrorResolutionMarkSolved         ProcessingErrorResolution = "mark-solved"
	ErrorResolutionReimportMessage    ProcessingErrorResolution = "reimport-message"
	ErrorResolutionDeleteMessage      ProcessingErrorResolution = "delete-message"
	ErrorResolutionDeleteTransferFile ProcessingErrorResolution = "delete-transfer-file"
	ErrorResolutionObsolete           ProcessingErrorResolution = "obsolete"
)

// ProcessingError represents any problem that should be communicated to
// clearing.
//
// Functions that encounter such a problem should return a ProcessingError.
// Higher-level functions are responsible for calling
// clearing.PassProcessingError.
type ProcessingError struct {
	ID           primitive.ObjectID        `bson:"_id,omitempty" json:"id"`
	CreatedAt    time.Time                 `bson:"created_at" json:"createdAt"`
	Resolved     bool                      `json:"resolved"`
	ResolvedAt   time.Time                 `bson:"resolved_at" json:"resolvedAt"`
	Resolution   ProcessingErrorResolution `json:"resolution"`
	Title        string                    `json:"title"`
	Info         string                    `bson:"info" json:"info"`
	Stack        string                    `json:"stack"`
	Agency       *Agency                   `json:"agency"` // Copy, needs to be kept in sync
	ProcessID    uuid.UUID                 `bson:"process_id" json:"processId"`
	MessageType  MessageType               `bson:"message_type" json:"messageType"`
	ProcessStep  ProcessStepType           `bson:"process_step" json:"-"`
	TransferPath string                    `bson:"transfer_path" json:"transferPath"`
}

func (e *ProcessingError) Error() string {
	return e.Title
}

// InsertProcessingError saves a processing error to the database.
//
// Do not call directly. Instead use clearing.HandleError.
func InsertProcessingError(e ProcessingError) primitive.ObjectID {
	coll := mongoDatabase.Collection("processing_errors")
	result, err := coll.InsertOne(context.Background(), e)
	if err != nil {
		panic(err)
	}
	// Update submission process
	if e.ProcessID != uuid.Nil && e.ProcessStep != "" {
		refreshUnresolvedErrorsForProcessStep(e.ProcessID, e.ProcessStep)
	}
	return result.InsertedID.(primitive.ObjectID)
}

func FindProcessingErrors(ctx context.Context) []ProcessingError {
	filter := bson.D{}
	return findProcessingErrors(ctx, filter)
}

func FindProcessingErrorsForProcess(ctx context.Context, processID uuid.UUID) []ProcessingError {
	filter := bson.D{{"process_id", processID}}
	return findProcessingErrors(ctx, filter)
}

func findProcessingErrors(ctx context.Context, filter interface{}) []ProcessingError {
	coll := mongoDatabase.Collection("processing_errors")
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	var e []ProcessingError
	err = cursor.All(ctx, &e)
	if err != nil {
		panic(err)
	}
	return e
}

func FindProcessingError(ctx context.Context, id primitive.ObjectID) (e ProcessingError, ok bool) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{{"_id", id}}
	err := coll.FindOne(ctx, filter).Decode(&e)
	if err == mongo.ErrNoDocuments {
		return e, false
	} else if err != nil {
		panic(err)
	}
	return e, true
}

// UpdateProcessingErrorResolve marks the given processing error as resolved.
func UpdateProcessingErrorResolve(e ProcessingError, r ProcessingErrorResolution) (ok bool) {
	coll := mongoDatabase.Collection("processing_errors")
	update := bson.D{{"$set", bson.D{
		{"resolved", true},
		{"resolved_at", time.Now()},
		{"resolution", r},
	}}}
	result, err := coll.UpdateByID(context.Background(), e.ID, update)
	if err != nil {
		panic(err)
	}
	if result.MatchedCount == 0 {
		return false
	}
	// Update submission process
	if e.ProcessID != uuid.Nil && e.ProcessStep != "" {
		refreshUnresolvedErrorsForProcessStep(e.ProcessID, e.ProcessStep)
	}
	return true
}

// DeleteProcessingErrorsForProcess deletes all processing errors associated
// with the given process.
//
// It expects the process to be deleted as well and will not update its values.
func DeleteProcessingErrorsForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{{"process_id", processID}}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}

func DeleteProcessingErrorsForMessage(processID uuid.UUID, messageType MessageType) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{
		{"process_id", processID},
		{"message_type", messageType},
	}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	switch messageType {
	case MessageType0501:
		refreshUnresolvedErrorsForProcessStep(processID, ProcessStepReceive0501)
	case MessageType0503:
		refreshUnresolvedErrorsForProcessStep(processID, ProcessStepReceive0503)
		refreshUnresolvedErrorsForProcessStep(processID, ProcessStepFormatVerification)
	}
}

func refreshUnresolvedErrorsForProcessStep(processID uuid.UUID, step ProcessStepType) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{
		{"process_id", processID},
		{"process_step", step},
		{"resolved", false},
	}
	n, err := coll.CountDocuments(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	updateUnresolvedErrorsForProcessStep(processID, step, int(n))
}

// updateAgencyForProcesses updates the `Agency` field of all processing errors
// associated with the given agency.
func updateAgencyForProcessingErrors(agency Agency) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{{"agency._id", agency.ID}}
	update := bson.D{{"$set", bson.D{{"agency", agency}}}}
	_, err := coll.UpdateMany(context.Background(), filter, update)
	if err != nil {
		panic(err)
	}
}
