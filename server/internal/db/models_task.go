package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	ProcessID uuid.UUID      `json:"processId"`
	// Process is the process that the task is for
	Process       *Process `json:"process"`
	ProcessStepID uint
	ProcessStep   *ProcessStep
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
