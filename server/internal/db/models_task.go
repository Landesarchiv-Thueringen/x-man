package db

import (
	"time"
)

type TaskState string
type TaskType string

const (
	TaskStateRunning   TaskState = "running"
	TaskStateFailed    TaskState = "failed"
	TaskStateSucceeded TaskState = "succeeded"
)

const (
	TaskTypeFormatVerification TaskType = "formatVerification"
	TaskTypeArchiving          TaskType = "archiving"
)

type Task struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ProcessID string    `json:"processId"`
	// Process is the process that the task is for
	Process       *Process `json:"process"`
	ProcessStepID uint     `json:"-"`
	// ProcessStep is the process step that the task is for
	ProcessStep *ProcessStep `json:"-"`
	// Type is one of a list of known tasks
	Type TaskType `json:"type"`
	// State describes the current condition of the task
	State TaskState `json:"state"`
	// ErrorMessage describes an error if `State == "failed"`
	ErrorMessage string `json:"errorMessage"`
	// ItemCount is the number of items that have to be processed in this task
	ItemCount uint `gorm:"default:0" json:"itemCount"`
	// ItemCompletedCount is the number of items that have successfully been processed
	ItemCompletedCount uint `gorm:"default:0" json:"itemCompletedCount"`
}
