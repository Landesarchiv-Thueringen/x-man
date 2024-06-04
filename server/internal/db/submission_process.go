package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SubmissionProcess struct {
	// ProcessID is the process ID as parsed from an Xdomea message (ProzessID).
	ProcessID       uuid.UUID    `bson:"process_id" json:"processId"`
	CreatedAt       time.Time    `bson:"created_at" json:"createdAt"`
	Agency          Agency       `json:"agency"` // Copy, needs to be kept in sync
	StoreDir        string       `json:"-"`
	Message0502Path string       `bson:"message_0502_path" json:"-"`
	Message0504Path string       `bson:"message_0504_path" json:"-"`
	Message0506Path string       `bson:"message_0506_path" json:"-"`
	Note            string       `json:"note"`
	ProcessState    ProcessState `bson:"process_state" json:"processState"`
}

type ProcessState struct {
	Receive0501        ProcessStep `bson:"receive_0501" json:"receive0501"`
	Appraisal          ProcessStep `bson:"appraisal" json:"appraisal"`
	Receive0505        ProcessStep `bson:"receive_0505" json:"receive0505"`
	Receive0503        ProcessStep `bson:"receive_0503" json:"receive0503"`
	FormatVerification ProcessStep `bson:"format_verification" json:"formatVerification"`
	Archiving          ProcessStep `bson:"archiving" json:"archiving"`
}

type ProcessStepType string

const (
	ProcessStepReceive0501        ProcessStepType = "receive_0501"
	ProcessStepAppraisal          ProcessStepType = "appraisal"
	ProcessStepReceive0505        ProcessStepType = "receive_0505"
	ProcessStepReceive0503        ProcessStepType = "receive_0503"
	ProcessStepFormatVerification ProcessStepType = "format_verification"
	ProcessStepArchiving          ProcessStepType = "archiving"
)

type ProcessStep struct {
	// UpdatedAt is the last time the process step was modified in any way.
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
	// Complete is true if the step completed successfully.
	Complete bool `json:"complete"`
	// CompletedAt is the time at which Complete was set to true.
	CompletedAt time.Time `bson:"completed_at" json:"completedAt"`
	// CompletedBy is the name of the user who performed the process step.
	CompletedBy string `bson:"completed_by" json:"completedBy"`
	// Progress is a short notice that indicates the state of a not yet completed
	// process step, e.g., "3 / 4"
	Progress string `json:"progress"`
	// Running indicates that there is a task being currently executed for the
	// process step.
	Running bool `json:"running"`
	// UnresolvedErrors is the number of unresolved processing errors associated
	// with the process step. A number greater than 0 indicates a failed state.
	UnresolvedErrors int `bson:"unresolved_errors" json:"unresolvedErrors"`
}

func FindProcesses(ctx context.Context) []SubmissionProcess {
	return findProcesses(ctx, bson.D{{}})
}

func FindProcessesForUser(ctx context.Context, userID string) []SubmissionProcess {
	return findProcesses(ctx, bson.D{{"agency.users", bson.D{{"$all", bson.A{userID}}}}})
}

func findProcesses(ctx context.Context, filter interface{}) []SubmissionProcess {
	coll := mongoDatabase.Collection("submission_processes")
	var processes []SubmissionProcess
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	err = cursor.All(ctx, &processes)
	if err != nil {
		panic(err)
	}
	return processes
}

func FindOrInsertProcess(
	processID uuid.UUID,
	agency Agency,
	storeDir string,
) SubmissionProcess {
	process, ok := FindProcess(context.Background(), processID)
	if !ok {
		process = insertProcess(processID, agency, storeDir)
	}
	return process
}

func FindProcess(ctx context.Context, processID uuid.UUID) (SubmissionProcess, bool) {
	coll := mongoDatabase.Collection("submission_processes")
	var process SubmissionProcess
	filter := bson.D{{"process_id", processID}}
	err := coll.FindOne(ctx, filter).Decode(&process)
	if err == mongo.ErrNoDocuments {
		return process, false
	} else if err != nil {
		panic(err)
	}
	return process, true
}

