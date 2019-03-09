package store

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

//RedisStore represents a session.Store backed by redis.
type RedisStore struct {
	Client         *redis.Client
	redisQueueName string
}

//NewRedisStore constructs a new RedisStore
func NewRedisStore(client *redis.Client, redisQueueName string) *RedisStore {
	//initialize and return a new RedisStore struct
	return &RedisStore{client, redisQueueName}
}

// GetCurrentQueue gets the current queue from redis
func (s *RedisStore) GetCurrentQueue() (*QuestionQueue, error) {
	returnQueue := &QuestionQueue{}
	getQueue := s.Client.Get("queue")
	if getQueue.Err() != nil {
		return nil, getQueue.Err()
	}
	if unmarshallErr := json.Unmarshal([]byte(getQueue.Val()), returnQueue); unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return returnQueue, nil
}

// Question is used for individual questions
type Question struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Class     string `json:"class,omitempty"`
	Topic     string `json:"topic,omitempty"`
	Problem   string `json:"problem,omitempty"`
	LocationX string `json:"loc.x,omitempty"`
	LocationY string `json:"loc.y,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

// QuestionQueue will be unmarshalled from the redis store
type QuestionQueue struct {
	Queue []Question `json:"queue"`
}
