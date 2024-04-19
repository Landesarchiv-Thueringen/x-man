package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProcessingErrorType string

const (
	ProcessingErrorPanic                    ProcessingErrorType = "panic"
	ProcessingErrorAgencyMismatch           ProcessingErrorType = "agency-mismatch"
	ProcessingErrorFormatVerificationFailed ProcessingErrorType = "format-verification-failed"
	ProcessingErrorArchivingFailed          ProcessingErrorType = "format-archiving-failed"
)

type ProcessingErrorResolution string

const (
	ErrorResolutionMarkSolved      ProcessingErrorResolution = "mark-solved"
	ErrorResolutionReimportMessage ProcessingErrorResolution = "reimport-message"
	ErrorResolutionDeleteMessage   ProcessingErrorResolution = "delete-message"
)

// ProcessingError represents any problem that should be communicated to
// clearing.
//
// Functions that encounter such a problem should return a ProcessingError.
// Higher-level functions are responsible for calling
// clearing.PassProcessingError.
type ProcessingError struct {
	ID             uint                      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time                 `json:"detectedAt"`
	UpdatedAt      time.Time                 `json:"-"`
	Type           ProcessingErrorType       `json:"type"`
	ProcessID      *string                   `json:"-"`
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

func (p ProcessingError) Error() string {
	return fmt.Sprintf("processing error: %v", p.Description)
}
