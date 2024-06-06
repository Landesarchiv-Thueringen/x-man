package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskState string

const (
	TaskStatePending TaskState = "pending"
	TaskStateRunning TaskState = "running"
	TaskStatePaused  TaskState = "paused"
	TaskStateFailed  TaskState = "failed"
	TaskStateDone    TaskState = "done"
)

type TaskAction string

const (
	TaskActionRun   TaskAction = "run"
	TaskActionRetry TaskAction = "retry"
	TaskActionPause TaskAction = "pause"
)

type TaskItem struct {
	Data  interface{} `json:"-"`
	Label string      `json:"label"`
	State TaskState   `json:"state"`
	Error string      `json:"error"`
}

type Task struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
	// ProcessID is the submission process that the task is for.
	ProcessID uuid.UUID `bson:"process_id" json:"processId"`
	// Type is the process step that the task is associated with.
	Type ProcessStepType `json:"type"`
	// State describes the current condition of the task.
	State TaskState `json:"state"`
	// Action is what should happen next with the task. E.g., when the task's
	// state is "running" and it's action is "retry", it should continue
	// running, retrying failed items. When it is done executing the action, the
	// action is set to the empty string.
	Action TaskAction `json:"action"`
	// Progress is a short notice that indicates the state of the task, e.g.,
	// "3 / 4".
	Progress string `json:"progress"`
	// Error describes an error if `State == "failed"`.
	Error string `bson:"error" json:"error"`
	// UserID is the LDAP user ID of the user who initiated the task, if any.
	UserID string `bson:"user_id" json:"userId"`
	// Items are the elements the task has to process.
	Items []TaskItem `json:"items"`
	// Data is additional data that is specific to the task type.
	Data interface{} `json:"data"`
}

func FindTasks(ctx context.Context) []Task {
	coll := mongoDatabase.Collection("tasks")
	filter := bson.D{}
	var tasks []Task
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	err = cursor.All(ctx, &tasks)
	if err != nil {
		panic(err)
	}
	return tasks
}

func FindTask(ctx context.Context, taskID primitive.ObjectID) (t Task, ok bool) {
	coll := mongoDatabase.Collection("tasks")
	filter := bson.D{{"_id", taskID}}
	err := coll.FindOne(ctx, filter).Decode(&t)
	if err == mongo.ErrNoDocuments {
		return t, false
	} else if err != nil {
		panic(err)
	}
	return t, true
}

func InsertTask(task Task) Task {
	coll := mongoDatabase.Collection("tasks")
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	task.State = TaskStateRunning
	result, err := coll.InsertOne(context.Background(), task)
	if err != nil {
		panic(err)
	}
	task.ID = result.InsertedID.(primitive.ObjectID)
	broadcastUpdate(Update{
		Collection: "tasks",
		ProcessID:  task.ProcessID,
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
		ProcessID:  t.ProcessID,
		Operation:  UpdateOperationUpdate,
	})
}

func mustUpdateTask(id primitive.ObjectID, update interface{}) {
	coll := mongoDatabase.Collection("tasks")
	result, err := coll.UpdateByID(context.Background(), id, update)
	if err != nil {
		panic(err)
	}
	if result.MatchedCount == 0 {
		panic(fmt.Sprintf("failed to update task %v: not found", id))
	}
	broadcastUpdate(Update{
		Collection: "tasks",
		// Note, that we currently don't populate the field ProcessID since it
		// is not used.
		Operation: UpdateOperationUpdate,
	})
}

func DeleteCompletedTasksOlderThan(t time.Time) (deletedCount int64) {
	coll := mongoDatabase.Collection("tasks")
	filter := bson.D{
		{"state", bson.D{{"$ne", "running"}}},
		{"updated_at", bson.D{{"$lt", t}}},
	}
	result, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount > 0 {
		broadcastUpdate(Update{
			Collection: "tasks",
			Operation:  UpdateOperationDelete,
		})
	}
	return result.DeletedCount
}
