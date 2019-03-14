package handlers

import (
	"encoding/json"
	"log"
	"questionqueue/servers/gateway/store"
	"sync"

	"github.com/streadway/amqp"

	"github.com/gorilla/websocket"
)

// Notifier is a struct that will be controlling all notifications from
// websocket connections
type Notifier struct {
	Connections map[string]*QueueConnection
	lock        sync.Mutex
}

// QueueConnection is a struct that will keep track of the connection
// and if the user connected is a teacher or not
type QueueConnection struct {
	IsTeacher  bool
	Connection *websocket.Conn
}

// InsertConnection will insert the websocket connection based on the provided identification
// which Teachers will provide as well during the websocket connection.
func (n *Notifier) InsertConnection(conn *websocket.Conn, id string, isTeacher bool) {
	n.lock.Lock()
	defer n.lock.Unlock()
	newConnection := &QueueConnection{
		isTeacher,
		conn,
	}
	if len(n.Connections) == 0 {
		n.Connections = make(map[string]*QueueConnection)
	}
	n.Connections[id] = newConnection
}

// RemoveConnection will remove the websocket connection based on the provided identification
func (n *Notifier) RemoveConnection(id string) {
	n.lock.Lock()
	defer n.lock.Unlock()
	delete(n.Connections, id)
}

// SendMessagesToWebsockets is
func (n *Notifier) SendMessagesToWebsockets(messages <-chan amqp.Delivery, sessAndQueueStore store.Store) {
	for message := range messages {
		n.lock.Lock()
		log.Print("sending message")
		// For any received message, we immediately know it is because the queue has been updated.
		// First,we grab the current queue from redis
		currQueue, err := sessAndQueueStore.GetCurrentQueue()
		if err != nil {
			log.Printf("Error getting the current queue: %v", err)
		}
		log.Print("Got current queue")
		// get studentized queue and marshal both regular queue and student queue positions
		studentPositions := currQueue.GetStudentPositions()
		queueMarshalled, err := json.Marshal(currQueue)
		if err != nil {
			log.Printf("Error marshalling queue: %v", err)
		}
		log.Print("Marshalled queue")

		// Notify all the users of a new queue state
		for id, conn := range n.Connections {
			if conn.IsTeacher {
				if err := conn.Connection.WriteMessage(websocket.TextMessage, queueMarshalled); err != nil {
					n.RemoveConnection(id)
					conn.Connection.Close()
				}
			} else {
				studentPositionedMarshalled, err := json.Marshal(studentPositions[id])
				if err != nil {
					log.Printf("Error marshalling student position for %v: %v", id, err)
				}
				if err := conn.Connection.WriteMessage(websocket.TextMessage, studentPositionedMarshalled); err != nil {
					n.RemoveConnection(id)
					conn.Connection.Close()
				}
			}
		}

		message.Ack(false)
		n.lock.Unlock()
	}
}
