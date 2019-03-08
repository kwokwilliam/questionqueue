package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
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
func (ms *MongoStore) getCollection(dbName, collName string) *mongo.Collection {
	return ms.Client.Database(dbName).Collection(collName)
}

// GetAll returns all documents from a given collection in a given database
func (ms *MongoStore) getAll(dbName, collName string) (*mongo.Cursor, error) {
	return ms.getCollection(dbName, collName).Find(nil, map[string]string{}, nil)
}

func (ms *MongoStore) GetAllClass() ([]*model.Class, error) {
	var classes []*model.Class

	cursor, err := ms.getAll("test", "testing")
	if err != nil {
		return nil, err
	}

	for cursor.Next(nil) {
		class := model.Class{}
		if err := cursor.Decode(&class);
		err != nil {
			log.Printf("cannot unmarshal class: %v", err)
			continue
		}

		if class.Code != "" && len(class.Type) > 0{
			classes = append(classes, &class)
		}
	}

	return classes, nil
}


// InsertClass adds a given `model.class` to MongoDB.
func (ms *MongoStore) InsertClass(class *model.Class) (*mongo.InsertOneResult, error) {
	return insert(ms.getCollection(dbName, collClass), class)
}

// FindClass returns a `model.Class` using a `model.Class.Code`.
func (ms *MongoStore) FindOneClass(code string) (*model.Class, error) {

	cursor, err := ms.getCollection(dbName, collClass).Find(nil, map[string]string{"Code": code}, nil)
	if err != nil {
		return nil, err
	}

	var class []model.Class

	for cursor.Next(nil) {
		c := model.Class{}
		err := json.Unmarshal([]byte(cursor.Current.String()), c)
		if err != nil {
			return nil, err
		}

		class = append(class, c)
	}

	if len(class) != 1 {
		return nil, errors.New(fmt.Sprintf("expect only 1 result, got %v results", len(class)))
	} else {
		return &class[0], nil
	}
}

// UpdateClass takes a `model.Class` to overwrite a current class with a new `model.Class`.
// Only needs `model.Class.Code` property
func (ms *MongoStore) UpdateClass(old, new *model.Class) (*mongo.UpdateResult, error) {
	return update(ms.getCollection(dbName, collClass), map[string]string{"code" : old.Code}, new)
}

// UpdateClass takes a class code to overwrite a current class with a new `model.Class`.
func (ms *MongoStore) UpdateClassByCode(code string, new *model.Class) (*mongo.UpdateResult, error) {
	return update(ms.getCollection(dbName, collClass), map[string]string{"code" : code}, new)
}

// InsertClass adds a given `model.class` to MongoDB.
func (ms *MongoStore) InsertQuestion(question model.Question) (*mongo.InsertOneResult, error) {
	return insert(ms.getCollection(dbName, collQuestion), question)
}

// SolveQuestion takes a question.ID and updates `question.resolvedAt` property to current time.
func (ms *MongoStore) SolveQuestion(id interface{}) (*mongo.UpdateResult, error) {
	b, _ := json.Marshal(id)

	//coll := ms.getCollection(dbName, collQuestion)
	//q, e := find(coll, map[string]interface{}{"_id": id})

	// TODO: get current, parse and unmarshal, change `resolvedAt`, overwrite

	return update(ms.getCollection(dbName, collClass), map[string]string{"_id" : string(b)}, nil)
}

// InsertTeacher adds a given `model.Teacher` to MongoDB.
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

// Update overwrites current document by taking a target collection ptr,
// a filter, and a new document to overwrite with
func update(coll *mongo.Collection, old, new interface{}) (*mongo.UpdateResult, error) {
	return coll.UpdateOne(nil, old, bson.M{"$set" : new}, options.Update().SetBypassDocumentValidation(false))
}