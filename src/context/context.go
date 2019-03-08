package context

import (
	"questionqueue/src/db"
	"questionqueue/src/sessions"
	"questionqueue/src/trie"
)

type Context struct {
	Key          string
	SessionStore *session.RedisStore
	MongoStore   *db.MongoStore
	Trie         *trie.Trie
}

func NewContext(key string, redis *session.RedisStore, mongo *db.MongoStore, trie *trie.Trie) *Context {
	return &Context{
		Key:          key,
		SessionStore: redis,
		MongoStore:   mongo,
		Trie:         trie,
	}
}