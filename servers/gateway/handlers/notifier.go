package handlers

import (
	"sync"

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
