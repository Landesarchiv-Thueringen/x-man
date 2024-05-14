package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ArchiveCollection refers to an archive collection within DIMAG.
type ArchiveCollection struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string             `json:"name"`
	DimagID string             `bson:"dimag_id" json:"dimagId"`
}

func FindArchiveCollections(ctx context.Context) []ArchiveCollection {
	coll := mongoDatabase.Collection("archive_collections")
	filter := bson.D{}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		panic(err)
	}
	var c []ArchiveCollection
	err = cursor.All(ctx, &c)
	if err != nil {
		panic(err)
	}
	return c
}

func FindArchiveCollection(ctx context.Context, id primitive.ObjectID) (c ArchiveCollection, ok bool) {
	coll := mongoDatabase.Collection("archive_collections")
	filter := bson.D{{"_id", id}}
	err := coll.FindOne(ctx, filter).Decode(&c)
	if err == mongo.ErrNoDocuments {
		return c, false
	} else if err != nil {
		panic(err)
	}
	return c, true
}

func InsertArchiveCollection(c ArchiveCollection) (id primitive.ObjectID) {
	coll := mongoDatabase.Collection("archive_collections")
	result, err := coll.InsertOne(context.Background(), c)
	if err != nil {
		panic(err)
	}
	return result.InsertedID.(primitive.ObjectID)
}

func ReplaceArchiveCollection(c ArchiveCollection) (ok bool) {
	coll := mongoDatabase.Collection("archive_collections")
	filter := bson.D{{"_id", c.ID}}
	result, err := coll.ReplaceOne(context.Background(), filter, c)
	if err != nil {
		panic(err)
	}
	if result.MatchedCount == 0 {
		return false
	}
	return true
}

func DeleteArchiveCollection(id primitive.ObjectID) (ok bool) {
	coll := mongoDatabase.Collection("archive_collections")
	filter := bson.D{{"_id", id}}
	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	if result.DeletedCount == 0 {
		return false
	}
	return true
}
