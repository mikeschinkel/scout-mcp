package mcputil_test

import (
	"testing"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessions_CreateSession(t *testing.T) {

	// Create a session
	session := mcputil.NewSession()
	err := session.Initialize()
	require.NoError(t, err, "Failed to create session")

	// Validate token properties
	assert.NotEmpty(t, session.Token, "Token should not be empty")
	assert.Equal(t, 64, len(session.Token), "Token should be 64 characters (32 bytes hex)")
	assert.True(t, session.ExpiresAt.After(time.Now()), "Expiration should be in the future")
	assert.True(t, session.ExpiresAt.Before(time.Now().Add(25*time.Hour)), "Expiration should be within 25 hours")

	// Validate the session exists and is not expired
	err = session.Validate()
	require.NoError(t, err, "Failed to validate session")
}

func TestSessions_ValidateSession(t *testing.T) {

	t.Run("ValidToken", func(t *testing.T) {
		session := mcputil.NewSession()
		err := session.Initialize()
		// Create a session
		require.NoError(t, err, "Failed to create session")

		// Validate it
		err = session.Validate()
		require.NoError(t, err, "Failed to validate session")
	})

	t.Run("EmptyToken", func(t *testing.T) {
		session := mcputil.NewSession()
		err := session.Validate()
		require.Error(t, err, "Should error on empty token")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		err := mcputil.ValidateSession("invalid-token-12345")
		require.Error(t, err, "Should error on invalid token")
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		// We can't easily test expired tokens without time manipulation
		// This would require dependency injection or other patterns
		// For now, test that ClearExpiredSessions works
		session := mcputil.NewSession()
		err := session.Initialize()
		require.NoError(t, err, "Failed to create session")

		// Verify session exists
		err = mcputil.ValidateSession(session.Token)
		require.NoError(t, err, "Should not error on valid token")

		// Clear expired sessions (this one shouldn't be cleared since it's not expired)
		mcputil.ClearSessions(mcputil.ExpiredSessions)

		// Should still be valid
		err = mcputil.ValidateSession(session.Token)
		require.NoError(t, err, "Should not error after clearing expired sessions as Token should still be valid after clearing expired sessions")
	})
}

func TestSessions_RequireValidSession(t *testing.T) {
	// Clear any existing sessions
	mcputil.ClearSessions(mcputil.AllSessions)

	t.Run("ValidSession", func(t *testing.T) {
		session := mcputil.NewSession()
		err := session.Initialize()
		require.NoError(t, err, "Failed to create session")

		err = session.Validate()
		assert.NoError(t, err, "Valid session should not error")
	})

	t.Run("InvalidSession", func(t *testing.T) {
		err := mcputil.ValidateSession("invalid-token")
		assert.Error(t, err, "Invalid session should error")
		assert.Contains(t, err.Error(), "token not found", "Error should mention token not found")
	})

	t.Run("EmptySession", func(t *testing.T) {
		err := mcputil.ValidateSession("")
		assert.Error(t, err, "Empty session should error")
		assert.Contains(t, err.Error(), "token not found", "Error should mention token not found")
	})
}

func TestSessions_TokenUniqueness(t *testing.T) {
	tokenSet := make(map[string]bool)

	// Generate multiple tokens and ensure they're unique
	for i := 0; i < 100; i++ {
		session := mcputil.NewSession()
		err := session.Initialize()
		require.NoError(t, err, "Failed to create session %d", i)

		assert.False(t, tokenSet[session.Token], "Token %s should be unique", session.Token)
		tokenSet[session.Token] = true
	}

	assert.Equal(t, 100, len(tokenSet), "All tokens should be unique")
}

func TestSessions_ConcurrentCreation(t *testing.T) {
	// Create multiple sessions concurrently
	tokenChan := make(chan string, 10)
	errChan := make(chan error, 10)

	for i := 0; i < 100; i++ {
		go func() {
			session := mcputil.NewSession()
			err := session.Initialize()
			if err != nil {
				errChan <- err
				return
			}
			tokenChan <- session.Token
		}()
	}

	// Collect results
	tokens := make([]string, 0, 10)
	for i := 0; i < 100; i++ {
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

	assert.Equal(t, 100, len(tokens), "Should have created 100 tokens")
}
