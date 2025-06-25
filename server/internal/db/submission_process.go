package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SubmissionProcess struct {
	// ProcessID is the process ID as parsed from an Xdomea message (ProzessID).
	ProcessID    string       `bson:"process_id" json:"processId"`
	CreatedAt    time.Time    `bson:"created_at" json:"createdAt"`
	Agency       Agency       `json:"agency"` // Copy, needs to be kept in sync
	StoreDir     string       `json:"-"`
	Note         string       `json:"note"`
	ProcessState ProcessState `bson:"process_state" json:"processState"`
	// UnresolvedErrors is the number of unresolved processing errors associated
	// with the submission process. A number greater than 0 indicates a failed
	// state.
	UnresolvedErrors int `bson:"unresolved_errors" json:"unresolvedErrors"`
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
	// If the process step consists of separate steps, Progress indicates how
	// many items are already processed and how many items there are in total.
	Progress *ItemProgress `json:"progress"`
	// If the process step is associated to a task, TaskID is its ID.
	TaskID primitive.ObjectID `bson:"task_id" json:"taskId"`
	// If the process step is associated to a task, TaskState represents its
	// current state. This field is set to the empty string when the process
	// step is completed, has errors, or is not associated to a task.
	TaskState TaskState `bson:"task_state" json:"taskState"`
	// HasError indicates whether there is one or more unresolved processing
	// error associated with the process step. True indicates a failed state.
	HasError bool `bson:"has_error" json:"hasError"`
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
	handleError(ctx, err)
	err = cursor.All(ctx, &processes)
	handleError(ctx, err)
	return processes
}

func FindOrInsertProcess(
	processID string,
	agency Agency,
	storeDir string,
) SubmissionProcess {
	process, ok := FindProcess(context.Background(), processID)
	if !ok {
		process = insertProcess(processID, agency, storeDir)
	}
	return process
}

func FindProcess(ctx context.Context, processID string) (SubmissionProcess, bool) {
	process, err := findProcess(ctx, processID)
	return process, handleError(ctx, err)
}

func TryFindProcess(ctx context.Context, processID string) (SubmissionProcess, bool) {
	process, err := findProcess(ctx, processID)
	return process, err == nil
}

func findProcess(ctx context.Context, processID string) (SubmissionProcess, error) {
	coll := mongoDatabase.Collection("submission_processes")
	var process SubmissionProcess
	filter := bson.D{{"process_id", processID}}
	err := coll.FindOne(ctx, filter).Decode(&process)
	return process, err
}

func insertProcess(
	processID string,
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
		ProcessID:  &processID,
		Operation:  UpdateOperationInsert,
	})
	return process
}

// DeleteProcess deletes the given submission process from the database.
//
// Do not call directly, instead use `xdomea.DeleteProcess`.
func DeleteProcess(processID string) (ok bool) {
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
		ProcessID:  &processID,
		Operation:  UpdateOperationDelete,
	})
	return true
}

func UpdateProcessNote(
	processID string,
	note string,
) (ok bool) {
	update := bson.D{{"$set", bson.D{{"note", note}}}}
	return updateProcess(processID, update)
}

func UpdateProcessStepCompletion(
	processID string,
	step ProcessStepType,
	complete bool,
	completedBy string,
) error {
	update := bson.D{{"$set", bson.D{
		{"process_state." + string(step) + ".updated_at", time.Now()},
		{"process_state." + string(step) + ".complete", complete},
		{"process_state." + string(step) + ".completed_at", time.Now()},
		{"process_state." + string(step) + ".completed_by", completedBy},
		{"process_state." + string(step) + ".task_state", ""},
		{"process_state." + string(step) + ".has_error", false},
	}}}
	ok := updateProcess(processID, update)
	if !ok {
		return fmt.Errorf("failed to update process step for process %v: not found", processID)
	}
	return nil
}

func MustUpdateProcessStepCompletion(
	processID string,
	step ProcessStepType,
	complete bool,
	completedBy string,
) {
	err := UpdateProcessStepCompletion(processID, step, complete, completedBy)
	if err != nil {
		panic(err)
	}
}

func MustUpdateProcessStepProgress(
	processID string,
	step ProcessStepType,
	progress *ItemProgress,
	taskID primitive.ObjectID,
	taskState TaskState,
) {
	update := bson.D{{"$set", bson.D{
		{"process_state." + string(step) + ".updated_at", time.Now()},
		{"process_state." + string(step) + ".progress", progress},
		{"process_state." + string(step) + ".task_id", taskID},
		{"process_state." + string(step) + ".task_state", taskState},
		{"process_state." + string(step) + ".complete", false},
		{"process_state." + string(step) + ".has_error", false},
	}}}
	ok := updateProcess(processID, update)
	if !ok {
		panic("failed to update process step for process " + processID + ": not found")
	}
}

func updateProcess(processID string, update interface{}) (ok bool) {
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
			ProcessID:  &processID,
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

func updateUnresolvedErrorsForProcess(processID string, unresolvedErrors []ProcessingError) {
	coll := mongoDatabase.Collection("submission_processes")
	filter := bson.D{{"process_id", processID}}
	old, err := coll.FindOne(context.Background(), filter).Raw()
	if err == mongo.ErrNoDocuments {
		return
	} else if err != nil {
		panic(err)
	}
	set := bson.D{{
		"unresolved_errors", len(unresolvedErrors),
	}}
	// hasErrorMap indicates for each process step whether it as any unresolved
	// errors.
	hasErrorMap := make(map[ProcessStepType]bool)
	for _, e := range unresolvedErrors {
		if e.ProcessStep != "" {
			hasErrorMap[e.ProcessStep] = true
		}
	}
	// For each process step, check whether the has_error changed and update the
	// step if so.
	for _, step := range []ProcessStepType{
		ProcessStepReceive0501,
		ProcessStepReceive0503,
		ProcessStepReceive0505,
		ProcessStepAppraisal,
		ProcessStepArchiving,
		ProcessStepFormatVerification,
	} {
		if hasErrorMap[step] != old.Lookup("process_state", string(step), "has_error").Boolean() {
			set = append(set,
				bson.E{"process_state." + string(step) + ".updated_at", time.Now()},
				bson.E{"process_state." + string(step) + ".has_error", hasErrorMap[step]},
			)
		}
	}
	update := bson.D{{"$set", set}}
	_, err = coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		panic(err)
	}
	broadcastUpdate(Update{
		Collection: "submission_processes",
		ProcessID:  &processID,
		Operation:  UpdateOperationUpdate,
	})
}
