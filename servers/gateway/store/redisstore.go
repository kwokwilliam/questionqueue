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
