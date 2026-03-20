package db

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var db *mongo.Database

var ErrDbInit = errors.New("database initialization failed")

// Init creates the client and context to access MongoDB.
func Init() error {
	clientOpts := options.Client().
		ApplyURI(os.Getenv("MONGODB_URL"))
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Println(err)
		return ErrDbInit
	}
	// check database connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*5))
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Println(err)
		return ErrDbInit
	}
	db = client.Database(os.Getenv("MONGODB_DATABASE"))
	return nil
}
