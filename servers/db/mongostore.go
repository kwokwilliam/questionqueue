package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"questionqueue/servers/model"
	"sync"
)

const (
	// TODO: get a proper name
	dbName = "whatever"
	collClass = "class"
	collTeacher = "teacher"
	collQuestion = "question"
)

// MongoStore wraps the client to MongoDB with a struct.
type MongoStore struct {
	Client	*mongo.Client
	lock	sync.Mutex
}

// NewMongoStore Establishes a persistent connection to MongoDB
// at a provided address with no expiration time.
func NewMongoStore(addr string) (*MongoStore, error) {
	if client, err := mongo.Connect(nil, options.Client().ApplyURI(addr)); err != nil {
		return nil, err
	} else {
		// ping one more time before returning
		if err = client.Ping(nil, nil); err != nil {
			return nil, err
		} else {
			return &MongoStore{
				Client: client,
				lock: sync.Mutex{}}, nil
		}
	}
}

// Disconnect disconnects current client's connection to MongoDB.
func (ms *MongoStore) Disconnect() error {
	return ms.Client.Disconnect(nil)
}

// getCollection returns a given collection in a given database.
func (ms *MongoStore) getCollection(dbName, collectionName string) *mongo.Collection {
	return ms.Client.Database(dbName).Collection(collectionName)
}

// InsertClass adds a given `model.class` to MongoDB.
func (ms *MongoStore) InsertClass(class *model.Class) (*mongo.InsertOneResult, error) {
	return insert(ms.getCollection(dbName, collClass), class)
}

func (ms *MongoStore) UpdateClass(old, new *model.Class) (*mongo.UpdateResult, error) {
	return update(ms.getCollection(dbName, collClass), old, new)
}

// InsertClass adds a given `model.class` to MongoDB.
func (ms *MongoStore) InsertQuestion(question model.Question) (*mongo.InsertOneResult, error) {
	return insert(ms.getCollection(dbName, collQuestion), question)
}

func (ms *MongoStore) InsertTeacher(teacher model.Teacher) (*mongo.InsertOneResult, error) {
	return insert(ms.getCollection(dbName, collTeacher), teacher)
}

// Find takes a map filter and returns the pointer of the Cursor; A wrapper of the driver's `Find()`
func find(coll *mongo.Collection, m map[string]interface{}) (*mongo.Cursor, error) {
	return coll.Find(nil, bson.M(m), nil)
}

// Insert takes and insert a map into a given collection, returns that result;
// A wrapper of the driver's `InsertOnce()`
func insert(coll *mongo.Collection, m interface{}) (*mongo.InsertOneResult, error) {
	return coll.InsertOne(nil, m, options.InsertOne().SetBypassDocumentValidation(false))
}

func update(coll *mongo.Collection, old, new interface{}) (*mongo.UpdateResult, error) {
	return coll.UpdateOne(nil, old, bson.M{"$set" : new}, options.Update().SetBypassDocumentValidation(false))
}