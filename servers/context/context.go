package context

import (
	"questionqueue/servers/db"
	"questionqueue/servers/sessions"
	"questionqueue/servers/trie"
)

type Context struct {
	Key          string
	SessionStore *sessions.RedisStore
	MongoStore   *db.MongoStore
	Trie         *trie.Trie
}

func NewContext(key string, redis *sessions.RedisStore, mongo *db.MongoStore, trie *trie.Trie) *Context {
	return &Context{
		Key:          key,
		SessionStore: redis,
		MongoStore:   mongo,
		Trie:         trie,
	}
}