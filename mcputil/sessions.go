package mcputil

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Session contains information about an active session
type Session struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	LastUsed  time.Time `json:"last_used"`
}

type Payload interface {
	Payload()
	Initialize(Tool, ToolRequest) error
}

var _ PayloadCarrier = (*StartSessionResult)(nil)

// StartSessionResult contains the generic response structure for session creation
type StartSessionResult struct {
	SessionToken    string    `json:"session_token"`
	TokenExpiresAt  time.Time `json:"token_expires_at"`
	Instructions    string    `json:"instructions"`
	PayloadTypeName string    `json:"payload_type"`
	Message         string    `json:"message"`
	Payload         Payload   `json:"payload"`
}

func (ssr *StartSessionResult) SetPayload(p Payload) {
	ssr.Payload = p
}

func (ssr *StartSessionResult) GetPayloadTypeName() string {
	return ssr.PayloadTypeName
}

// Package-level session storage
var (
	sessions      = make(map[string]*Session)
	sessionsMutex sync.RWMutex
)

// NewSession creates a new session and returns the Session instance
func NewSession() (session *Session) {
	return &Session{}
}

// GetSession retrieves a session by token
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

// Initialize session
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

// Validate checks if this session is valid and updates last used time
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

type SessionClearType int

const (
	ExpiredSessions SessionClearType = iota + 1
	AllSessions
)

// ClearSessions removes sessions based on the specified type
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

// clearExpiredSessions removes all expired sessions (private)
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

// GetSessionCount returns the number of active sessions
func GetSessionCount() int {
	var count int

	sessionsMutex.RLock()
	count = len(sessions)
	sessionsMutex.RUnlock()

	return count
}

// ValidateSession validates a session token and returns an error if invalid
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
