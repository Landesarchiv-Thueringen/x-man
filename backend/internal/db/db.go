package db

import (
	"errors"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var db *mongo.Database

var (
	ErrDbInit   = errors.New("database initialization failed")
	ErrDbInsert = errors.New("database insert failed")
	ErrDbFind   = errors.New("database find failed")
)

// Init creates the client to access MongoDB.
func Init() error {
	clientOpts := options.Client().
		ApplyURI(os.Getenv("MONGODB_URL"))
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Println(err)
		return ErrDbInit
	}
	db = client.Database(os.Getenv("MONGODB_DATABASE"))
	return nil
}
