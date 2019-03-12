package db

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"questionqueue/src/model"
	"sync"
	"time"
)

const (
	dbName       = "question_queue"
	collClass    = "class"
	collTeacher  = "teacher"
	collQuestion = "question"
)

var (
	ErrEmailUsed = errors.New("this email address is already being used")
)

// MongoStore wraps the client to MongoDB with a struct.
type MongoStore struct {
	Client *mongo.Client
	lock   sync.Mutex
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
				lock:   sync.Mutex{}}, nil
		}
	}
}

// Disconnect disconnects current client's connection to MongoDB.
func (ms *MongoStore) Disconnect() error {
	return ms.Client.Disconnect(nil)
}

/*
Class
*/

// InsertClass adds a given `model.class` to MongoDB.
func (ms *MongoStore) InsertClass(class *model.Class) (*mongo.InsertOneResult, error) {
	return insert(ms.GetCollection(dbName, collClass), class)
}

// FindClass returns a `model.Class` using a `model.Class.Code`.
func (ms *MongoStore) GetOneClass(code string) (*model.Class, error) {

	cursor, err := ms.GetCollection(dbName, collClass).Find(nil, map[string]string{"Code": code}, nil)
	if err != nil {
		return nil, err
	}

	class := scanClass(cursor)

	// TODO: can this be done?
	//var cls []*model.Class
	//scanModel(cursor, model.Class{}, cls)

	if len(class) != 1 {
		return nil, errors.New(fmt.Sprintf("expect only 1 result, got %v results", len(class)))
	} else {
		return class[0], nil
	}
}

// GetAllClass returns all class documents from MongoDB.
func (ms *MongoStore) GetAllClass() ([]*model.Class, error) {
	cursor, err := ms.getAll(dbName, collClass)
	if err != nil {
		return nil, err
	} else {
		return scanClass(cursor), nil
	}
}

// UpdateClass takes a `model.Class` to overwrite a current class with a new `model.Class`.
// Only needs `model.Class.Code` property
func (ms *MongoStore) UpdateClass(old, new *model.Class) (*mongo.UpdateResult, error) {
	return update(ms.GetCollection(dbName, collClass), map[string]string{"code": old.Code}, new)
}

// UpdateClass takes a class code to overwrite a current class with a new `model.Class`.
func (ms *MongoStore) UpdateClassByCode(code string, new *model.Class) (*mongo.UpdateResult, error) {
	return update(ms.GetCollection(dbName, collClass), map[string]string{"code": code}, new)
}

// ScanClass takes a `mongo.Cursor`, parses all classes and return a slice of class pointers
func scanClass(cursor *mongo.Cursor) []*model.Class {
	var class []*model.Class
	for cursor.Next(nil) {
		c := model.Class{}
		// TODO: maybe return error?
		if err := cursor.Decode(&c); err != nil {
			log.Printf("cannot unmarshal class: %v", err)
			continue
		} else {
			class = append(class, &c)
		}
	}
	return class
}

/*
Question
*/

// GetAllQuestions returns all question documents from MongoDB.
func (ms *MongoStore) GetAllQuestions() ([]*model.Question, error) {
	cursor, err := ms.getAll(dbName, collQuestion)
	if err != nil {
		return nil, err
	} else {
		return scanQuestion(cursor), nil
	}
}

// GetActiveQuestions returns all questions that have not been solved
// by querying uninitialized `resolved_at` properties.
func (ms *MongoStore) GetActiveQuestions() ([]*model.Question, error) {
	if cursor, err := ms.GetCollection(dbName, collQuestion).
		Find(nil, bson.M{"resolvedat": bson.M{"$eq": time.Time{}}}, nil);
		err != nil {
		return nil, err
	} else {
		return scanQuestion(cursor), nil
	}
}

// InsertClass adds a given `model.class` to MongoDB.
func (ms *MongoStore) InsertQuestion(question *model.Question) (*mongo.InsertOneResult, error) {
	return insert(ms.GetCollection(dbName, collQuestion), question)
}

// SolveQuestion takes a `question.belongsTo` and updates `question.resolvedAt` property to current time.
func (ms *MongoStore) SolveQuestion(belongsTo string) (*mongo.UpdateResult, error) {
	return ms.GetCollection(dbName, collQuestion).
		// use UpdateMany() instead of UpdateOne() since only one result should be matched
		UpdateMany(nil,
			bson.M{
				"belongsto":  bson.M{"$eq": belongsTo},
				"resolvedat": bson.M{"$eq": time.Time{}}},
			bson.M{
				"$set": bson.M{"resolvedat": time.Now()}},
			options.Update().SetBypassDocumentValidation(false))
}

// ScanQuestion takes a `mongo.Cursor`, parses and return a slice of all classes found.
func scanQuestion(cursor *mongo.Cursor) []*model.Question {
	var question []*model.Question
	for cursor.Next(nil) {
		q := model.Question{}
		// TODO: maybe return error?
		if err := cursor.Decode(&q); err != nil {
			log.Printf("cannot unmarshal class: %v", err)
			continue
		} else {
			question = append(question, &q)
		}
	}
	return question
}

