package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// Warning represents a problem with a submission that is communicated to the
// archivist.
type Warning struct {
	CreatedAt   time.Time   `bson:"created_at" json:"createdAt"`
	Title       string      `bson:"title" json:"title"`
	ProcessID   string      `bson:"process_id" json:"processId"`
	MessageType MessageType `bson:"message_type" json:"-"`
}

// FindWarningsForProcess returns all warnings for the given submission process
// from the database.
func FindWarningsForProcess(ctx context.Context, processID string) []Warning {
	coll := mongoDatabase.Collection("warnings")
	filter := bson.D{{"process_id", processID}}
	cursor, err := coll.Find(ctx, filter)
	handleError(ctx, err)
	var w []Warning
	err = cursor.All(ctx, &w)
	handleError(ctx, err)
	return w
}

// InsertWarning saves a warning to the database.
func InsertWarning(w Warning) {
	coll := mongoDatabase.Collection("warnings")
	_, err := coll.InsertOne(context.Background(), w)
	if err != nil {
		panic(err)
	}
	// Update submission process
	broadcastUpdate(Update{
		Collection: "warnings",
		ProcessID:  &w.ProcessID,
		Operation:  UpdateOperationInsert,
	})
}

// DeleteWarningsForProcess deletes all warnings associated with the given
// submission process.
func DeleteWarningsForProcess(processID string) {
	coll := mongoDatabase.Collection("warnings")
	filter := bson.D{{"process_id", processID}}
	result, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount > 0 {
		broadcastUpdate(Update{
			Collection: "warnings",
			ProcessID:  &processID,
			Operation:  UpdateOperationDelete,
		})
	}
}

// DeleteWarningsForMessage deletes all warnings associated with the given
// message.
func DeleteWarningsForMessage(processID string, messageType MessageType) {
	coll := mongoDatabase.Collection("warnings")
	filter := bson.D{
		{"process_id", processID},
		{"message_type", messageType},
	}
	result, err := coll.DeleteMany(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount > 0 {
		broadcastUpdate(Update{
			Collection: "warnings",
			ProcessID:  &processID,
			Operation:  UpdateOperationDelete,
		})
	}
}