func insertProcess(
	processID uuid.UUID,
	agency Agency,
	storeDir string,
) SubmissionProcess {
	coll := mongoDatabase.Collection("submission_processes")
	process := SubmissionProcess{
		ProcessID: processID,
		CreatedAt: time.Now(),
		Agency:    agency,
		StoreDir:  storeDir,
	}
	_, err := coll.InsertOne(context.Background(), process)
	if err != nil {
		panic(err)
	}
	broadcastUpdate(Update{
		Collection: "submission_processes",
		ProcessID:  processID,
		Operation:  UpdateOperationInsert,
	})
	return process
}

// DeleteProcess deletes the given submission process from the database.
//
// Do not call directly, instead use `xdomea.DeleteProcess`.
func DeleteProcess(processID uuid.UUID) (ok bool) {
	coll := mongoDatabase.Collection("submission_processes")
	filter := bson.D{{"process_id", processID}}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount == 0 {
		return false
	}
	broadcastUpdate(Update{
		Collection: "submission_processes",
		ProcessID:  processID,
		Operation:  UpdateOperationDelete,
	})
	return true
}

func UpdateProcessNote(
	processID uuid.UUID,
	note string,
) (ok bool) {
	update := bson.D{{"$set", bson.D{{"note", note}}}}
	return updateProcess(processID, update)
}

func MustUpdateProcessMessagePath(processID uuid.UUID, messageType MessageType, messagePath string) {
	var field string
	switch messageType {
	case MessageType0502, MessageType0504, MessageType0506:
		field = "message_" + string(messageType) + "_path"
	default:
		panic("unhandled message type: " + messageType)
	}
	update := bson.D{{"$set", bson.D{{field, messagePath}}}}
	ok := updateProcess(processID, update)
	if !ok {
		panic("failed to find process: " + processID.String())
	}
}

func MustUpdateProcessStepCompletion(
	processID uuid.UUID,
	step ProcessStepType,
	complete bool,
	completedBy string,
) {
	update := bson.D{{"$set", bson.D{
		{"process_state." + string(step) + ".updated_at", time.Now()},
		{"process_state." + string(step) + ".complete", complete},
		{"process_state." + string(step) + ".completed_at", time.Now()},
		{"process_state." + string(step) + ".completed_by", completedBy},
		{"process_state." + string(step) + ".running", false},
	}}}
	ok := updateProcess(processID, update)
	if !ok {
		panic("failed to update process step for process " + processID.String() + ": not found")
	}
}

func MustUpdateProcessStepProgress(
	processID uuid.UUID,
	step ProcessStepType,
	progress string,
	running bool,
) {
	update := bson.D{{"$set", bson.D{
		{"process_state." + string(step) + ".updated_at", time.Now()},
		{"process_state." + string(step) + ".progress", progress},
		{"process_state." + string(step) + ".running", running},
		{"process_state." + string(step) + ".complete", false},
	}}}
	ok := updateProcess(processID, update)
	if !ok {
		panic("failed to update process step for process " + processID.String() + ": not found")
	}
}

func updateProcess(processID uuid.UUID, update interface{}) (ok bool) {
	coll := mongoDatabase.Collection("submission_processes")
	filter := bson.D{{"process_id", processID}}
	result, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		panic(err)
	}
	ok = result.MatchedCount > 0
	if ok {
		broadcastUpdate(Update{
			Collection: "submission_processes",
			ProcessID:  processID,
			Operation:  UpdateOperationUpdate,
		})
	}
	return
}

// updateAgencyForProcesses updates the `Agency` field of all processes
// associated with the given agency.
func updateAgencyForProcesses(agency Agency) {
	coll := mongoDatabase.Collection("submission_processes")
	filter := bson.D{{"agency._id", agency.ID}}
	update := bson.D{{"$set", bson.D{{"agency", agency}}}}
	_, err := coll.UpdateMany(context.Background(), filter, update)
	if err != nil {
		panic(err)
	}
	broadcastUpdate(Update{
		Collection: "submission_processes",
		Operation:  UpdateOperationUpdate,
	})
}

func updateUnresolvedErrorsForProcessStep(processID uuid.UUID, step ProcessStepType, n int) {
	coll := mongoDatabase.Collection("submission_processes")
	filter := bson.D{{"process_id", processID}}
	update := bson.D{{"$set", bson.D{
		{"process_state." + string(step) + ".updated_at", time.Now()},
		{"process_state." + string(step) + ".unresolved_errors", n},
	}}}
	_, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		panic(err)
	}
	broadcastUpdate(Update{
		Collection: "submission_processes",
		ProcessID:  processID,
		Operation:  UpdateOperationUpdate,
	})
}
