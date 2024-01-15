package db

import (
	"time"

	"gorm.io/gorm"
)

type Agency struct {
	ID                  uint               `gorm:"primaryKey" json:"id"`
	CreatedAt           time.Time          `json:"-"`
	UpdatedAt           time.Time          `json:"-"`
	DeletedAt           gorm.DeletedAt     `gorm:"index" json:"-"`
	Name                string             `json:"name"`
	Abbreviation        string             `json:"abbreviation"`
	TransferDirectoryID *uint              `json:"-"`
	TransferDirectory   *TransferDirectory `gorm:"foreignKey:TransferDirectoryID;references:ID" json:"transferDirectory"`
}

type TransferDirectory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	URL       string         `json:"url"`
	User      string         `json:"-"`
	Password  string         `json:"-"`
	Agencies  []Agency       `gorm:"many2many:transfer_directory_agencies;" json:"agencies"`
}
