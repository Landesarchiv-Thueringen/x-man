// Package routines calls reoccurring tasks in regular intervals.
package routines

import (
	"fmt"
	"lath/xman/internal/db"
	"lath/xman/internal/format"
	"lath/xman/internal/messagestore"
	"lath/xman/internal/tasks"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"time"
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
			cleanupArchivedProcesses()
			time.Sleep(interval)
		}
	}()
}

// tryRestartRunningTasks searches for tasks that are marked 'running' and tries
// to restart them.
func tryRestartRunningTasks() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error: tryRestartRunningTasks panicked:", r)
			debug.PrintStack()
		}
	}()
	ts := db.GetTasks()
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
	case db.TaskTypeFormatVerification:
		process, found := db.GetProcess(task.ProcessID)
		if found {
			go format.VerifyFileFormats(process, *process.Message0503)
			couldRestart = true
		}
	}
	tasks.MarkFailed(task, "Abgebrochen durch Neustart von X-Man", !couldRestart)
}

// cleanupArchivedProcesses deletes processes that have been archived
// successfully in the past.
//
// The time after which processes are deleted can be configured via the
// environment variable `DELETE_ARCHIVED_PROCESSES_AFTER_DAYS`.
func cleanupArchivedProcesses() {
	log.Println("Starting cleanupArchivedProcesses...")
	defer logDoneOrRecover("cleanupArchivedProcesses")
	deleteDeltaDays, err := strconv.Atoi(os.Getenv("DELETE_ARCHIVED_PROCESSES_AFTER_DAYS"))
	if err != nil {
		panic(err)
	}
	deleteBeforeTime := time.Now().Add(-1 * time.Hour * 24 * time.Duration(deleteDeltaDays))
	processes := db.GetProcesses()
	for _, process := range processes {
		if process.ProcessState.Archiving.Complete &&
			process.ProcessState.Archiving.CompletionTime.Before(deleteBeforeTime) {
			log.Println("Deleting process", process.ID)
			deleteProcess(process)
		}
	}
}

// deleteProcess deletes the given process and all associated data from the
// database and removes all associated message files from the message store.
func deleteProcess(process db.Process) {
	found := messagestore.DeleteProcess(process.ID)
	if !found {
		panic(fmt.Sprintf("failed to delete process %v: not found", process.ID))
	}
}

// logDoneOrRecover prints a short log message when no error occurred or
// recovers from a panic.
func logDoneOrRecover(name string) {
	if r := recover(); r != nil {
		log.Println("Error:", name, "panicked:", r)
		debug.PrintStack()
	} else {
		log.Println(name, "done")
	}
}
