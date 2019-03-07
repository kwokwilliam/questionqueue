package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"questionqueue/servers/gateway/sessions"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// fmt.Printf(r.Header.Get("Origin"))
		// if r.Header.Get("Origin") == "https://uwinfotutor.me" {
		// 	return true
		// }
		// return false
		return true
	},
}

// WebSocketConnectionHandler handles the connection updator
// to a WebSocket connection.
func (ctx *HandlerContext) WebSocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	sessionState := &SessionState{}

	// Users must be authenticated to upgrade to a websocket
	_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState)
	if err != nil {
		http.Error(w, "Unauthorized Request", http.StatusUnauthorized)
		return
	}

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

// Notifier is a struct that will be controlling all notifications
// from websocket connections
type Notifier struct {
	Connections map[int64]*websocket.Conn
	lock        sync.Mutex
}

// InsertConnection is a thread safe method for inserting a connection
func (n *Notifier) InsertConnection(conn *websocket.Conn, uid int64) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if len(n.Connections) == 0 {
		n.Connections = make(map[int64]*websocket.Conn)
	}
	n.Connections[uid] = conn
}

// RemoveConnection will remove the connection
func (n *Notifier) RemoveConnection(uid int64) {
	n.lock.Lock()
	defer n.lock.Unlock()
	delete(n.Connections, uid)
}

// SendMessagesToWebsockets consumes messages from the rabbitmq message channel
// and sends it to either all the connections or a subset of the connections
func (n *Notifier) SendMessagesToWebsockets(messages <-chan amqp.Delivery) {
	for message := range messages {
		n.lock.Lock()
		messageToSend := &MQMessage{}
		if err := json.Unmarshal(message.Body, messageToSend); err != nil {
			fmt.Print("Error unmarshalling JSON-- should never reach here")
			return
		}
		// If the unmarshal has a single value, it creates an empty
		// slice as the userIDs because it fails to unmarshal it into
		// an array or something
		if len(messageToSend.UserIDs) == 0 {
			// notify all users
			for id, conn := range n.Connections {
				if err := conn.WriteMessage(websocket.TextMessage, message.Body); err != nil {
					n.RemoveConnection(id)
					conn.Close()
				}
			}
		} else {
			// notify subset of users
			for _, id := range messageToSend.UserIDs {
				conn, ok := n.Connections[id]
				if !ok {
					n.RemoveConnection(id)
				}

				if err := conn.WriteMessage(websocket.TextMessage, message.Body); err != nil {
					n.RemoveConnection(id)
					conn.Close()
				}
			}
		}

		message.Ack(false)
		n.lock.Unlock()
	}
}

// MQMessage is a struct for the messagequeue messages to unmarshal JSON into
type MQMessage struct {
	UserIDs []int64 `json:"userIDs,omitempty"`
}

//TODO: start a goroutine that connects to the RabbitMQ server,
//reads events off the queue, and broadcasts them to all of
//the existing WebSocket connections that should hear about
//that event. If you get an error writing to the WebSocket,
//just close it and remove it from the list
//(client went away without closing from
//their end). Also make sure you start a read pump that
//reads incoming control messages, as described in the
//Gorilla WebSocket API documentation:
//http://godoc.org/github.com/gorilla/websocket
