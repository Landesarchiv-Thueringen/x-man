package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProcessingErrorType string

const (
	ProcessingErrorAgencyMismatch           ProcessingErrorType = "agency-mismatch"
	ProcessingErrorFormatVerificationFailed ProcessingErrorType = "format-verification-failed"
	ProcessingErrorArchivingFailed          ProcessingErrorType = "format-archiving-failed"
)

type ProcessingErrorResolution string

const (
	ErrorResolutionReimportMessage ProcessingErrorResolution = "reimport-message"
	ErrorResolutionDeleteMessage   ProcessingErrorResolution = "delete-message"
)

type ProcessingError struct {
	ID             uint                      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time                 `json:"detectedAt"`
	UpdatedAt      time.Time                 `json:"-"`
	DeletedAt      gorm.DeletedAt            `gorm:"index" json:"-"`
	Type           ProcessingErrorType       `json:"type"`
	ProcessID      *uuid.UUID                `json:"-"`
	Process        *Process                  `json:"process"`
	ProcessStepID  *uint                     `json:"-"`
	ProcessStep    *ProcessStep              `json:"-"`
	MessageID      *uuid.UUID                `json:"-"`
	Message        *Message                  `json:"message"`
	AgencyID       *uint                     `json:"-"`
	Agency         *Agency                   `gorm:"foreignKey:AgencyID;references:ID" json:"agency"`
	TransferPath   *string                   `json:"transferPath"`
	Resolved       bool                      `gorm:"default:false" json:"resolved"`
	Resolution     ProcessingErrorResolution `json:"resolution"`
	Description    string                    `json:"description"`
	AdditionalInfo string                    `json:"additionalInfo"`
}
