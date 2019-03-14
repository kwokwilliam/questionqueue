package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"questionqueue/servers/gateway/store"

	"github.com/streadway/amqp"

	"github.com/gorilla/websocket"
)

var errFailNewContext = errors.New("Failed to create context")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandlerContext is a struct that will be a receiver on any
// HTTP handler functions that need access to globals
type HandlerContext struct {
	SessAndQueueStore store.Store
	Notifier          *Notifier
	Channel           *amqp.Channel
}

// NewHandlerContext creates a new handler context
func NewHandlerContext(SessAndQueueStore store.Store, ch *amqp.Channel) (*HandlerContext, error) {
	if SessAndQueueStore != nil {
		return &HandlerContext{SessAndQueueStore, &Notifier{}, ch}, nil
	}
	return nil, errFailNewContext
}

// WebSocketConnectionHandler handles the connection updator
// to a WebSocket connection.
func (ctx *HandlerContext) WebSocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// user is allowed to connect a websocket even if not authenticated

	// Upgrade connection to websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to open websocket connection", 401)
		return
	}

	// check if is teacher
	isTeacher := false
	if authHeader := r.URL.Query().Get("auth"); authHeader != "" {
		isTeacher = ctx.SessAndQueueStore.IsFoundSessionID(authHeader)
	}

	identification := r.URL.Query().Get("identification")
	if identification != "" {
		// insert connection to list
		ctx.Notifier.InsertConnection(conn, identification, isTeacher)

		// For each new websocket connection, start a goroutine to handler connection defer
		go (func(conn *websocket.Conn, ctx *HandlerContext, identification string) {
			defer conn.Close()
			defer ctx.Notifier.RemoveConnection(identification)
			for {
				messageType, p, err := conn.ReadMessage()
				if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
					fmt.Print("Client says", p)
					err := ctx.Channel.Publish(
						"",
						"queue",
						false,
						false,
						amqp.Publishing{
							ContentType: "text/plain",
							Body:        []byte("a"),
						})
					if err != nil {
						log.Printf("Failed to publish ws")
					}

				} else if messageType == websocket.CloseMessage || err != nil {
					break
				}
			}
		})(conn, ctx, identification)
	}

}
