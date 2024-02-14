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
	ts, err := db.GetTasks()
	if err != nil {
		log.Println(err)
		return
	}
	for _, t := range ts {
		if t.State == db.TaskStateRunning {
			err = tryRestart(&t)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// tryRestart tries to restart a task after X-Man was shut down during
// execution.
//
// It marks the existing task as failed. In case the task can be restarted
// safely, it creates a new task that is equal to the existing one. Otherwise,
// it updates the process to reflect the failed task.
func tryRestart(task *db.Task) error {
	err := tasks.MarkFailed(task, "Abgebrochen durch Neustart von X-Man")
	if err != nil {
		return err
	}
	switch task.Type {
	case db.TaskTypeFormatVerification:
		process, err := db.GetProcess(task.ProcessID)
		if err == nil {
			go format.VerifyFileFormats(process, *process.Message0503)
		}
	}
	return nil
}

// cleanupArchivedProcesses deletes processes that have been archived
// successfully in the past.
//
// The time after which processes are deleted can be configured via the
// environment variable `DELETE_ARCHIVED_PROCESSES_AFTER_DAYS`.
func cleanupArchivedProcesses() {
	log.Println("Running cleanupArchivedProcesses")
	deleteDeltaDays, err := strconv.Atoi(os.Getenv("DELETE_ARCHIVED_PROCESSES_AFTER_DAYS"))
	if err != nil {
		log.Println(err)
		return
	}
	deleteBeforeTime := time.Now().Add(-1 * time.Hour * 24 * time.Duration(deleteDeltaDays))
	processes, err := db.GetProcesses()
	if err != nil {
		log.Println(err)
		return
	}
	for _, process := range processes {
		if process.ProcessState.Archiving.Complete &&
			process.ProcessState.Archiving.CompletionTime.Before(deleteBeforeTime) {
			log.Println("Deleting process", process.XdomeaID)
			err = deleteProcess(process)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// deleteProcess deletes the given process and all associated data from the
// database and removes all associated message files from the message store.
func deleteProcess(process db.Process) error {
	deleted, err := messagestore.DeleteProcess(process.XdomeaID)
	if err != nil {
		return err
	} else if !deleted {
		return fmt.Errorf("could not delete process", process.XdomeaID)
	}
	return nil
}
