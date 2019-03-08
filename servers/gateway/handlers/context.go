package handlers

import (
	"errors"
	"questionqueue/servers/gateway/indexes"
	"questionqueue/servers/gateway/models/users"
	"questionqueue/servers/gateway/sessions"
)

// HandlerContext is a struct that will be a receiver on any
// HTTP handler functions that need access to globals, such
// as the key used for signing and verifying SessionIDs,
// the session store and the user store
type HandlerContext struct {
	SigningKey   string
	SessionStore sessions.Store
	userStore    users.Store
	sessionID    sessions.SessionID
	trie         *indexes.Trie
	Notifier     *Notifier
}

// NewHandlerContext constructs a new HandlerContext,
// ensuring that the dependencies are valid values
// It creates an empty sessionID
func NewHandlerContext(SigningKey string, SessionStore sessions.Store, userStore users.Store) (*HandlerContext, error) {
	if SessionStore != nil && userStore != nil {
		return &HandlerContext{SigningKey, SessionStore, userStore, "", indexes.NewTrie(), &Notifier{}}, nil
	}
	return nil, errors.New("Unable to find session store or user store")
}

// SetSessionID sets the session id
func (ctx *HandlerContext) SetSessionID(sid sessions.SessionID) {
	ctx.sessionID = sid
}

// DeleteCurrentSession will delete the current session stored in context
func (ctx *HandlerContext) DeleteCurrentSession() error {
	return ctx.SessionStore.Delete(ctx.sessionID)
}

// InitiateTrie initiates the user trie
func (ctx *HandlerContext) InitiateTrie() error {
	if err := ctx.userStore.LoadUsersToTrie(ctx.trie); err != nil {
		return err
	}
	return nil
}

// CurrentUser gets the current user
func (ctx *HandlerContext) CurrentUser() (*users.User, error) {
	currentSession := &SessionState{}
	if err := ctx.SessionStore.Get(ctx.sessionID, currentSession); err != nil {
		return nil, err
	}
	return &currentSession.User, nil
}

// initialize user 1
// initialize user 2
//  use auth token to query for user
//	change first/last name as second user
// 		ERRORS HAPPEN INTERNALLY in the remove function (removeSplitToTrie/removeNamesInTrie)
//	try to query for second user with name
//		because the removal failed, it did not continue with the addition of the new names
//		but it did update in database.
