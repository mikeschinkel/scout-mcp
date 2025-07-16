package mcptools

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Sessions handles session token creation and validation
type Sessions struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// Session contains information about an active session
type Session struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	LastUsed  time.Time `json:"last_used"`
}

// SessionResponse contains the complete response from start_session
type SessionResponse struct {
	SessionToken   string             `json:"session_token"`
	TokenExpiresAt time.Time          `json:"token_expires_at"`
	ToolHelp       string             `json:"tool_help"`
	ServerConfig   map[string]any     `json:"server_config"`
	Instructions   InstructionsConfig `json:"instructions"`
	Message        string             `json:"message"`
}

// InstructionsConfig contains all instruction content
type InstructionsConfig struct {
	General           string                 `json:"general"`
	Languages         []LanguageInstructions `json:"languages"`
	ExtensionMappings map[string]string      `json:"extension_mappings"`
}

// LanguageInstructions contains instructions for a specific language
type LanguageInstructions struct {
	Language   string   `json:"language"`          // "python", "go", "javascript"
	Version    string   `json:"version,omitempty"` // "2", "3", "5.7", "8.4", etc.
	Content    string   `json:"content"`           // Markdown content
	Extensions []string `json:"extensions"`        // File extensions this applies to
}

var (
	sessions *Sessions
	once     sync.Once
)

// GetSessions returns the singleton session manager
func GetSessions() *Sessions {
	once.Do(func() {
		sessions = &Sessions{
			sessions: make(map[string]*Session),
		}
	})
	return sessions
}

// CreateSession creates a new session and returns the token
func (sm *Sessions) CreateSession() (token string, expiresAt time.Time, err error) {
	var tokenBytes []byte

	// Generate random token
	tokenBytes = make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		goto end
	}

	token = hex.EncodeToString(tokenBytes)
	expiresAt = time.Now().Add(24 * time.Hour)

	// Store session info
	sm.mutex.Lock()
	sm.sessions[token] = &Session{
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		LastUsed:  time.Now(),
	}
	sm.mutex.Unlock()

end:
	return token, expiresAt, err
}

// ValidateSession checks if a session token is valid and updates last used time
func (sm *Sessions) ValidateSession(token string) (valid bool, err error) {
	var session *Session
	var exists bool

	if token == "" {
		goto end
	}

	sm.mutex.RLock()
	session, exists = sm.sessions[token]
	sm.mutex.RUnlock()

	if !exists {
		goto end
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		// Remove expired token
		sm.mutex.Lock()
		delete(sm.sessions, token)
		sm.mutex.Unlock()
		goto end
	}

	// Update last used time
	sm.mutex.Lock()
	session.LastUsed = time.Now()
	sm.mutex.Unlock()

	valid = true

end:
	return valid, err
}

// ClearExpiredSessions removes all expired sessions
func (sm *Sessions) ClearExpiredSessions() {
	var now time.Time
	var expiredTokens []string

	now = time.Now()
	expiredTokens = make([]string, 0)

	sm.mutex.RLock()
	for token, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			expiredTokens = append(expiredTokens, token)
		}
	}
	sm.mutex.RUnlock()

	if len(expiredTokens) > 0 {
		sm.mutex.Lock()
		for _, token := range expiredTokens {
			delete(sm.sessions, token)
		}
		sm.mutex.Unlock()
	}
}

// ClearAllSessions removes all sessions (used on server restart)
func (sm *Sessions) ClearAllSessions() {
	sm.mutex.Lock()
	sm.sessions = make(map[string]*Session)
	sm.mutex.Unlock()
}

// GetSessionCount returns the number of active sessions
func (sm *Sessions) GetSessionCount() int {
	var count int

	sm.mutex.RLock()
	count = len(sm.sessions)
	sm.mutex.RUnlock()

	return count
}

// RequireValidSession validates a session token and returns an error if invalid
func RequireValidSession(token string) (err error) {
	var valid bool
	var sm *Sessions

	sm = GetSessions()
	valid, err = sm.ValidateSession(token)
	if err != nil {
		goto end
	}

	if !valid {
		err = fmt.Errorf("invalid or expired session token. Call 'start_session' first to start an MCP session and get instructions on how to correctly use this MCP Server")
		goto end
	}

end:
	return err
}
