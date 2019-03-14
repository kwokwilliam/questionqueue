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
	getQueue := s.Client.Get(s.redisQueueName)
	if getQueue.Err() != nil {
		if getQueue.Err().Error() == "redis: nil" {
			return returnQueue, nil
		}
		return nil, getQueue.Err()
	}
	if unmarshallErr := json.Unmarshal([]byte(getQueue.Val()), returnQueue); unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return returnQueue, nil
}

// IsFoundSessionID will search the redis database for the session ID.
// if it is found it returns true. Otherwise it returns false, even
// in the case of an error.
func (s *RedisStore) IsFoundSessionID(bearerToken string) bool {
	getSessionID := s.Client.Get(bearerToken)
	if getSessionID.Err() != nil {
		return false
	}

	return getSessionID.Val() != ""
}
