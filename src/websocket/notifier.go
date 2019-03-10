package websocket

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
	"sync"
)

func NewRabbitMQ(addr string) (*amqp.Connection, error) {
	if conn, err := amqp.Dial(addr);
	err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

// a socket store consists of a list of connections that is available to the user, and a thread lock
type Notifier struct {
	// the int64 key denotes the userID.
	Connections map[interface{}]*WSClient
	// thread lock.
	lock sync.Mutex
	// a channel used as a context for valid message exchange.
	channel *amqp.Channel
	// queue to write messages to and read messages from.
	queue amqp.Queue
}

const (
	QuestionNew    = "question-new"
	QuestionDelete = "question-delete"
)

type Message struct {
	// type of the message
	Type    string      `json:"type"`
	// content of the message
	Content interface{} `json:"content"`
	// creator of the message
	UserID  interface{} `json:"userID"`
}

// create and return a new socket store
func NewNotifier(channel *amqp.Channel, queue amqp.Queue) *Notifier {
	return &Notifier{
		Connections: make(map[interface{}]*WSClient),
		lock:        sync.Mutex{},
		channel:     channel,
		queue:       queue,
	}
}

// insert a websocket client to a list of connections that is available and mapped to the user ID
func (n *Notifier) startClient(c *WSClient) {

	n.lock.Lock()
	defer n.lock.Unlock()

	n.Connections[c.Interface] = c
	go n.heartbeat(c)
}

// remove a websocket connection from the list of connections
func (n *Notifier) stopClient(userID interface{}) error {

	n.lock.Lock()
	defer n.lock.Unlock()

	// close connection
	err := n.Connections[userID].Connection.Close()

	// release resource connection cannot be closed
	if _, ok := n.Connections[userID]; ok {
		delete(n.Connections, userID)
	}

	return err
}

// monitor a connection from a user ensuring the user is still connected
// If an error is received while reading incoming control messages, close the WebSocket and remove it from the list.
func (n *Notifier) heartbeat(c *WSClient) {
	for {
		if _, _, err := c.Connection.NextReader(); err != nil {
			_ = n.stopClient(c.Interface)
			break
		}
	}
}

// indefinitely consumes messages from the mq
func (n *Notifier) MessageListener(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		n.lock.Lock()

		_ = msg.Ack(false)

		m := &Message{}
		err := json.Unmarshal(msg.Body, m)
		if err != nil {
			rawMsg, _ := json.Marshal(msg)
			log.Printf("cannot unmarshal message body: %v\nmsg: %v", err, string(rawMsg))
			return
		}

		js, _ := json.Marshal(msg)
		n.broadcastPublic(js)

		n.lock.Unlock()
	}
}

// write a marshaled js byte slice to all `Notifier` users
func (n *Notifier) broadcastPublic(js []byte) {
	for user, conn := range n.Connections {
		if err := conn.Connection.WriteJSON(js); err != nil {
			log.Printf("cannot write message to %v, closing connection.", user)
			_ = n.stopClient(user)
		}
	}
}

// write a marshaled js byte slice to a list of `Notifier` users
func (n *Notifier) broadcastPrivate(js []byte, users []int64) {
	for _, user := range users {
		if _, ok := n.Connections[user]; ok {
			conn := n.Connections[user]
			if err := conn.Connection.WriteJSON(js); err != nil {
				log.Printf("cannot write message to %v, closing connection.", user)
				_ = n.stopClient(user)
			}
		}
	}
}

// take a message from user and push to the queue defined in `Notifier`
func (n *Notifier) PublishMessage(message *Message) {

	//n.lock.Lock()
	//defer n.lock.Unlock()

	m, _ := json.Marshal(message)
	err := n.channel.Publish(
		"",
		n.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        m,
		})

	if err != nil {
		log.Printf("cannot publish message: %v", err)
		return
	}
}
