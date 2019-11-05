package types

import "time"

// Session represents a deployinator session.
type Session struct {
	SessionID string    `json:"sessionID" yaml:"sessionID"`
	Username  string    `json:"username" yaml:"username"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
	User      *User     `json:"user" yaml:"user"`
}
