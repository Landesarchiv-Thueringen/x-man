// Package db provides types and accessor functions for database entries.
//
// Other business logic should be kept at a minimum.
//
// Accessor functions adhere to the following rules:
//   - Functions panic on unexpected errors.
//   - `Find`, `Update`, `Replace`, and `Delete` functions return an `ok` value
//     to indicate whether the document could be found.
//   - `MustFind`, `MustUpdate`, ... functions can be provided that panic when
//     the document could not be found.
//   - `FindWithDefault` functions return a default value when the document could not
//     be found.
//   - `Upsert` functions either insert a new document or update an existing
//     one.
//   - `Find` and `Delete` functions for multiple documents succeed without
//     errors, even if no documents were found or deleted.
package db

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDatabase *mongo.Database

// Init creates the client and context to access MongoDB.
func Init() {
	credential := options.Credential{
		Username: os.Getenv("MONGODB_USER"),
		Password: os.Getenv("MONGODB_PASSWORD"),
	}
	clientOpts := options.Client().
		ApplyURI(os.Getenv("MONGODB_URL")).
		SetAuth(credential).
		SetTimeout(10 * time.Second)
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		panic(err)
	}
	mongoDatabase = client.Database(os.Getenv("MONGODB_DB"))
	createIndexes()
}

func createIndexes() {
	createIndex("transfer_files", mongo.IndexModel{
		Keys: bson.D{
			{"agency_id", 1},
			{"path", 1},
		},
		Options: options.Index().SetUnique(true),
	})
	createIndex("transfer_files", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
		},
	})
	// We use an additional field because mongo express doesn't like UUIDs for
	// _id.
	createIndex("submission_processes", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
		},
		Options: options.Index().SetUnique(true),
	})
	createIndex("messages", mongo.IndexModel{
		Keys: bson.D{
			{"message_head.process_id", 1},
			{"message_type", 1},
		},
		Options: options.Index().SetUnique(true),
	})
	createIndex("archive_packages", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
		},
	})
	createIndex("root_records", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
			{"message_type", 1},
			{"contained_records", 1},
		},
	})
	createIndex("appraisals", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
			{"record_id", 1},
		},
		Options: options.Index().SetUnique(true),
	})
	createIndex("record_options", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
			{"record_id", 1},
		},
		Options: options.Index().SetUnique(true),
	})
	createIndex("primary_documents_data", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
			{"filename", 1},
		},
		Options: options.Index().SetUnique(true),
	})
	createIndex("warnings", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
		},
	})
	createIndex("processing_errors", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
			{"resolved", 1},
		},
	})
	createIndex("processing_errors", mongo.IndexModel{
		Keys: bson.D{
			{"task_id", 1},
			{"resolved", 1},
		},
	})
	createIndex("processing_errors", mongo.IndexModel{
		Keys: bson.D{
			{"resolved", 1},
			{"error_type", 1},
		},
	})
	createIndex("user_preferences", mongo.IndexModel{
		Keys: bson.D{
			{"user_id", 1},
		},
		Options: options.Index().SetUnique(true),
	})
}

func createIndex(collectionName string, model mongo.IndexModel) {
	coll := mongoDatabase.Collection(collectionName)
	_, err := coll.Indexes().CreateOne(context.Background(), model)
	if err != nil {
		panic(err)
	}
}

// handleError handles an error returned by a mongo function.
//
// It
//   - returns true if the function call succeeded,
//   - returns false if it was canceled by the context or didn't find any results,
//   - panics if there was any other error.
//
// The return value of this function might not be relevant to callers. For
// example, if values read from the database are discarded anyway when the
// context is canceled, this information is not needed anymore.
func handleError(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	} else if err == mongo.ErrNoDocuments {
		return false
	} else if err != nil {
		panic(err)
	}
	return true
}

func UnmarshalData[T any](d interface{}) T {
	t := reflect.TypeOf([0]T{}).Elem()
	if t.Kind() == reflect.Slice {
		return unmarshalData[struct{ A T }](struct{ A interface{} }{A: d}).A
	}
	return unmarshalData[T](d)
}

func unmarshalData[T any](d interface{}) T {
	var t T
	b, err := bson.Marshal(d)
	if err != nil {
		panic(fmt.Errorf("failed to marshal data: %w", err))
	}
	err = bson.Unmarshal(b, &t)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal data: %w", err))
	}
	return t
}

func UnmarshalArray[T any](a interface{}) []T {
	arr := a.(primitive.A)
	result := make([]T, len(arr))
	for i, d := range a.(primitive.A) {
		result[i] = UnmarshalData[T](d)
	}
	return result
}
