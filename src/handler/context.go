package handler

import (
	"questionqueue/src/db"
	"questionqueue/src/notifier"
	"questionqueue/src/session"
	"questionqueue/src/trie"
)

type Context struct {
	Key          string
	SessionStore *session.RedisStore
	MongoStore   *db.MongoStore
	Trie         *trie.Trie
	Notifier     *notifier.Notifier
}

func NewContext(key string, redis *session.RedisStore, mongo *db.MongoStore, trie *trie.Trie, notifier *notifier.Notifier) *Context {
	return &Context{
		Key:          key,
		SessionStore: redis,
		MongoStore:   mongo,
		Trie:         trie,
		Notifier:     notifier,
	}
}