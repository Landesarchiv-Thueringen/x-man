package db

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppraisalDecisionOption string

const (
	AppraisalDecisionEmpty AppraisalDecisionOption = ""
	AppraisalDecisionA     AppraisalDecisionOption = "A"
	AppraisalDecisionB     AppraisalDecisionOption = "B"
	AppraisalDecisionV     AppraisalDecisionOption = "V"
)

type Appraisal struct {
	ProcessID uuid.UUID               `bson:"process_id" json:"-"`
	RecordID  uuid.UUID               `bson:"record_id" json:"recordId"`
	Decision  AppraisalDecisionOption `json:"decision"`
	Note      string                  `json:"note"`
}

func FindAppraisalsForProcess(ctx context.Context, processID uuid.UUID) (appraisals []Appraisal) {
	coll := mongoDatabase.Collection("appraisals")
	filter := bson.D{{"process_id", processID}}
	cursor, err := coll.Find(ctx, filter)
	handleError(ctx, err)
	var a []Appraisal
	err = cursor.All(ctx, &a)
	handleError(ctx, err)
	return a
}

func FindAppraisal(processID uuid.UUID, recordID uuid.UUID) (a Appraisal, ok bool) {
	coll := mongoDatabase.Collection("appraisals")
	filter := bson.D{
		{"process_id", processID},
		{"record_id", recordID},
	}
	err := coll.FindOne(context.Background(), filter).Decode(&a)
	if err == mongo.ErrNoDocuments {
		return a, false
	} else if err != nil {
		panic(err)
	}
	return a, true
}

func UpsertAppraisal(
	processID uuid.UUID,
	recordID uuid.UUID,
	decision AppraisalDecisionOption,
	note string,
) {
	upsertAppraisal(processID, recordID, bson.D{
		{"decision", decision},
		{"note", note},
	})
}

func UpsertAppraisalDecision(
	processID uuid.UUID,
	recordID uuid.UUID,
	decision AppraisalDecisionOption,
) {
	upsertAppraisal(processID, recordID, bson.D{
		{"decision", decision},
	})
}

func UpsertAppraisalNote(
	processID uuid.UUID,
	recordID uuid.UUID,
	note string,
) {
	upsertAppraisal(processID, recordID, bson.D{
		{"note", note},
	})
}

func upsertAppraisal(
	processID uuid.UUID,
	recordID uuid.UUID,
	setItems bson.D,
) {
	coll := mongoDatabase.Collection("appraisals")
	filter := bson.D{
		{"process_id", processID},
		{"record_id", recordID},
	}
	update := bson.D{{"$set", append(bson.D{
		{"process_id", processID},
		{"record_id", recordID},
	}, setItems...)}}
	opts := options.Update().SetUpsert(true)
	_, err := coll.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		panic(err)
	}
}

func DeleteAppraisalsForProcess(processID uuid.UUID) {
	coll := mongoDatabase.Collection("appraisals")
	filter := bson.D{{"process_id", processID}}
	_, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
}
