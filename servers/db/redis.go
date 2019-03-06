package db

import (
	"github.com/go-redis/redis"
	"time"
)

// RedisStore represents a session.Store backed by redis.
type RedisStore struct {
	// Redis client used to talk to redis server.
	Client *redis.Client
	// Used for key expiry time on redis.
	SessionDuration time.Duration
}

// NewRedis wraps and handles redis connection and returns a RedisStore object
func NewRedis(addr string, duration time.Duration) *RedisStore {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	return &RedisStore{c, duration}
}