/*
Teacher
*/

// InsertTeacher adds a given `model.Teacher` to MongoDB.
func (ms *MongoStore) InsertTeacher(teacher *model.NewTeacher) (*mongo.InsertOneResult, error) {

	// check existing teachers with the same email address
	teachers, err := ms.GetTeacherByEmail(teacher.Email)
	if err != nil {
		return nil, err
	} else if len(teachers) > 0 {
		return nil, ErrEmailUsed
	}


	type t struct {
		Email        string `json:"email"`
		PasswordHash string `json:"password_hash"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
	}

	pwd, err := generatePassword(teacher.Password)
	if err != nil {
		return nil, err
	}

	return insert(ms.GetCollection(dbName, collTeacher), t{
		Email:         teacher.Email,
		PasswordHash:  pwd,
		FirstName:     teacher.FirstName,
		LastName:      teacher.LastName,
	})
}

// generatePassword takes a plaintext and returns the salted hash
func generatePassword(pwd string) (string, error) {

	const cost = 13

	if h, err := bcrypt.GenerateFromPassword([]byte(pwd), cost); err != nil {
		return "", err
	} else {
		return string(h), nil
	}
}

// UpdateTeacher takes a `model.TeacherUpdate` model, updates accordingly and returns results
func (ms *MongoStore) UpdateTeacher(tu *model.TeacherUpdate) (*mongo.UpdateResult, error) {

	const (
		h  = "password_hash"
		fn = "first_name"
		ln = "last_name"
	)

	m := make(map[string]interface{})

	if len(tu.NewPassword) > 0 {
		if pwd, err := generatePassword(tu.NewPassword);
		err != nil {
			return nil, err
		} else {
			m[h] = pwd
		}
	}

	if len(tu.FirstName) > 0 {
		m[fn] = tu.FirstName
	}

	if len(tu.LastName) > 0 {
		m[ln] = tu.LastName
	}

	return ms.GetCollection(dbName, collQuestion).
		// use UpdateMany() instead of UpdateOne() since only one result should be matched
		UpdateMany(nil,
			bson.M{
				"email":  bson.M{"$eq": tu.Email}},
			bson.M{
				"$set": m},
			options.Update().SetBypassDocumentValidation(false))
}

// GetTeacherByEmail gets teacher profile from MongoDB by taking an email address.
func (ms *MongoStore) GetTeacherByEmail(email string) ([]*model.Teacher, error) {
	if cursor, err := ms.GetCollection(dbName, collTeacher).
		Find(nil, bson.M{"email": bson.M{"$eq": email}}, nil); err != nil {
		return nil, err
	} else {
		return scanTeacher(cursor), nil
	}
}

// GetAllTeacher returns all teacher documents from MongoDB.
func (ms *MongoStore) GetAllTeacher() ([]*model.Teacher, error) {
	cursor, err := ms.getAll(dbName, collTeacher)
	if err != nil {
		return nil, err
	} else {
		return scanTeacher(cursor), nil
	}
}

// ScanTeacher takes a `mongo.Cursor`, parses and return a slice of all teachers found.
func scanTeacher(cursor *mongo.Cursor) []*model.Teacher {
	var teacher []*model.Teacher
	for cursor.Next(nil) {
		t := model.Teacher{}
		if err := cursor.Decode(&t); err != nil {
			log.Printf("cannot unmarshal teacher: %v", err)
			continue
		} else {
			teacher = append(teacher, &t)
		}
	}
	return teacher
}

/*
Helper
*/

// GetCollection returns a given collection in a given database.
func (ms *MongoStore) GetCollection(dbName, collName string) *mongo.Collection {
	return ms.Client.Database(dbName).Collection(collName)
}

// GetAll returns all documents from a given collection in a given database
func (ms *MongoStore) getAll(dbName, collName string) (*mongo.Cursor, error) {
	return ms.GetCollection(dbName, collName).Find(nil, map[string]string{}, nil)
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
	return coll.UpdateOne(nil, old, bson.M{"$set": new}, options.Update().SetBypassDocumentValidation(false))
}

//// TODO: refactor scanners
//// TODO: can this be done?
//func scanModel(cursor *mongo.Cursor, i interface{}, coll []interface{}) {
//	for cursor.Next(nil) {
//		if err := cursor.Decode(&i);
//			err != nil {
//			log.Printf("cannot unmarshal interface: %v", err)
//			continue
//		} else {
//			coll = append(coll, i)
//		}
//		clear(i)
//	}
//}
//
//// Clears the values of a given interface
//func clear(v interface{}) {
//	p := reflect.ValueOf(v).Elem()
//	p.Set(reflect.Zero(p.Type()))
//}
