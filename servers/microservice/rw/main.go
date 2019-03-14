package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"questionqueue/src/db"
	"questionqueue/src/handler"
	"questionqueue/src/notifier"
	"questionqueue/src/session"
	"time"
)

func main() {

	addr := os.Getenv("ADDR")
	if len(addr) == 0 { addr = ":80" }

	mongoAddr := os.Getenv("MONGOADDR")
	if len(mongoAddr) == 0 { mongoAddr = "mongo:27017" }

	redisAddr := os.Getenv("REDISADDR")
	if len(redisAddr) == 0 { redisAddr = "redis:6379" }

	rabbitAddr := os.Getenv("RABBITADDR")
	if len(rabbitAddr) == 0 { rabbitAddr = "amqp://guest:guest@rabbitmq:5672" }

	sessionKey := os.Getenv("SESSIONKEY")
	if len(sessionKey) == 0 { sessionKey = "default_key" }

	log.Println("mongoAddr:",mongoAddr)
	ms, err := db.NewMongoStore(mongoAddr)
	if err != nil {
		log.Fatalf("cannot connect to MongoDB: %v", err)
	}

	redis := session.NewRedisStore(session.NewRedisClient(redisAddr), time.Hour)

	mq, err := notifier.NewRabbitMQ(rabbitAddr)
	if err != nil {
		log.Fatalf("cannot connect to RabbitMQ: %v", err)
	}

	ch, err := mq.Channel()
	if err != nil {
		log.Fatalf("cannot get channel from RabbitMQ: %v", err)
	}

	queueName := "queue"
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("cannot declare queue: %v", err)
	}

	n := notifier.NewNotifier(ch, q)

	ctx := handler.Context{
		Key:          sessionKey,
		SessionStore: redis,
		MongoStore:   ms,
		Trie:         nil,
		Notifier:     n,
	}

	router := mux.NewRouter()

	// test connection
	router.HandleFunc("/test", ctx.OkHandler)
	// Teacher control: POST; PATCH
	router.HandleFunc("/v1/teacher", ctx.TeacherHandler)
	// TA/teacher session control: POST, DELETE
	router.HandleFunc("/v1/teacher/login", ctx.TeacherSessionHandler)
	// Specific TA/teacher control: GET
	// only accepts `me` or `all`
	router.HandleFunc("/v1/teacher/{id}", ctx.TeacherProfileHandler)
	// Question control - POSTing new questions and enqueue: POST
	router.HandleFunc("/v1/student", ctx.PostQuestionHandler)
	// Question control - DELETE dequeues an existing question: DELETE
	router.HandleFunc("/v1/student/{id}", ctx.DeleteQuestionHandler)

	log.Println("mongo:", mongoAddr)
	log.Println("redis:",redisAddr)
	log.Println("sessionKey:",sessionKey)
	log.Println("mq:",rabbitAddr)

	log.Printf("microservice is running at http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, handler.NewLogger(router)))
}

