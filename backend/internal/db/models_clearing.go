package db

import (
	"time"

	"gorm.io/gorm"
)

type ProcessingError struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	CreatedAt        time.Time      `json:"detectedAt"`
	UpdatedAt        time.Time      `json:"-"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Description      string         `json:"description"`
	MessageStorePath *string        `json:"messageStorePath"`
	TransferDirPath  *string        `json:"transferDirPath"`
}
