package db

import (
	"time"

	"gorm.io/gorm"
)

type TaskState string

const (
	Running   TaskState = "running"
	Failed    TaskState = "failed"
	Succeeded TaskState = "succeeded"
)

type Task struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Title        string         `json:"title"`
	State        TaskState      `json:"state"`
	ErrorMessage string         `json:"errorMessage"`
}
