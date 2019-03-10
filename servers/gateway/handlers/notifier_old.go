package handlers

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

// Notifier is a struct that will be controlling all notifications
// from websocket connections
type Notifiera struct {
	Connections map[int64]*websocket.Conn
	lock        sync.Mutex
}

// InsertConnection is a thread safe method for inserting a connection
func (n *Notifiera) InsertConnection(conn *websocket.Conn, uid int64) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if len(n.Connections) == 0 {
		n.Connections = make(map[int64]*websocket.Conn)
	}
	n.Connections[uid] = conn
}

// RemoveConnection will remove the connection
func (n *Notifiera) RemoveConnection(uid int64) {
	n.lock.Lock()
	defer n.lock.Unlock()
	delete(n.Connections, uid)
}

// SendMessagesToWebsockets consumes messages from the rabbitmq message channel
// and sends it to either all the connections or a subset of the connections
func (n *Notifiera) SendMessagesToWebsockets(messages <-chan amqp.Delivery) {
	for message := range messages {
		n.lock.Lock()
		messageToSend := &MQMessagea{}
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
type MQMessagea struct {
	UserIDs []int64 `json:"userIDs,omitempty"`
}
