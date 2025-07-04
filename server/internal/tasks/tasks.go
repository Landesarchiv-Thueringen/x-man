// Package tasks handles time-consuming tasks handled by the backend. Tasks can
// be paused, resumed, and aborted by the user.
package tasks

import (
	"context"
	"fmt"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"log"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HandlerTemplate func(t *db.Task) (ItemHandler, error)

type ItemHandler interface {
	// HandleItem processes a single item of the task.
	HandleItem(
		ctx context.Context,
		itemData interface{},
		updateItemData func(data interface{}),
	) error
	// Finish handles cleanup work after all items of the task are handled or
	// the task is paused. The task is considered running until Finish returned.
	Finish()
	// AfterDone is called after the task finished successfully. It's execution
	// is not considered part of the task.
	AfterDone()
}

type Options struct {
	// ConcurrentTasks is the number of tasks of the given task type that may be
	// run concurrently. 0 indicates no limit.
	ConcurrentTasks int
	// ConcurrentItems is the total number of task items of the given task type
	// that may be handled concurrently across all running tasks.
	ConcurrentItems int
	// SafeRepeat indicates whether failed items or the entire task can safely be
	// rerun without risk of damage. When `true`, interrupted tasks will be
	// resumed automatically. In either case administrators have the possibility
	// to retry tasks manually.
	SafeRepeat bool
}

type taskHandler struct {
	HandlerTemplate
	Options
	TaskGuard chan struct{}
	ItemGuard chan struct{}
}

type runningTask struct {
	Task        *db.Task
	PauseSignal chan struct{}
	Cancel      func()
	Done        chan struct{}
}

var handlers = make(map[db.ProcessStepType]taskHandler)
var runningTasks = make(map[primitive.ObjectID]runningTask)

func Action(taskID primitive.ObjectID, action db.TaskAction) error {
	r, ok := runningTasks[taskID]
	switch action {
	case db.TaskActionPause:
		if !ok {
			return fmt.Errorf("task not running")
		}
		pause(r)
	case db.TaskActionResume:
		if ok {
			return fmt.Errorf("task already running")
		}
		t, ok := db.FindTask(context.Background(), taskID)
		if !ok {
			return fmt.Errorf("task not found")
		}
		resume(&t)
	case db.TaskActionRetry:
		if ok {
			return fmt.Errorf("task already running")
		}
		t, ok := db.FindTask(context.Background(), taskID)
		if !ok {
			return fmt.Errorf("task not found")
		}
		retry(&t)
	case db.TaskActionCancel:
		if ok {
			cancelRunning(r)
		} else {
			t, ok := db.FindTask(context.Background(), taskID)
			if !ok {
				return fmt.Errorf("task not found")
			}
			if t.State == db.TaskStatePaused {
				cancelPaused(&t)
			} else {
				return fmt.Errorf("cannot cancel task with state %s", t.State)
			}
		}
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
	return nil
}

func RegisterTaskHandler(taskType db.ProcessStepType, h HandlerTemplate, o Options) {
	t := taskHandler{
		HandlerTemplate: h,
		Options:         o,
	}
	if o.ConcurrentTasks != 0 {
		t.TaskGuard = make(chan struct{}, o.ConcurrentTasks)
	}
	t.ItemGuard = make(chan struct{}, o.ConcurrentItems)
	handlers[taskType] = t
}

// Run starts or resumes a task.
//
// At the time `Run` is called, items are expected to have the state 'pending',
// 'failed', or 'done'.
func Run(t *db.Task) {
	go func() {
		defer errors.HandlePanic("run task "+string(t.Type), &db.ProcessingError{
			ProcessID:   &t.ProcessID,
			ProcessStep: t.Type,
		})
		run(t)
	}()
}

func resume(t *db.Task) {
	log.Printf("Resuming %s for process %v...\n", t.Type, t.ProcessID)
	Run(t)
}

func retry(t *db.Task) {
	log.Printf("Retrying %s for process %v...\n", t.Type, t.ProcessID)
	if e, ok := db.FindUnresolvedProcessingErrorForTask(context.Background(), t.ID); ok {
		db.UpdateProcessingErrorResolve(e, db.ErrorResolutionRetryTask)
	}
	t.Error = ""
	for i, item := range t.Items {
		switch item.State {
		case db.TaskStateRunning, db.TaskStateFailed:
			t.Items[i].State = db.TaskStatePending
			t.Items[i].Error = ""
		}
	}
	Run(t)
}

func pause(r runningTask) {
	log.Printf("Pausing %s for process %v...\n", r.Task.Type, r.Task.ProcessID)
	select {
	case r.PauseSignal <- struct{}{}:
		// ok
	default:
		// No more pending tasks. Pause is a no-op at this point.
		//
		// Mark as "pausing" anyway, so the UI reflects the action.
	}
	r.Task.State = db.TaskStatePausing
	updateProgress(r.Task)
}

func cancelRunning(r runningTask) {
	log.Printf("Canceling %s for process %v...\n", r.Task.Type, r.Task.ProcessID)
	r.Cancel()
	<-r.Done
}

func cancelPaused(t *db.Task) {
	log.Printf("Canceling %s for process %v...\n", t.Type, t.ProcessID)
	markFailed(t, "Abgebrochen")
}

// Run starts or resumes a task.
//
// Don't call directly but use `Run`.
func run(t *db.Task) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := runningTask{
		Task:        t,
		PauseSignal: make(chan struct{}),
		Cancel:      cancel,
		Done:        make(chan struct{}),
	}
	defer close(r.Done)
	runningTasks[t.ID] = r
	defer delete(runningTasks, t.ID)
	th := handlers[t.Type]
	t.State = db.TaskStatePending
	updateProgress(t)
	if g := th.TaskGuard; g != nil {
		select {
		case g <- struct{}{}:
			defer func() { <-g }()
		case <-r.PauseSignal:
			t.State = db.TaskStatePaused
			updateProgress(t)
			return
		}
	}
	t.State = db.TaskStateRunning
	updateProgress(t)
	h, err := th.HandlerTemplate(t)
	if err != nil {
		markFailed(t, err.Error())
		return
	}
	defer h.Finish()
	var wg sync.WaitGroup
	hasFailedItems := false
	hasUnexpectedError := false
ItemLoop:
	for i, item := range t.Items {
		switch item.State {
		case db.TaskStateDone:
			// Do nothing
		case db.TaskStateFailed:
			hasFailedItems = true
		case db.TaskStatePending:
			wg.Add(1)
			select {
			case <-ctx.Done():
				wg.Done()
				break ItemLoop
			case <-r.PauseSignal:
				wg.Done()
				break ItemLoop
			case th.ItemGuard <- struct{}{}:
				// continue
			}
			if hasUnexpectedError {
				wg.Done()
				<-th.ItemGuard
				break ItemLoop
			}
			t.Items[i].State = db.TaskStateRunning
			db.MustReplaceTask(*t)
			go func() {
				defer func() {
					wg.Done()
					<-th.ItemGuard
				}()
				defer errors.HandlePanic("process item for task "+string(t.Type), &db.ProcessingError{
					ProcessID:   &t.ProcessID,
					ProcessStep: t.Type,
				}, func(e interface{}) {
					t.Items[i].State = db.TaskStateFailed
					t.Items[i].Error = fmt.Sprintf("%v", e)
					hasFailedItems = true
					hasUnexpectedError = true
				})
				log.Printf("Processing item %s...\n", item.Label)
				err := h.HandleItem(ctx, item.Data, func(data interface{}) {
					t.Items[i].Data = data
					db.MustReplaceTask(*t)
				})
				if err != nil {
					log.Printf("Error when processing item %s: %s\n", item.Label, err.Error())
					t.Items[i].State = db.TaskStateFailed
					t.Items[i].Error = err.Error()
					hasFailedItems = true
				} else {
					log.Printf("%s done\n", item.Label)
					t.Items[i].State = db.TaskStateDone
				}
				updateProgress(t)
			}()
		default:
			panic("encountered item with unexpected state '" + item.State + "'")
		}
	}
	wg.Wait()
	if t.State == db.TaskStatePausing && t.Progress.Done < t.Progress.Total {
		t.State = db.TaskStatePaused
		updateProgress(t)
	} else if ctx.Err() != nil {
		markFailed(t, "Abgebrochen")
	} else if hasFailedItems {
		markFailed(t, "")
	} else {
		var userName string
		if t.UserID != "" {
			userName = auth.GetDisplayName(t.UserID)
		}
		markDone(t, userName)
		h.AfterDone()
	}
}

// updateProgress updates the database entries for the task and the process step
// of the associated submission process based on the task's item states.
func updateProgress(t *db.Task) {
	itemsDone := 0
	for _, item := range t.Items {
		switch item.State {
		case db.TaskStateDone, db.TaskStateFailed:
			itemsDone++
		}
	}
	t.Progress.Total = len(t.Items)
	t.Progress.Done = itemsDone
	log.Printf("%s for process %v: %s (%v)\n", t.Type, t.ProcessID, t.State, t.Progress)
	db.MustReplaceTask(*t)
	db.MustUpdateProcessStepProgress(t.ProcessID, t.Type, &t.Progress, t.ID, t.State)
}

// ResumeAfterAppRestart searches for tasks that are marked 'running' and tries
// to restart them.
func ResumeAfterAppRestart() {
	defer errors.HandlePanic("tryRestartRunningTasks", nil)
	ts := db.FindTasks(context.Background())
	for _, t := range ts {
		switch t.State {
		case db.TaskStatePending:
			Run(&t)
		case db.TaskStateRunning, db.TaskStatePausing:
			if handlers[t.Type].SafeRepeat {
				for i, item := range t.Items {
					if item.State == db.TaskStateRunning {
						t.Items[i].State = db.TaskStatePending
					}
				}
				if t.State == db.TaskStateRunning {
					Run(&t)
				} else {
					t.State = db.TaskStatePaused
					updateProgress(&t)
				}
			} else {
				for i, item := range t.Items {
					if item.State == db.TaskStateRunning {
						t.Items[i].State = db.TaskStateFailed
						t.Items[i].Error = "Unterbrochen durch Neustart von x-man"
					}
				}
				markFailed(&t, "Unterbrochen durch Neustart von x-man")
			}
		}
	}
}

// markDone marks the task and its associated process step completed successfully.
func markDone(t *db.Task, completedBy string) {
	t.State = db.TaskStateDone
	t.Progress.Done = len(t.Items)
	updateProgress(t)
	db.MustUpdateProcessStepCompletion(t.ProcessID, t.Type, true, completedBy)
	if e, ok := db.FindUnresolvedProcessingErrorForTask(context.Background(), t.ID); ok {
		db.UpdateProcessingErrorResolve(e, db.ErrorResolutionObsolete)
	}

}

// markFailed marks the task and its associated process step failed and creates
// a processing error.
func markFailed(t *db.Task, errMsg string) {
	t.State = db.TaskStateFailed
	t.Error = errMsg
	if errMsg == "" {
		var itemErrs []string
		for _, item := range t.Items {
			switch item.State {
			case db.TaskStateFailed:
				itemErrs = append(itemErrs, item.Label+":\n\t"+item.Error)
			}
		}
		if len(itemErrs) > 0 {
			t.Error = fmt.Sprintf("%d / %d fehlgeschlagen", len(itemErrs), len(t.Items))
			errMsg = fmt.Sprintf(
				"%d / %d fehlgeschlagen\n\n%s",
				len(itemErrs), len(t.Items), strings.Join(itemErrs, "\n"),
			)
		}
	}
	updateProgress(t)
	e := db.ProcessingError{
		ProcessID:   &t.ProcessID,
		ProcessStep: t.Type,
		Title:       getDisplayName(t.Type) + " fehlgeschlagen",
		Info:        errMsg,
		TaskID:      &t.ID,
	}
	errors.AddProcessingError(e)
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

// CancelAndDeleteTasksForProcess cancels all running tasks for and deletes all
// tasks for the given process ID.
//
// If types is given, it cancels and deletes only the tasks matching the type.
// Otherwise, it cancels and deletes all tasks for the process.
func CancelAndDeleteTasksForProcess(processID string, types map[db.ProcessStepType]bool) {
	for _, r := range runningTasks {
		if r.Task.ProcessID == processID {
			if types == nil || types[r.Task.Type] {
				cancelRunning(r)
				db.DeleteTask(r.Task.ID)
			}
		}
	}
	for _, t := range db.FindTasksForProcess(context.Background(), processID) {
		if types == nil || types[t.Type] {
			db.DeleteTask(t.ID)
		}
	}
}
