package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskState string

const (
	TaskStateRunning   TaskState = "running"
	TaskStateFailed    TaskState = "failed"
	TaskStateSucceeded TaskState = "succeeded"
)

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
	// Progress is a short notice that indicates the state of the task, e.g.,
	// "3 / 4".
	Progress string `json:"progress"`
	// ErrorMessage describes an error if `State == "failed"`.
	ErrorMessage string `bson:"error_message" json:"errorMessage"`
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
	return task
}

func MustUpdateTaskProgress(id primitive.ObjectID, progress string) {
	update := bson.D{{"$set", bson.D{
		{"updated_at", time.Now()},
		{"progress", progress},
	}}}
	mustUpdateTask(id, update)
}

func MustUpdateTaskState(id primitive.ObjectID, state TaskState, errorMessage string) {
	update := bson.D{{"$set", bson.D{
		{"updated_at", time.Now()},
		{"state", state},
		{"error_message", errorMessage},
	}}}
	mustUpdateTask(id, update)
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
}
