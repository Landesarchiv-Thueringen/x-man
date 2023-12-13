package db

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Agency struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         string         `json:"name"`
	Abbreviation string         `json:"abbreviation"`
	TransferDir  string         `json:"-"`
}

func (a *Agency) IsFromTransferDir(path string) bool {
	return strings.Contains(path, a.TransferDir)
}
