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
//   - `TryFind` functions return a default value when the document could not
//     be found.
//   - `Upsert` functions either insert a new document or update an existing
//     one.
//   - `Find` and `Delete` functions for multiple documents succeed without
//     errors, even if no documents were found or deleted.
package db

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
		ApplyURI("mongodb://mongo:27017").
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
	createIndex("transfer_dir_files", mongo.IndexModel{
		Keys: bson.D{
			{"agency_id", 1},
			{"transfer_dir_path", 1},
		},
		Options: options.Index().SetUnique(true),
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
			{"record_type", 1},
		},
	})
	createIndex("appraisals", mongo.IndexModel{
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
	createIndex("processing_errors", mongo.IndexModel{
		Keys: bson.D{
			{"process_id", 1},
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
