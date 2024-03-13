package db

import (
	"time"

	"github.com/google/uuid"
)

type AppraisalDecisionOption string

const (
	AppraisalDecisionEmpty AppraisalDecisionOption = ""
	AppraisalDecisionA     AppraisalDecisionOption = "A"
	AppraisalDecisionB     AppraisalDecisionOption = "B"
	AppraisalDecisionV     AppraisalDecisionOption = "V"
)

type Appraisal struct {
	ID uint `gorm:"primaryKey" `
	// RecordObjectID is the xdomea ID of the appraised record object.
	RecordObjectID uuid.UUID               `json:"recordObjectID"`
	CreatedAt      time.Time               `json:"-"`
	UpdatedAt      time.Time               `json:"-"`
	ProcessID      string                  `json:"-"`
	Process        *Process                `gorm:"foreignKey:ProcessID;constraint:OnDelete:CASCADE" json:"-"`
	Decision       AppraisalDecisionOption `json:"decision"`
	InternalNote   string                  `json:"internalNote"`
}
