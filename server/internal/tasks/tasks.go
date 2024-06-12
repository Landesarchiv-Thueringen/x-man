package tasks

import (
	"context"
	"fmt"
	"lath/xman/internal/auth"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskHandler func(t *db.Task) (ItemHandler, error)

type ItemHandler interface {
	// HandleItem processes a single item of the task.
	HandleItem(itemData interface{}) error
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
	// RetrySafe indicates whether failed items or the entire task can safely be
	// rerun without risk of damage. When `true`, interrupted tasks will be
	// resumed automatically. In either case administrators have the possibility
	// to retry tasks manually.
	RetrySafe bool
}

type guard struct {
	Task chan struct{}
	Item chan struct{}
}

var handlers = make(map[db.ProcessStepType]TaskHandler)
var guards = make(map[db.ProcessStepType]guard)
var options = make(map[db.ProcessStepType]Options)
var activeTasks = make(map[primitive.ObjectID]*db.Task)
var pauseSignals = make(map[primitive.ObjectID]chan struct{})

func Action(taskID primitive.ObjectID, action db.TaskAction) error {
	t := activeTasks[taskID]
	switch action {
	case db.TaskActionPause:
		if t == nil {
			return fmt.Errorf("task not running")
		}
		pause(t)
	case db.TaskActionRun:
		if t != nil {
			return fmt.Errorf("task already running")
		}
		t, ok := db.FindTask(context.Background(), taskID)
		if !ok {
			return fmt.Errorf("task not found")
		}
		go func() {
			defer errors.HandlePanic("run task "+string(t.Type), &db.ProcessingError{
				ProcessID:   t.ProcessID,
				ProcessStep: t.Type,
			})
			Run(&t)
		}()
	case db.TaskActionRetry:
		if t != nil {
			return fmt.Errorf("task already running")
		}
		t, ok := db.FindTask(context.Background(), taskID)
		if !ok {
			return fmt.Errorf("task not found")
		}
		go func() {
			defer errors.HandlePanic("retry task "+string(t.Type), &db.ProcessingError{
				ProcessID:   t.ProcessID,
				ProcessStep: t.Type,
			})
			retry(&t)
		}()
	}
	return nil
}

func RegisterTaskHandler(taskType db.ProcessStepType, h TaskHandler, o Options) {
	handlers[taskType] = h
	options[taskType] = o
	g := guard{}
	if o.ConcurrentTasks != 0 {
		g.Task = make(chan struct{}, o.ConcurrentTasks)
	}
	g.Item = make(chan struct{}, o.ConcurrentItems)
	guards[taskType] = g
}

// Run starts or resumes a task.
//
// At the time `Run` is called, items are expected to have the state 'pending',
// 'failed', or 'done'.
func Run(t *db.Task) {
	log.Printf("Running %s for process %v...\n", t.Type, t.ProcessID)
	run(t)
}

func retry(t *db.Task) {
	log.Printf("Retrying %s for process %v...\n", t.Type, t.ProcessID)
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

func pause(t *db.Task) {
	select {
	case pauseSignals[t.ID] <- struct{}{}:
		// ok
	default:
		// No more pending tasks. Pause is a no-op at this point.
		//
		// Mark as "pausing" anyway, so the UI reflects the action.
	}
	t.State = db.TaskStatePausing
	updateProgress(t)
}

// run starts or resumes a task.
//
// Don't call directly. Instead use `Run` or `retry`.
func run(t *db.Task) {
	activeTasks[t.ID] = t
	defer delete(activeTasks, t.ID)
	pauseSignals[t.ID] = make(chan struct{})
	defer delete(pauseSignals, t.ID)
	t.State = db.TaskStatePending
	updateProgress(t)
	if g := guards[t.Type].Task; g != nil {
		select {
		case g <- struct{}{}:
			defer func() { <-g }()
		case <-pauseSignals[t.ID]:
			t.State = db.TaskStatePaused
			updateProgress(t)
			return
		}
	}
	t.State = db.TaskStateRunning
	updateProgress(t)
	h, err := handlers[t.Type](t)
	if err != nil {
		markFailed(t, err.Error())
		return
	}
	defer h.Finish()
	var wg sync.WaitGroup
	hasFailedItems := false
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
			case guards[t.Type].Item <- struct{}{}:
				// continue
			case <-pauseSignals[t.ID]:
				wg.Done()
				break ItemLoop
			}
			t.Items[i].State = db.TaskStateRunning
			go func() {
				defer errors.HandlePanic("process item for task "+string(t.Type), &db.ProcessingError{
					ProcessID:   t.ProcessID,
					ProcessStep: t.Type,
				})
				defer func() {
					wg.Done()
					<-guards[t.Type].Item
				}()
				err := h.HandleItem(item.Data)
				if err != nil {
					t.Items[i].State = db.TaskStateFailed
					t.Items[i].Error = err.Error()
					hasFailedItems = true
				} else {
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
	db.MustUpdateProcessStepProgress(t.ProcessID, t.Type, &t.Progress, t.State)
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
			if options[t.Type].RetrySafe {
				for i, item := range t.Items {
					if item.State == db.TaskStateRunning {
						t.Items[i].State = db.TaskStatePending
					}
				}
				Run(&t)
			} else {
				for i, item := range t.Items {
					if item.State == db.TaskStateRunning {
						t.Items[i].State = db.TaskStateFailed
						t.Items[i].Error = "Unterbrochen durch Neustart von X-Man"
					}
				}
				markFailed(&t, "Unterbrochen durch Neustart von X-Man")
			}
		}
	}
}

// markDone marks the task and its associated process step completed successfully.
func markDone(t *db.Task, completedBy string) {
	log.Printf("Task %s for process %v done\n", t.Type, t.ProcessID)
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
	if errMsg == "" {
		itemsFailed := 0
		if itemsFailed > 0 {
			errMsg = fmt.Sprintf("%d / %d fehlgeschlagen", itemsFailed, len(t.Items))
		}
	}
	t.Error = errMsg
	updateProgress(t)
	db.MustUpdateProcessStepError(t.ProcessID, t.Type)
	e, ok := db.FindUnresolvedProcessingErrorForTask(context.Background(), t.ID)
	if ok {
		e.Info = t.Error
		db.MustReplaceProcessingError(e)
	} else {
		e := db.ProcessingError{
			ProcessID:   t.ProcessID,
			ProcessStep: t.Type,
			Title:       getDisplayName(t.Type) + " fehlgeschlagen",
			Info:        t.Error,
			TaskID:      t.ID,
		}
		errors.AddProcessingError(e)
	}
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
