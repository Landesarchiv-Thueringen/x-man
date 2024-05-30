package tasks

import (
	"lath/xman/internal/db"
	"log"

	"github.com/google/uuid"
)

// Start creates a task and marks the process step started.
func Start(taskType db.ProcessStepType, processID uuid.UUID, progress string) db.Task {
	log.Printf("Starting %s for process %v...\n", taskType, processID)
	// Create task
	task := db.InsertTask(db.Task{
		Type:      taskType,
		State:     db.TaskStateRunning,
		ProcessID: processID,
		Progress:  progress,
	})
	// Update process step
	db.MustUpdateProcessStepProgress(processID, taskType, task.Progress, true)
	return task
}

func Progress(task db.Task, progress string) {
	log.Printf("%s for process %v: %s\n", task.Type, task.ProcessID, progress)
	// Update task
	db.MustUpdateTaskProgress(task.ID, progress)
	// Update process step
	db.MustUpdateProcessStepProgress(task.ProcessID, task.Type, progress, true)
}

// MarkFailed marks the task and its process step failed.
//
// It returns a matching ProcessingError to be passed on.
func MarkFailed(task *db.Task, errorMessage string) db.ProcessingError {
	// Update task
	db.MustUpdateTaskState(task.ID, db.TaskStateFailed, errorMessage)
	// Update processing step
	db.MustUpdateProcessStepProgress(task.ProcessID, task.Type, "", false)
	// Create processing error
	return db.ProcessingError{
		ProcessID:   task.ProcessID,
		ProcessStep: task.Type,
		Title:       getDisplayName(task.Type) + " fehlgeschlagen",
		Info:        errorMessage,
	}
}

// MarkDone marks the task and its process stop completed successfully.
func MarkDone(task db.Task, completedBy string) {
	log.Printf("Task %s for process %v done\n", task.Type, task.ProcessID)
	// Update task
	db.MustUpdateTaskState(task.ID, db.TaskStateSucceeded, "")
	// Update process step
	db.MustUpdateProcessStepCompletion(task.ProcessID, task.Type, true, completedBy)
}

func getDisplayName(taskType db.ProcessStepType) string {
	switch taskType {
	case db.ProcessStepArchiving:
		return "Archivierung"
	case db.ProcessStepFormatVerification:
		return "Formatverifikation"
	default:
		return string(taskType)
	}
}
