package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client
var mongoDatabase *mongo.Database

// InitMongo creates the client and context to access MongoDB.
func InitMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	credential := options.Credential{
		Username: "root",
		Password: "example",
	}
	clientOpts := options.Client().
		ApplyURI("mongodb://mongo:27017").
		SetAuth(credential)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		panic(err)
	}
	mongoClient = client
	mongoDatabase = mongoClient.Database("xman")
}

// DisconnectMongo disconnects client and cancels the context.
func DisconnectMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := mongoClient.Disconnect(ctx); err != nil {
		panic(err)
	}
}

// AddMessageMongo adds xdomea message to collection.
func AddMessageMongo(message Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	// save root file record objects
	if len(message.FileRecordObjects) > 0 {
		//Convert record object type to empty interface.
		rootFiles := make([]interface{}, len(message.FileRecordObjects))
		for i, v := range message.FileRecordObjects {
			rootFiles[i] = v
		}
		col := mongoDatabase.Collection("rootFileRecordObjects")
		result, err := col.InsertMany(ctx, rootFiles)
		log.Println(result.InsertedIDs...)
		if err != nil {
			panic(err)
		}
	}
	// save root process record objects
	if len(message.ProcessRecordObjects) > 0 {
		//Convert record object type to empty interface.
		rootFiles := make([]interface{}, len(message.ProcessRecordObjects))
		for i, v := range message.ProcessRecordObjects {
			rootFiles[i] = v
		}
		col := mongoDatabase.Collection("rootProcessRecordObjects")
		result, err := col.InsertMany(ctx, rootFiles)
		log.Println(result.InsertedIDs...)
		if err != nil {
			panic(err)
		}
	}
	// save root document record objects
	if len(message.DocumentRecordObjects) > 0 {
		//Convert record object type to empty interface.
		rootFiles := make([]interface{}, len(message.DocumentRecordObjects))
		for i, v := range message.DocumentRecordObjects {
			rootFiles[i] = v
		}
		col := mongoDatabase.Collection("rootDocumentRecordObjects")
		result, err := col.InsertMany(ctx, rootFiles)
		log.Println(result.InsertedIDs...)
		if err != nil {
			panic(err)
		}
	}
	col := mongoDatabase.Collection("messages")
	_, err := col.InsertOne(ctx, message)
	if err != nil {
		panic(err)
	}
}
