package handler

import (
	"time"
)

// SessionState tracks when the session was started and who
// started the session.
type SessionState struct {
	SessionStart time.Time   `json:"sessionStart"`
	Interface    interface{} `json:"state"`
}

// NewSessionState creates a new session state given the start time
// and user
func NewSessionState(sessionStart time.Time, i interface{}) *SessionState {
	return &SessionState{sessionStart, i}
}
