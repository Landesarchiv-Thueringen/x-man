// Package routines calls reoccurring tasks in regular intervals.
package routines

import (
	"context"
	"fmt"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"lath/xman/internal/tasks"
	"lath/xman/internal/xdomea"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// interval is the time interval between scheduled routine runs.
const interval = 1 * time.Hour

// Init schedules regular execution for all routines.
func Init() {
	// Run on application start
	tryRestartRunningTasks()
	// Run periodically
	go func() {
		for {
			log.Println("Starting cleanup routines...")
			cleanupArchivedProcesses()
			cleanupTasksAndErrors()
			log.Println("Cleanup routines done")
			time.Sleep(interval)
		}
	}()
}

// tryRestartRunningTasks searches for tasks that are marked 'running' and tries
// to restart them.
func tryRestartRunningTasks() {
	defer errors.HandlePanic("tryRestartRunningTasks", nil, nil)
	ts := db.FindTasks(context.Background())
	for _, t := range ts {
		if t.State == db.TaskStateRunning {
			tryRestart(&t)
		}
	}
}

// tryRestart tries to restart a task after X-Man was shut down during
// execution.
//
// It marks the existing task as failed. In case the task can be restarted
// safely, it creates a new task that is equal to the existing one. Otherwise,
// it updates the process to reflect the failed task.
func tryRestart(task *db.Task) {
	var couldRestart = false
	switch task.Type {
	case db.ProcessStepFormatVerification:
		process, found := db.FindProcess(context.Background(), task.ProcessID)
		if found {
			message, found := db.FindMessage(context.Background(), process.ProcessID, db.MessageType0503)
			if found {
				go func() {
					defer errors.HandlePanic(
						"tryRestart FormatVerification",
						&db.ProcessingError{
							ProcessID:   process.ProcessID,
							MessageType: message.MessageType,
						},
						task,
					)
					xdomea.VerifyFileFormats(process, message)
				}()
				couldRestart = true
			}
		}
	}
	processingError := tasks.MarkFailed(task, "Abgebrochen durch Neustart von X-Man")
	if !couldRestart {
		errors.AddProcessingError(processingError)
	}
}

// cleanupArchivedProcesses deletes submission processes that have been archived
// successfully in the past.
//
// The time after which submission processes are deleted can be configured via
// the environment variable `DELETE_ARCHIVED_SUBMISSIONS_AFTER_DAYS`.
func cleanupArchivedProcesses() {
	defer errors.HandlePanic("cleanupArchivedProcesses", nil, nil)
	deleteDeltaDays, err := strconv.Atoi(os.Getenv("DELETE_ARCHIVED_SUBMISSIONS_AFTER_DAYS"))
	if err != nil {
		panic("missing or improper env variable DELETE_ARCHIVED_SUBMISSIONS_AFTER_DAYS")
	}
	deleteBeforeTime := time.Now().Add(-1 * time.Hour * 24 * time.Duration(deleteDeltaDays))
	processes := db.FindProcesses(context.Background())
	for _, process := range processes {
		if process.ProcessState.Archiving.Complete &&
			process.ProcessState.Archiving.CompletedAt.Before(deleteBeforeTime) {
			deleteProcess(process)
		}
	}
}

// cleanupTasksAndErrors deletes solved processing errors, that are not
// associated with a still existing submission process, and completed tasks.
//
// Processing errors that _are_ associated with a submission process will be
// deleted with the submission process (except application errors).
//
// The time after which tasks and errors are deleted can be configured with the
// environment variable `DELETE_TASKS_AND_ERRORS_AFTER_DAYS`
func cleanupTasksAndErrors() {
	defer errors.HandlePanic("cleanupTasksAndErrors", nil, nil)
	deleteDeltaDays, err := strconv.Atoi(os.Getenv("DELETE_TASKS_AND_ERRORS_AFTER_DAYS"))
	if err != nil {
		panic("missing or improper env variable DELETE_TASKS_AND_ERRORS_AFTER_DAYS")
	}
	deleteBeforeTime := time.Now().Add(-1 * time.Hour * 24 * time.Duration(deleteDeltaDays))
	// Delete resolved process errors, that are not associated with a still
	// existing submission process.
	processIDs := make(map[uuid.UUID]bool)
	for _, p := range db.FindProcesses(context.Background()) {
		processIDs[p.ProcessID] = true
	}
	for _, e := range db.FindResolvedProcessingErrorsOlderThan(context.Background(), deleteBeforeTime) {
		if e.ProcessID == uuid.Nil || !processIDs[e.ProcessID] {
			log.Println("Deleting processing error", e.Title)
			db.DeleteProcessingError(e.ID)
		}
	}
	// Delete completed tasks
	deletedCount := db.DeleteCompletedTasksOlderThan(deleteBeforeTime)
	if deletedCount > 0 {
		log.Println("Deleted", deletedCount, "tasks")
	}
}

// deleteProcess deletes the given process and all associated data from the
// database and removes all associated message files from the message store.
func deleteProcess(process db.SubmissionProcess) {
	found := xdomea.DeleteProcess(process.ProcessID)
	if !found {
		panic(fmt.Sprintf("failed to delete process %v: not found", process.ProcessID))
	}
}
