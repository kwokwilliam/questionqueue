package store

import (
	"errors"
)

//ErrStateNotFound is returned from Store.Get() when the requested
//session id was not found in the store
var ErrStateNotFound = errors.New("no session state was found in the session store")

//Store represents a session data store.
//This is an abstract interface that can be implemented
//against several different types of data stores. For example,
//session data could be stored in memory in a concurrent map,
//or more typically in a shared key/value server store like redis.
type Store interface {
	// GetCurrentQueue gets the current queue
	// In the future this can be changed to manage more than one queue
	GetCurrentQueue() (*QuestionQueue, error)
}
