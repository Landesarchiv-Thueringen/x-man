package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskState string

const (
	TaskStatePending TaskState = "pending"
	TaskStateRunning TaskState = "running"
	TaskStatePaused  TaskState = "paused"
	TaskStatePausing TaskState = "pausing"
	TaskStateFailed  TaskState = "failed"
	TaskStateDone    TaskState = "done"
)

type TaskAction string

const (
	TaskActionResume TaskAction = "resume"
	TaskActionRetry  TaskAction = "retry"
	TaskActionPause  TaskAction = "pause"
	TaskActionCancel TaskAction = "cancel"
)

type TaskItem struct {
	Data  interface{} `json:"-"`
	Label string      `json:"label"`
	State TaskState   `json:"state"`
	Error string      `json:"error"`
}

type ItemProgress struct {
	Done  int `json:"done"`
	Total int `json:"total"`
}

func (p ItemProgress) String() string {
	return fmt.Sprintf("%d / %d", p.Done, p.Total)
}

type Task struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
	// ProcessID is the submission process that the task is for.
	ProcessID string `bson:"process_id" json:"processId"`
	// Type is the process step that the task is associated with.
	Type ProcessStepType `json:"type"`
	// State describes the current condition of the task.
	State TaskState `json:"state"`
	// Progress indicates how many items are already processed and how many
	// items there are in total.
	Progress ItemProgress `json:"progress"`
	// Error describes an error if `State == "failed"`.
	Error string `bson:"error" json:"error"`
	// UserID is the LDAP user ID of the user who initiated the task, if any.
	UserID string `bson:"user_id" json:"userId"`
	// Items are the elements the task has to process.
	Items []TaskItem `json:"items"`
	// Data is additional data that is specific to the task type.
	Data interface{} `json:"-"`
}

func FindTasksMetadata(ctx context.Context) []Task {
	coll := mongoDatabase.Collection("tasks")
	filter := bson.D{}
	opts := options.Find().SetProjection(bson.D{{"items", 0}})
	var tasks []Task
	cursor, err := coll.Find(ctx, filter, opts)
	handleError(ctx, err)
	err = cursor.All(ctx, &tasks)
	handleError(ctx, err)
	return tasks
}

func FindTasks(ctx context.Context) []Task {
	filter := bson.D{}
	return findTasks(ctx, filter)
}

func FindTasksForProcess(ctx context.Context, processID string) []Task {
	filter := bson.D{{"process_id", processID}}
	return findTasks(ctx, filter)
}

func findTasks(ctx context.Context, filter interface{}) []Task {
	coll := mongoDatabase.Collection("tasks")
	var tasks []Task
	cursor, err := coll.Find(ctx, filter)
	handleError(ctx, err)
	err = cursor.All(ctx, &tasks)
	handleError(ctx, err)
	return tasks
}

func FindTask(ctx context.Context, taskID primitive.ObjectID) (t Task, ok bool) {
	coll := mongoDatabase.Collection("tasks")
	filter := bson.D{{"_id", taskID}}
	err := coll.FindOne(ctx, filter).Decode(&t)
	return t, handleError(ctx, err)
}

func InsertTask(task Task) Task {
	coll := mongoDatabase.Collection("tasks")
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	result, err := coll.InsertOne(context.Background(), task)
	if err != nil {
		panic(err)
	}
	task.ID = result.InsertedID.(primitive.ObjectID)
	broadcastUpdate(Update{
		Collection: "tasks",
		ProcessID:  &task.ProcessID,
		Operation:  UpdateOperationInsert,
	})
	return task
}

func MustReplaceTask(t Task) {
	coll := mongoDatabase.Collection("tasks")
	filter := bson.D{{"_id", t.ID}}
	t.UpdatedAt = time.Now()
	result, err := coll.ReplaceOne(context.Background(), filter, t)
	if err != nil {
		panic(err)
	}
	if result.MatchedCount == 0 {
		panic(fmt.Sprintf("failed to update task %v: not found", t.ID))
	}
	broadcastUpdate(Update{
		Collection: "tasks",
		ProcessID:  &t.ProcessID,
		Operation:  UpdateOperationUpdate,
	})
}

func DeleteTask(ID primitive.ObjectID) {
	coll := mongoDatabase.Collection("tasks")
	filter := bson.D{
		{"_id", ID},
	}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount > 0 {
		broadcastUpdate(Update{
			Collection: "tasks",
			Operation:  UpdateOperationDelete,
		})
	}
}
