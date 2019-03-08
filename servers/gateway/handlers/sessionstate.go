package handlers

import (
	"questionqueue/servers/gateway/models/users"
	"time"
)

// SessionState tracks when the session was started and who
// started the session.
type SessionState struct {
	SessionStart time.Time  `json:"sessionStart"`
	User         users.User `json:"user"`
}

// NewSessionState creates a new session state given the start time
// and user
func NewSessionState(sessionStart time.Time, user *users.User) *SessionState {
	return &SessionState{sessionStart, *user}
}
