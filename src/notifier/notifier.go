package notifier

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
		lock:        sync.Mutex{},
		channel:     channel,
		queue:       queue,
	}
}

// take a message from user and push to the queue defined in `Notifier`
func (n *Notifier) PublishMessage(message *Message) {

	n.lock.Lock()
	defer n.lock.Unlock()

	m, _ := json.Marshal(message)

	log.Printf("MQ got message: %v", string(m))

	err := n.channel.Publish(
		"",
		n.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        m,
		})

	if err != nil {
		log.Printf("cannot publish message: %v", err)
		return
	}
}
