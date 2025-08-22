package mcputil

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Session contains information about an active MCP session including
// authentication token, creation time, expiration, and usage tracking.
// Sessions provide secure access control for MCP tool operations.
type Session struct {
	Token     string    `json:"token"`      // Cryptographically secure session token
	CreatedAt time.Time `json:"created_at"` // When the session was created
	ExpiresAt time.Time `json:"expires_at"` // When the session expires (24 hours from creation)
	LastUsed  time.Time `json:"last_used"`  // When the session was last accessed
}

// Payload defines the interface for session payload data that can be
// attached to session creation responses. Payloads allow tools to
// include additional configuration or context data in session responses.
type Payload interface {
	Payload() // Marker method
	Initialize(Tool, ToolRequest) error
}

var _ PayloadCarrier = (*StartSessionResult)(nil)

// StartSessionResult contains the response structure for session creation.
// This includes the session token, expiration information, user instructions,
// and optional payload data for tool-specific context.
type StartSessionResult struct {
	SessionToken    string    `json:"session_token"`    // Generated session token for authentication
	TokenExpiresAt  time.Time `json:"token_expires_at"` // When the token expires
	Instructions    string    `json:"instructions"`     // User instructions for using MCP tools
	PayloadTypeName string    `json:"payload_type"`     // Type name of the payload for deserialization
	Message         string    `json:"message"`          // Success message for the user
	Payload         Payload   `json:"payload"`          // Optional payload data
}

// SetPayload sets the payload data for the session start response.
// This method allows attaching additional tool-specific context to session creation.
func (ssr *StartSessionResult) SetPayload(p Payload) {
	ssr.Payload = p
}

// GetPayloadTypeName returns the type name of the payload for deserialization.
// This method provides the payload type information needed for proper unmarshaling.
func (ssr *StartSessionResult) GetPayloadTypeName() string {
	return ssr.PayloadTypeName
}

// Package-level session storage
var (
	sessions      = make(map[string]*Session)
	sessionsMutex sync.RWMutex
)

// NewSession creates a new session and returns the Session instance.
// This function creates an uninitialized session that must be initialized before use.
func NewSession() (session *Session) {
	return &Session{}
}

// GetSession retrieves a session by token from the session store.
// Returns nil and false if the token is empty or not found in the store.
func GetSession(token string) (session *Session, exists bool) {
	if token == "" {
		goto end
	}

	sessionsMutex.RLock()
	session, exists = sessions[token]
	sessionsMutex.RUnlock()

end:
	return session, exists
}

// Initialize initializes the session with a cryptographically secure token and timestamps.
// This method generates a random 32-byte token, sets creation and expiration times,
// and stores the session in the global session store.
func (s *Session) Initialize() (err error) {
	var tokenBytes []byte

	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	_, ok := sessions[s.Token]
	if ok {
		goto end
	}

	// Generate random token
	tokenBytes = make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		goto end
	}

	s.Token = hex.EncodeToString(tokenBytes)
	s.CreatedAt = time.Now()
	s.ExpiresAt = time.Now().Add(24 * time.Hour)
	s.LastUsed = time.Now()

	// Store session info
	sessions[s.Token] = s

end:
	return err
}

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrNoPayloadType = errors.New("unable to get payload type; you might need to call mcputil.RegisterPayloadType() first")
)

// Validate checks if this session is valid and updates last used time.
// This method verifies the session exists, hasn't expired, and updates the last used timestamp.
func (s *Session) Validate() (err error) {
	var ok bool

	if s == nil {
		goto end
	}
	_, ok = sessions[s.Token]
	if !ok {
		err = ErrTokenNotFound
		goto end
	}

	// Check if expired
	if time.Now().After(s.ExpiresAt) {
		// Remove expired token
		sessionsMutex.Lock()
		delete(sessions, s.Token)
		sessionsMutex.Unlock()
		err = ErrTokenExpired
		goto end
	}

	// Update last used time
	sessionsMutex.Lock()
	s.LastUsed = time.Now()
	sessionsMutex.Unlock()

end:
	return err
}

// SessionClearType specifies which sessions to clear from the session store.
// This is used with the ClearSessions function to control session cleanup behavior.
type SessionClearType int

const (
	// ExpiredSessions clears only sessions that have expired based on their ExpiresAt timestamp.
	// Active sessions within their validity period are preserved.
	ExpiredSessions SessionClearType = iota + 1

	// AllSessions clears all sessions regardless of their expiration status.
	// This is typically used during server shutdown or testing cleanup.
	AllSessions
)

// ClearSessions removes sessions from the session store based on the specified type.
// This function provides controlled cleanup of expired or all sessions.
func ClearSessions(which SessionClearType) {
	switch which {
	case ExpiredSessions:
		clearExpiredSessions()
	case AllSessions:
		sessionsMutex.Lock()
		sessions = make(map[string]*Session)
		sessionsMutex.Unlock()
	default:
		panic(fmt.Sprintf("Unsupported session clear type: %d", which))
	}
}

// clearExpiredSessions removes all expired sessions from the session store.
// This is an internal function called by ClearSessions when ExpiredSessions is specified.
// It safely removes sessions that have passed their expiration timestamp.
func clearExpiredSessions() {
	var now time.Time
	var expiredTokens []string

	now = time.Now()
	expiredTokens = make([]string, 0)

	sessionsMutex.RLock()
	for token, session := range sessions {
		if now.After(session.ExpiresAt) {
			expiredTokens = append(expiredTokens, token)
		}
	}
	sessionsMutex.RUnlock()

	if len(expiredTokens) > 0 {
		sessionsMutex.Lock()
		for _, token := range expiredTokens {
			delete(sessions, token)
		}
		sessionsMutex.Unlock()
	}
}

// GetSessionCount returns the number of active sessions in the session store.
// This function provides a way to monitor session usage and memory consumption.
func GetSessionCount() int {
	var count int

	sessionsMutex.RLock()
	count = len(sessions)
	sessionsMutex.RUnlock()

	return count
}

// ValidateSession validates a session token and returns an error if invalid.
// This is a convenience function that combines session lookup and validation.
func ValidateSession(token string) (err error) {
	var session *Session
	var exists bool

	session, exists = GetSession(token)
	if !exists {
		err = ErrTokenNotFound
		goto end
	}

	err = session.Validate()
	if err != nil {
		goto end
	}

end:
	return err
}
