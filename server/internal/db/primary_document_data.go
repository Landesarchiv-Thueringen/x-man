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
	RecordID           uuid.UUID `bson:"record_id" json:"recordId"`
	PrimaryDocument    `bson:"inline"`
	FileSize           int64               `bson:"file_size" json:"fileSize"`
	FormatVerification *FormatVerification `bson:"format_verification" json:"formatVerification"`
}

type FormatVerification struct {
	Summary     Summary                   `bson:"summary" json:"summary"`
	Features    map[string][]FeatureValue `bson:"features" json:"features"`
	ToolResults []ToolResult              `bson:"tool_results" json:"toolResults"`
}

type Summary struct {
	Valid            bool   `bson:"valid" json:"valid"`
	Invalid          bool   `bson:"invalid" json:"invalid"`
	FormatUncertain  bool   `bson:"format_uncertain" json:"formatUncertain"`
	ValidityConflict bool   `bson:"validity_conflict" json:"validityConflict"`
	Error            bool   `bson:"error" json:"error"`
	PUID             string `bson:"puid" json:"puid"`
	MimeType         string `bson:"mime_type" json:"mimeType"`
	FormatVersion    string `bson:"format_version" json:"formatVersion"`
}

type FeatureValue struct {
	Value           string             `bson:"value" json:"value"`
	Score           float64            `bson:"score" json:"score"`
	SupportingTools map[string]float64 `bson:"supporting_tools" json:"supportingTools"`
}

type ToolResult struct {
	ToolName     string            `bson:"tool_name" json:"toolName"`
	ToolType     string            `bson:"tool_type" json:"toolType"`
	ToolVersion  string            `bson:"tool_version" json:"toolVersion"`
	ToolOutput   string            `bson:"tool_output" json:"toolOutput"`
	OutputFormat string            `bson:"output_format" json:"outputFormat"`
	Features     map[string]string `bson:"features" json:"features"`
	Error        string            `bson:"error" json:"error"`
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
	handleError(ctx, err)
	err = cursor.All(ctx, &data)
	handleError(ctx, err)
	return data
}

func FindPrimaryDocumentData(ctx context.Context, processID uuid.UUID, filename string) (PrimaryDocumentData, bool) {
	coll := mongoDatabase.Collection("primary_documents_data")
	filter := bson.D{{"process_id", processID}, {"filename", filename}}
	var data PrimaryDocumentData
	err := coll.FindOne(ctx, filter).Decode(&data)
	ok := handleError(ctx, err)
	return data, ok
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
	handleError(ctx, err)
	var results []bson.M
	err = cursor.All(ctx, &results)
	handleError(ctx, err)
	r := results[0]["total_file_size"]
	return r.(int64)
}
