package handler

import (
	"net/http"
	"questionqueue/src/db"
	"questionqueue/src/session"
	"questionqueue/src/trie"
	"questionqueue/src/websocket"
	"time"
)

type Context struct {
	Key          string
	SessionStore *session.RedisStore
	MongoStore   *db.MongoStore
	Trie         *trie.Trie
	Notifier     *websocket.Notifier
}

func NewContext(key string, redis *session.RedisStore, mongo *db.MongoStore, trie *trie.Trie, notifier *websocket.Notifier) *Context {
	return &Context{
		Key:          key,
		SessionStore: redis,
		MongoStore:   mongo,
		Trie:         trie,
		Notifier:     notifier,
	}
}

func (ctx *Context) startNewSession(w http.ResponseWriter, i interface{}) error {
	ns := SessionState{time.Now(), i}
	if _, err := session.BeginSession(ctx.Key, ctx.SessionStore, ns, w); err != nil {
		return err
	} else {
		return nil
	}
}