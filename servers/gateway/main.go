package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"questionqueue/servers/gateway/handlers"
	"questionqueue/servers/gateway/store"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

// Director is the director used for routing to microservices
type Director func(r *http.Request)

// CustomDirector forwards to the microservice and passes it the current user.
func CustomDirector(targets []*url.URL, ctx *handlers.HandlerContext) Director {
	var counter int32
	counter = 0
	mutex := sync.Mutex{}

	return func(r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		targ := targets[counter%int32(len(targets))]
		atomic.AddInt32(&counter, 1)
		r.Header.Add("X-Forwarded-Host", r.Host)
		r.Header.Del("X-User")
		// sessionState := &handlers.SessionState{}
		// _, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState)

		// If there is an error, we cannot deal with it here,
		// so forward it to the API to deal with it. (Could probably
		// deal with it here but don't know if I should pass the
		// responsewriter)
		if err != nil {
			r.Header.Add("X-User", "{}")
		} else {
			user := sessionState.User
			userJSON, err := json.Marshal(user)
			if err != nil {
				r.Header.Add("X-User", "{}")
			} else {
				r.Header.Add("X-User", string(userJSON))
			}
		}
		r.Host = targ.Host
		r.URL.Host = targ.Host
		r.URL.Scheme = targ.Scheme
	}
}

func getURLs(addrString string) []*url.URL {
	addrsSplit := strings.Split(addrString, ",")
	URLs := make([]*url.URL, len(addrsSplit))
	for i, c := range addrsSplit {
		URL, err := url.Parse(c)
		if err != nil {
			log.Fatal(fmt.Printf("Failure to parse url %v", err))
		}
		URLs[i] = URL
	}
	return URLs
}

//main is the main entry point for the server
func main() {
	// Read ADDR environment variable. If empty, default to ":80"
	addr := os.Getenv("ADDR")
	tlscert := getENVOrExit("TLSCERT")
	tlskey := getENVOrExit("TLSKEY")
	sessionKey := getENVOrExit("SESSIONKEY")
	redisAddr := getENVOrExit("REDISADDR")
	dsn := getENVOrExit("DSN")
	rabbitAddr := getENVOrExit("RABBITADDR")
	queueName := getENVOrExit("QUEUENAME")

	// Set up rabbit stuff
	conn, err := amqp.Dial(rabbitAddr)
	if err != nil {
		log.Fatalf("Error connecting to RabbitMQ: %s", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error opening a channel: %s", err)
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Error when declaring a queue: %s", err)
	}
	slackQueueMessages, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Error when setting up consumer: %s", err)
	}

	// Create new redis and mysql store
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	redisStore := store.NewRedisStore(client)
	if err != nil {
		log.Fatal("Unable to connect to mysql database")
	}
	ctx, err := handlers.NewHandlerContext(sessionKey, redisStore)
	if err != nil {
		log.Fatal("Unable to create new handler context")
	}

	// Set port
	if len(addr) == 0 {
		addr = ":443"
	}

	go ctx.Notifier.SendMessagesToWebsockets(slackQueueMessages)

	// set up proxies

	// Create new mux for web server and set routes
	mux := mux.NewRouter()
	mux.HandleFunc("/v1/ws", ctx.WebSocketConnectionHandler)

	// Wrap mux with CORS handler
	wrappedMux := handlers.NewCORS(mux)

	// Start web server, log errors
	log.Printf("server is listening at %s...", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlscert, tlskey, wrappedMux))
}

// gets the value of the environment variable of envName and returns it
// terminates the process if there is not value for envName
// Given from a lab exercise
func getENVOrExit(envName string) string {
	if env := os.Getenv(envName); len(env) > 0 {
		return env
	}
	log.Fatalf("no value set for %s, please set a value for %s", envName, envName)
	return ""
}
