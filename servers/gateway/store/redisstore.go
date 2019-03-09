package store

import (
	"github.com/go-redis/redis"
)

//RedisStore represents a session.Store backed by redis.
type RedisStore struct {
	Client *redis.Client
}

//NewRedisStore constructs a new RedisStore
func NewRedisStore(client *redis.Client) *RedisStore {
	//initialize and return a new RedisStore struct
	return &RedisStore{client}
}
