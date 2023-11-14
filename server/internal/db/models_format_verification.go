package db

type FormatVerification struct {
	Summary                   map[string]Feature `json:"summary"`
	FileIdentificationResults []ToolResponse     `json:"fileIdentificationResults"`
	FileValidationResults     []ToolResponse     `json:"fileValidationResults"`
}

type ToolResponse struct {
	ToolName          string             `json:"toolName"`
	ToolVersion       string             `json:"toolVersion"`
	ToolOutput        *string            `json:"toolOutput"`
	OutputFormat      *string            `json:"outputFormat"`
	ExtractedFeatures *map[string]string `json:"extractedFeatures"`
	Error             *string            `json:"error"`
}

type ToolConfidence struct {
	ToolName   string  `json:"toolName"`
	Confidence float64 `json:"confidence"`
}

type FeatureValue struct {
	Value string           `json:"value"`
	Score float64          `json:"score"`
	Tools []ToolConfidence `json:"tools"`
}

type Feature struct {
	Key    string         `json:"key"`
	Values []FeatureValue `json:"values"`
}
