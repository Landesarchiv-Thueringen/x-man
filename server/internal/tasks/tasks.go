package tasks

import (
	"fmt"
	"lath/xman/internal/db"
	"time"

	"github.com/google/uuid"
)

// Start creates a task and marks the process step started.
func Start(taskType db.TaskType, process db.Process, itemCount uint) db.Task {
	processStep := getProcessStep(taskType, process)
	// Create task
	task := db.CreateTask(db.Task{
		Type:          taskType,
		State:         db.TaskStateRunning,
		ProcessID:     process.ID,
		ProcessStepID: processStep.ID,
		ItemCount:     itemCount,
	})
	task.Process = &process
	task.ProcessStep = &processStep
	// Update process step
	db.UpdateProcessStep(*&task.ProcessStep.ID, db.ProcessStep{
		Complete: false,
	})
	return task
}

func MarkItemComplete(task *db.Task) {
	task.ItemCompletedCount = task.ItemCompletedCount + 1
	db.UpdateTask(task.ID, db.Task{
		ItemCompletedCount: task.ItemCompletedCount,
	})
}

// MarkFailed marks the task and its process step failed.
//
// It returns a matching ProcessingError to be passed on.
func MarkFailed(task *db.Task, errorMessage string) db.ProcessingError {
	// Update task
	task.State = db.TaskStateFailed
	task.ErrorMessage = errorMessage
	db.UpdateTask(task.ID, db.Task{
		State:        task.State,
		ErrorMessage: task.ErrorMessage,
	})
	// The process step is marked failed by the processing error

	// Create processing error
	var processingErrorType db.ProcessingErrorType
	var messageID uuid.UUID
	switch task.Type {
	case db.TaskTypeArchiving:
		processingErrorType = db.ProcessingErrorArchivingFailed
		messageID = *task.Process.Message0503ID
	case db.TaskTypeFormatVerification:
		processingErrorType = db.ProcessingErrorFormatVerificationFailed
		messageID = *task.Process.Message0503ID
	}
	return db.ProcessingError{
		ProcessID:      &task.ProcessID,
		ProcessStepID:  &task.ProcessStepID,
		Type:           processingErrorType,
		AgencyID:       &task.Process.AgencyID,
		Description:    getDisplayName(task.Type) + " fehlgeschlagen",
		MessageID:      &messageID,
		AdditionalInfo: errorMessage,
	}
}

// MarkDone marks the task and its process stop completed successfully.
func MarkDone(task *db.Task, completedBy *string) {
	// Update task
	task.State = db.TaskStateSucceeded
	db.UpdateTask(task.ID, db.Task{
		State: task.State,
	})
	// Update process step
	completionTime := time.Now()
	db.UpdateProcessStep(task.ProcessStep.ID, db.ProcessStep{
		Complete:       true,
		CompletionTime: &completionTime,
		CompletedBy:    completedBy,
	})
}

// getProcessStep returns the process step to which the task belongs.
func getProcessStep(taskType db.TaskType, process db.Process) db.ProcessStep {
	switch taskType {
	case db.TaskTypeArchiving:
		return process.ProcessState.Archiving
	case db.TaskTypeFormatVerification:
		return process.ProcessState.FormatVerification
	default:
		panic(fmt.Errorf("unknown task type: %s", taskType))
	}
}

func getDisplayName(taskType db.TaskType) string {
	switch taskType {
	case db.TaskTypeArchiving:
		return "Archivierung"
	case db.TaskTypeFormatVerification:
		return "Formatverifikation"
	default:
		return string(taskType)
	}
}
