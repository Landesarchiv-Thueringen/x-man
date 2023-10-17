package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProcessingError struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	CreatedAt        time.Time      `json:"detectedAt"`
	UpdatedAt        time.Time      `json:"-"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Resolved         bool           `gorm:"default:false" json:"resolved"`
	Description      string         `json:"description"`
	MessageID        *uuid.UUID     `json:"messageID"`
	MessageStorePath *string        `json:"messageStorePath"`
	TransferDirPath  *string        `json:"transferDirPath"`
}
