package tasks

import (
	"fmt"
	"lath/xman/internal/db"
	"time"

	"github.com/google/uuid"
)

// Start creates a task and marks the process step started.
func Start(taskType db.TaskType, process db.Process, itemCount uint) (db.Task, error) {
	processStep := getProcessStep(taskType, process)
	// Create task
	task, err := db.CreateTask(db.Task{
		Type:          taskType,
		State:         db.TaskStateRunning,
		ProcessID:     process.ID,
		Process:       &process,
		ProcessStepID: processStep.ID,
		ProcessStep:   &processStep,
		ItemCount:     itemCount,
	})
	if err != nil {
		return task, err
	}
	// Update process step
	task.ProcessStep.Complete = false
	task.ProcessStep.CompletionTime = nil
	err = db.UpdateProcessStep(*task.ProcessStep)
	if err != nil {
		return task, err
	}
	return task, nil
}

func MarkItemComplete(task *db.Task) error {
	task.ItemCompletedCount = task.ItemCompletedCount + 1
	return db.UpdateTask(*task)
}

// MarkFailed marks the task and its process step failed.
func MarkFailed(task *db.Task, errorMessage string, createProcessingError bool) error {
	// Update task
	task.State = db.TaskStateFailed
	task.ErrorMessage = errorMessage
	err := db.UpdateTask(*task)
	if err != nil {
		return err
	}
	// The process step is marked failed by the processing error

	// Create processing error
	if createProcessingError {
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
		db.AddProcessingError(db.ProcessingError{
			ProcessID:      task.ProcessID,
			ProcessStepID:  task.ProcessStepID,
			Type:           processingErrorType,
			AgencyID:       task.Process.AgencyID,
			Description:    getDisplayName(task.Type) + " fehlgeschlagen",
			MessageID:      messageID,
			AdditionalInfo: errorMessage,
		})
	}
	return nil
}

// MarkDone marks the task and its process stop completed successfully.
func MarkDone(task *db.Task) error {
	// Update task
	task.State = db.TaskStateSucceeded
	err := db.UpdateTask(*task)
	if err != nil {
		return err
	}
	// Update process step
	task.ProcessStep.Complete = true
	completionTime := time.Now()
	task.ProcessStep.CompletionTime = &completionTime
	return db.UpdateProcessStep(*task.ProcessStep)
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
