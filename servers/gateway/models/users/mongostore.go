package users

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStore is a user.Store backed by MongoDB
type MongoStore struct {
	MongoClient *mongo.Client
}

// NewMongoStore constructs a new MongoStore and returns an error
// if there is a problem along the way
func NewMongoStore(uri string) (*MongoStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // See if this causes any issues. If it does, maybe try underscore
	MongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &MongoStore{MongoClient}, nil
}
