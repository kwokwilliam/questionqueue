package main

import (
	"github.com/gorilla/mux"
	"log"
	"os"
	"questionqueue/servers/gateway/handlers"
	"questionqueue/src/context"
	"questionqueue/src/db"
)

func main() {
	mongoAddr := os.Getenv("MONGOADDR")
	if len(mongoAddr) == 0 { mongoAddr = "mongo:27017" }

	ms, err := db.NewMongoStore(mongoAddr)
	if err != nil {
		log.Fatalf("cannot connect to MongoDB: %v", err)
	}

	ctx := context.Context{
		Key:          "",
		SessionStore: nil,
		MongoStore:   ms,
		Trie:         nil,
	}

	router := mux.NewRouter()
	router.HandleFunc("/v1/class", handlers.ClassHandler)
	router.HandleFunc("/v1/class/{id}", handlers.SpecificClassHandler)
	router.HandleFunc("/v1/teacher", handlers.TeacherHandler)
	router.HandleFunc("/v1/teacher/{id}", handlers.SpecificTeacherHandler)
	router.HandleFunc("/v1/teacher/login", handlers.TeacherLoginHandler)
}

