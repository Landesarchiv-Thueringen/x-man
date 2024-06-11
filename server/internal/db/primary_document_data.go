package db

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// PrimaryDocumentData represents data we gathered for a primary document in
// addition to the metadata given in an xdomea message.
type PrimaryDocumentData struct {
	ProcessID          uuid.UUID `bson:"process_id" json:"processId"`
	PrimaryDocument    `bson:"inline"`
	FileSize           int64               `bson:"file_size" json:"fileSize"`
	FormatVerification *FormatVerification `bson:"format_verification" json:"formatVerification"`
}

type FormatVerification struct {
	ProcessID                 uuid.UUID          `bson:"process_id" json:"-"`
	Filename                  string             `json:"-"`
	Summary                   map[string]Feature `json:"summary"`
	FileIdentificationResults []ToolResponse     `bson:"file_identification_results" json:"fileIdentificationResults"`
	FileValidationResults     []ToolResponse     `bson:"file_validation_results" json:"fileValidationResults"`
}

type ToolResponse struct {
	ToolName          string            `bson:"tool_name" json:"toolName"`
	ToolVersion       string            `bson:"tool_version" json:"toolVersion"`
	ToolOutput        string            `bson:"tool_output" json:"toolOutput"`
	OutputFormat      string            `bson:"output_format" json:"outputFormat"`
	ExtractedFeatures map[string]string `bson:"extracted_features" json:"extractedFeatures"`
	Error             string            `json:"error"`
}

type ExtractedFeature struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Feature struct {
	Key    string         `json:"key"`
	Values []FeatureValue `json:"values"`
}

type FeatureValue struct {
	Value string           `json:"value"`
	Score float64          `json:"score"`
	Tools []ToolConfidence `json:"tools"`
}

type ToolConfidence struct {
	ToolName   string  `bson:"tool_name" json:"toolName"`
	Confidence float64 `json:"confidence"`
}

// InsertPrimaryDocumentsData inserts multiple primary-document-data entries.
func InsertPrimaryDocumentsData(data []PrimaryDocumentData) {
	coll := mongoDatabase.Collection("primary_documents_data")
	entries := make([]interface{}, len(data))
	for i, d := range data {
		entries[i] = d
	}
	_, err := coll.InsertMany(context.Background(), entries)
	if err != nil {
		panic(err)
	}
}

func UpdatePrimaryDocumentFormatVerification(processID uuid.UUID, filename string, formatVerification *FormatVerification) {
	coll := mongoDatabase.Collection("primary_documents_data")
	filter := bson.D{{"process_id", processID}, {"filename", filename}}
	update := bson.D{{"$set", bson.D{{"format_verification", formatVerification}}}}
	_, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		panic(err)
	}
}

func FindPrimaryDocumentsDataForProcess(ctx context.Context, processID uuid.UUID) []PrimaryDocumentData {
	coll := mongoDatabase.Collection("primary_documents_data")
	filter := bson.D{{"process_id", processID}}
	var data []PrimaryDocumentData
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	err = cursor.All(ctx, &data)
	if err != nil {
		panic(err)
	}
	return data
}

func DeletePrimaryDocumentsDataForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("primary_documents_data")
	filter := bson.D{{"process_id", processID}}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}

func CalculateTotalFileSize(ctx context.Context, processID uuid.UUID, filenames []string) int64 {
	coll := mongoDatabase.Collection("primary_documents_data")
	matchStage := bson.D{{"$match", bson.D{
		{"process_id", processID},
		{"filename", bson.D{{"$in", filenames}}},
	}}}
	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", ""},
			{"total_file_size", bson.D{{"$sum", "$file_size"}}},
		}}}
	cursor, err := coll.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
	if err != nil {
		panic(err)
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		panic(err)
	}
	r := results[0]["total_file_size"]
	return r.(int64)
}
