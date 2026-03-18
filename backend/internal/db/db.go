package db

import (
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var mongoDatabase *mongo.Database

// Init creates the client and context to access MongoDB.
func Init() {
	clientOpts := options.Client().
		ApplyURI(os.Getenv("MONGODB_URL"))
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		panic(err)
	}
	mongoDatabase = client.Database(os.Getenv("asdasd"))
}
