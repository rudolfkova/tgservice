// Package model ...
package model

// SessionStatus ...
type SessionStatus string

const (
	StatusPendingAuth SessionStatus = "pending_auth"
	StatusReady       SessionStatus = "ready"
)

// Session ...
type Session struct {
	ID     string
	Status SessionStatus
}
