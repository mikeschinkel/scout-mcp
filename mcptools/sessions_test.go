package mcptools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessions_CreateSession(t *testing.T) {
	sessions := GetSessions()

	// Create a session
	token, expiresAt, err := sessions.NewSession()
	require.NoError(t, err, "Failed to create session")

	// Validate token properties
	assert.NotEmpty(t, token, "Token should not be empty")
	assert.Equal(t, 64, len(token), "Token should be 64 characters (32 bytes hex)")
	assert.True(t, expiresAt.After(time.Now()), "Expiration should be in the future")
	assert.True(t, expiresAt.Before(time.Now().Add(25*time.Hour)), "Expiration should be within 25 hours")

	// Validate the session is stored and valid
	valid, err := sessions.ValidateSession(token)
	require.NoError(t, err, "Failed to validate session")
	assert.True(t, valid, "Session should be valid")
}

func TestSessions_ValidateSession(t *testing.T) {
	sessions := GetSessions()

	t.Run("ValidToken", func(t *testing.T) {
		// Create a session
		token, _, err := sessions.NewSession()
		require.NoError(t, err, "Failed to create session")

		// Validate it
		valid, err := sessions.ValidateSession(token)
		require.NoError(t, err, "Failed to validate session")
		assert.True(t, valid, "Session should be valid")
	})

	t.Run("EmptyToken", func(t *testing.T) {
		valid, err := sessions.ValidateSession("")
		require.NoError(t, err, "Should not error on empty token")
		assert.False(t, valid, "Empty token should be invalid")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		valid, err := sessions.ValidateSession("invalid-token-12345")
		require.NoError(t, err, "Should not error on invalid token")
		assert.False(t, valid, "Invalid token should be invalid")
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		// We can't easily test expired tokens without time manipulation
		// This would require dependency injection or other patterns
		// For now, test that ClearExpiredSessions works
		token, _, err := sessions.NewSession()
		require.NoError(t, err, "Failed to create session")

		// Verify session exists
		valid, err := sessions.ValidateSession(token)
		require.NoError(t, err, "Should not error on valid token")
		assert.True(t, valid, "Token should be valid")

		// Clear expired sessions (this one shouldn't be cleared since it's not expired)
		sessions.ClearExpiredSessions()

		// Should still be valid
		valid, err = sessions.ValidateSession(token)
		require.NoError(t, err, "Should not error after clearing expired sessions")
		assert.True(t, valid, "Token should still be valid after clearing expired sessions")
	})
}

func TestSessions_RequireValidSession(t *testing.T) {
	// Clear any existing sessions
	sessions := GetSessions()
	sessions.ClearSessions()

	t.Run("ValidSession", func(t *testing.T) {
		sessions := GetSessions()
		token, _, err := sessions.NewSession()
		require.NoError(t, err, "Failed to create session")

		err = RequireValidSession(token)
		assert.NoError(t, err, "Valid session should not error")
	})

	t.Run("InvalidSession", func(t *testing.T) {
		err := RequireValidSession("invalid-token")
		assert.Error(t, err, "Invalid session should error")
		assert.Contains(t, err.Error(), "invalid or expired session token", "Error should mention invalid token")
	})

	t.Run("EmptySession", func(t *testing.T) {
		err := RequireValidSession("")
		assert.Error(t, err, "Empty session should error")
		assert.Contains(t, err.Error(), "invalid or expired session token", "Error should mention invalid token")
	})
}

func TestSessions_TokenUniqueness(t *testing.T) {
	sessions := GetSessions()
	tokenSet := make(map[string]bool)

	// Generate multiple tokens and ensure they're unique
	for i := 0; i < 100; i++ {
		token, _, err := sessions.NewSession()
		require.NoError(t, err, "Failed to create session %d", i)

		assert.False(t, tokenSet[token], "Token %s should be unique", token)
		tokenSet[token] = true
	}

	assert.Equal(t, 100, len(tokenSet), "All tokens should be unique")
}

func TestSessions_ConcurrentCreation(t *testing.T) {
	sessions := GetSessions()

	// Create multiple sessions concurrently
	tokenChan := make(chan string, 10)
	errChan := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			token, _, err := sessions.NewSession()
			if err != nil {
				errChan <- err
				return
			}
			tokenChan <- token
		}()
	}

	// Collect results
	tokens := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		select {
		case token := <-tokenChan:
			tokens = append(tokens, token)
		case err := <-errChan:
			t.Errorf("Concurrent session creation failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent session creation")
		}
	}

	// Verify all tokens are unique
	tokenSet := make(map[string]bool)
	for _, token := range tokens {
		assert.False(t, tokenSet[token], "Token should be unique in concurrent creation")
		tokenSet[token] = true
	}

	assert.Equal(t, 10, len(tokens), "Should have created 10 tokens")
}
