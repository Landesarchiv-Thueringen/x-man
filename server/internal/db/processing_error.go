package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProcessingErrorResolution string

const (
	ErrorResolutionMarkSolved         ProcessingErrorResolution = "mark-solved"
	ErrorResolutionMarkDone           ProcessingErrorResolution = "mark-done"
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
	ErrorType    string                    `bson:"error_type" json:"-"`
	Stack        string                    `json:"stack"`
	Agency       *Agency                   `json:"agency"` // Copy, needs to be kept in sync
	ProcessID    uuid.UUID                 `bson:"process_id" json:"processId"`
	MessageType  MessageType               `bson:"message_type" json:"messageType"`
	ProcessStep  ProcessStepType           `bson:"process_step" json:"processStep"`
	TransferPath string                    `bson:"transfer_path" json:"transferPath"`
	TaskID       primitive.ObjectID        `bson:"task_id" json:"taskId"`
}

func (e *ProcessingError) Error() string {
	return e.Title
}

// InsertProcessingError saves a processing error to the database.
//
// Do not call directly. Instead use clearing.HandleError.
func InsertProcessingError(e ProcessingError) {
	coll := mongoDatabase.Collection("processing_errors")
	_, err := coll.InsertOne(context.Background(), e)
	if err != nil {
		panic(err)
	}
	// Update submission process
	if e.ProcessID != uuid.Nil {
		refreshUnresolvedErrorsForProcess(e.ProcessID)
	}
	broadcastUpdate(Update{
		Collection: "processing_errors",
		ProcessID:  e.ProcessID,
		Operation:  UpdateOperationInsert,
	})
}

func FindProcessingErrors(ctx context.Context) []ProcessingError {
	filter := bson.D{}
	return findProcessingErrors(ctx, filter)
}

func FindProcessingErrorsForProcess(ctx context.Context, processID uuid.UUID) []ProcessingError {
	filter := bson.D{{"process_id", processID}}
	return findProcessingErrors(ctx, filter)
}

func FindUnresolvedProcessingErrorsByType(ctx context.Context, errorType string) []ProcessingError {
	filter := bson.D{
		{"resolved", false},
		{"error_type", errorType},
	}
	return findProcessingErrors(ctx, filter)
}

func FindResolvedProcessingErrorsOlderThan(ctx context.Context, t time.Time) []ProcessingError {
	filter := bson.D{
		{"resolved", true},
		{"resolved_at", bson.D{{"$lt", t}}},
	}
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
	filter := bson.D{{"_id", id}}
	return findProcessingError(ctx, filter)
}

func FindUnresolvedProcessingErrorForTask(
	ctx context.Context,
	taskID primitive.ObjectID,
) (ProcessingError, bool) {
	filter := bson.D{
		{"task_id", taskID},
		{"resolved", false},
	}
	return findProcessingError(ctx, filter)
}

func findProcessingError(ctx context.Context, filter interface{}) (e ProcessingError, ok bool) {
	coll := mongoDatabase.Collection("processing_errors")
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
	if e.ProcessID != uuid.Nil {
		refreshUnresolvedErrorsForProcess(e.ProcessID)
	}
	broadcastUpdate(Update{
		Collection: "processing_errors",
		ProcessID:  e.ProcessID,
		Operation:  UpdateOperationUpdate,
	})
	return true
}

func MustReplaceProcessingError(e ProcessingError) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{{"_id", e.ID}}
	e.CreatedAt = time.Now()
	result, err := coll.ReplaceOne(context.Background(), filter, e)
	if err != nil {
		panic(err)
	}
	if result.MatchedCount == 0 {
		panic(fmt.Sprintf("failed to replace processing error %v: not found", e.ID))
	}
	broadcastUpdate(Update{
		Collection: "processing_errors",
		ProcessID:  e.ProcessID,
		Operation:  UpdateOperationUpdate,
	})
}

func DeleteProcessingError(ID primitive.ObjectID) (ok bool) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{{"_id", ID}}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	ok = result.DeletedCount > 0
	if ok {
		broadcastUpdate(Update{
			Collection: "processing_errors",
			Operation:  UpdateOperationDelete,
		})
	}
	return
}

// DeleteProcessingErrorsForProcess deletes all processing errors associated
// with the given process except application errors.
//
// It expects the process to be deleted as well and will not update its values.
func DeleteProcessingErrorsForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{
		{"process_id", processID},
		{"error_type", bson.D{{"$ne", "application-error"}}},
	}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	broadcastUpdate(Update{
		Collection: "processing_errors",
		ProcessID:  processID,
		Operation:  UpdateOperationDelete,
	})
}

func DeleteProcessingErrorsForMessage(processID uuid.UUID, messageType MessageType) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{
		{"process_id", processID},
		{"message_type", messageType},
		{"error_type", bson.D{{"$ne", "application-error"}}},
	}
	result, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount > 0 {
		refreshUnresolvedErrorsForProcess(processID)
		broadcastUpdate(Update{
			Collection: "processing_errors",
			ProcessID:  processID,
			Operation:  UpdateOperationDelete,
		})
	}
}

func refreshUnresolvedErrorsForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("processing_errors")
	filter := bson.D{
		{"process_id", processID},
		{"resolved", false},
	}
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	var errors []ProcessingError
	err = cursor.All(context.Background(), &errors)
	if err != nil {
		panic(err)
	}
	updateUnresolvedErrorsForProcess(processID, errors)
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
	broadcastUpdate(Update{
		Collection: "processing_errors",
		Operation:  UpdateOperationUpdate,
	})
}
