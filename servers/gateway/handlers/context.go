package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"questionqueue/servers/gateway/store"

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
}

// NewHandlerContext creates a new handler context
func NewHandlerContext(SessAndQueueStore store.Store) (*HandlerContext, error) {
	if SessAndQueueStore != nil {
		return &HandlerContext{SessAndQueueStore, &Notifier{}}, nil
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

	// insert connection to list
	ctx.Notifier.InsertConnection(conn, sessionState.User.ID)
	// For each new websocket connection, start a goroutine
	// 		this goroutine will read incoming messages like the tutorial
	//		If receive error while reading, close websocket and remove from list
	go (func(conn *websocket.Conn, userID int64, ctx *HandlerContext) {
		defer conn.Close()
		defer ctx.Notifier.RemoveConnection(userID)
		for {
			messageType, p, err := conn.ReadMessage()
			if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
				fmt.Print("Client says", p)
			} else if messageType == websocket.CloseMessage || err != nil {
				break
			}
		}
	})(conn, sessionState.User.ID, ctx)
}
