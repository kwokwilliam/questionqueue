package main

import (
	"github.com/gorilla/mux"
	"log"
	"os"
	"questionqueue/src/db"
	"questionqueue/src/handler"
	"questionqueue/src/session"
	"questionqueue/src/websocket"
	"time"
)

func main() {
	mongoAddr := os.Getenv("MONGOADDR")
	if len(mongoAddr) == 0 { mongoAddr = "mongo:27017" }

	redisAddr := os.Getenv("REDISADDR")
	if len(redisAddr) == 0 { redisAddr = "redis:6379" }

	rabbitAddr := os.Getenv("RABBITADDR")
	if len(rabbitAddr) == 0 { rabbitAddr = "amqp://guest:guest@rabbitmq:5672" }

	sessionKey := os.Getenv("SESSIONKEY")
	if len(sessionKey) == 0 { sessionKey = "default_key" }

	ms, err := db.NewMongoStore(mongoAddr)
	if err != nil {
		log.Fatalf("cannot connect to MongoDB: %v", err)
	}

	redis := session.NewRedisStore(session.NewRedisClient(redisAddr), time.Hour)

	mq, err := websocket.NewRabbitMQ(rabbitAddr)
	if err != nil {
		log.Fatalf("cannot connect to RabbitMQ: %v", err)
	}

	ch, err := mq.Channel()
	if err != nil {
		log.Fatalf("cannot get channel from RabbitMQ: %v", err)
	}

	q, err := ch.QueueDeclare(rabbitAddr, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("cannot declare queue: %v", err)
	}

	notifier := websocket.NewNotifier(ch, q)

	ctx := handler.Context{
		Key:          sessionKey, 		// TODO: get a key
		SessionStore: redis,
		MongoStore:   ms,
		Trie:         nil,
		Notifier:     notifier,
	}

	router := mux.NewRouter()

	// Teacher control: POST; PATCH
	router.HandleFunc("/v1/teacher", ctx.TeacherHandler)
	// Specific TA/teacher control: GET
	// only accepts `me`
	router.HandleFunc("/v1/teacher/{id}", ctx.TeacherProfileHandler)
	// TA/teacher session control: POST, DELETE
	router.HandleFunc("/v1/teacher/login", ctx.TeacherSessionHandler)
	// Student control - POSTing new questions and enqueue: POST
	router.HandleFunc("/v1/student", ctx.PostQuestionHandler)
}

