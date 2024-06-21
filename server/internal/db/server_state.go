package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServerStateXman struct {
	Version     string
	TokenSecret []byte `bson:"token_secret"`
}

func FindServerStateXman() (ServerStateXman, bool) {
	coll := mongoDatabase.Collection("server_state")
	filter := bson.D{{"_id", "xman"}}
	var s ServerStateXman
	err := coll.FindOne(context.Background(), filter).Decode(&s)
	if err == mongo.ErrNoDocuments {
		return s, false
	} else if err != nil {
		panic(err)
	}
	return s, true
}

func UpsertServerStateXmanVersion(version string) {
	coll := mongoDatabase.Collection("server_state")
	update := bson.D{{"$set", bson.D{{"version", version}}}}
	opts := options.Update().SetUpsert(true)
	coll.UpdateByID(context.Background(), "xman", update, opts)
}

func UpsertServerStateXmanTokenSecret(tokenSecret []byte) {
	coll := mongoDatabase.Collection("server_state")
	update := bson.D{{"$set", bson.D{{"token_secret", tokenSecret}}}}
	opts := options.Update().SetUpsert(true)
	coll.UpdateByID(context.Background(), "xman", update, opts)
}
