package db

import (
	"time"

	"gorm.io/gorm"
)

type FormatVerification struct {
	ID                        uint               `gorm:"primaryKey" json:"id"`
	CreatedAt                 time.Time          `json:"-"`
	UpdatedAt                 time.Time          `json:"-"`
	DeletedAt                 gorm.DeletedAt     `gorm:"index" json:"-"`
	Summary                   map[string]Feature `gorm:"-" json:"summary"`
	Features                  []Feature          `gorm:"many2many:format_verification_features;" json:"-"`
	FileIdentificationResults []ToolResponse     `gorm:"many2many:format_verification_identification_results;" json:"fileIdentificationResults"`
	FileValidationResults     []ToolResponse     `gorm:"many2many:format_verification_validation_results;" json:"fileValidationResults"`
}

type ToolResponse struct {
	ID                uint               `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time          `json:"-"`
	UpdatedAt         time.Time          `json:"-"`
	DeletedAt         gorm.DeletedAt     `gorm:"index" json:"-"`
	ToolName          string             `json:"toolName"`
	ToolVersion       string             `json:"toolVersion"`
	ToolOutput        *string            `json:"toolOutput"`
	OutputFormat      *string            `json:"outputFormat"`
	ExtractedFeatures *map[string]string `gorm:"-" json:"extractedFeatures"`
	Features          []ExtractedFeature `gorm:"many2many:tool_response_features;" json:"-"`
	Error             *string            `json:"error"`
}

type ExtractedFeature struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Key       string         `json:"key"`
	Value     string         `json:"value"`
}

type Feature struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Key       string         `json:"key"`
	Values    []FeatureValue `gorm:"many2many:feature_values;" json:"values"`
}

type FeatureValue struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time        `json:"-"`
	UpdatedAt time.Time        `json:"-"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`
	Value     string           `json:"value"`
	Score     float64          `json:"score"`
	Tools     []ToolConfidence `gorm:"many2many:feature_value_tools;" json:"tools"`
}

type ToolConfidence struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time      `json:"-"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	ToolName   string         `json:"toolName"`
	Confidence float64        `json:"confidence"`
}
