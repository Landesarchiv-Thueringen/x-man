package db

import (
	"time"
)

type FormatVerification struct {
	ID                        uint               `gorm:"primaryKey" json:"id"`
	CreatedAt                 time.Time          `json:"-"`
	UpdatedAt                 time.Time          `json:"-"`
	Summary                   map[string]Feature `gorm:"-" json:"summary"`
	Features                  []Feature          `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	FileIdentificationResults []ToolResponse     `gorm:"foreignKey:ParentIdentificationID;constraint:OnDelete:CASCADE" json:"fileIdentificationResults"`
	FileValidationResults     []ToolResponse     `gorm:"foreignKey:ParentValidationID;constraint:OnDelete:CASCADE" json:"fileValidationResults"`
}

type ToolResponse struct {
	ID                     uint               `gorm:"primaryKey" json:"id"`
	CreatedAt              time.Time          `json:"-"`
	UpdatedAt              time.Time          `json:"-"`
	ParentIdentificationID *uint              `json:"-"`
	ParentValidationID     *uint              `json:"-"`
	ToolName               string             `json:"toolName"`
	ToolVersion            string             `json:"toolVersion"`
	ToolOutput             *string            `json:"toolOutput"`
	OutputFormat           *string            `json:"outputFormat"`
	ExtractedFeatures      *map[string]string `gorm:"-" json:"extractedFeatures"`
	Features               []ExtractedFeature `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Error                  *string            `json:"error"`
}

type ExtractedFeature struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
	ToolResponseID uint      `json:"-"`
	Key            string    `json:"key"`
	Value          string    `json:"value"`
}

type Feature struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	CreatedAt            time.Time      `json:"-"`
	UpdatedAt            time.Time      `json:"-"`
	FormatVerificationID uint           `json:"-"`
	Key                  string         `json:"key"`
	Values               []FeatureValue `gorm:"constraint:OnDelete:CASCADE;" json:"values"`
}

type FeatureValue struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time        `json:"-"`
	UpdatedAt time.Time        `json:"-"`
	FeatureID uint             `json:"-"`
	Value     string           `json:"value"`
	Score     float64          `json:"score"`
	Tools     []ToolConfidence `gorm:"constraint:OnDelete:CASCADE;" json:"tools"`
}

type ToolConfidence struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
	FeatureValueID uint      `json:"-"`
	ToolName       string    `json:"toolName"`
	Confidence     float64   `json:"confidence"`
}
