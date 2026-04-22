package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"netplay-platform/backend/internal/db"
)

func TestService_Register(t *testing.T) {
	database := setupTestDB(t)
	service := NewService(database, testSessionDuration)

	t.Run("valid registration", func(t *testing.T) {
		user, err := service.Register("testuser", "password123")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.NotEmpty(t, user.PasswordHash)
	})

	t.Run("username too short", func(t *testing.T) {
		user, err := service.Register("ab", "password123")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("password too short", func(t *testing.T) {
		user, err := service.Register("testuser2", "12345")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("duplicate username", func(t *testing.T) {
		_, err := service.Register("duplicate", "password123")
		require.NoError(t, err)

		user, err := service.Register("duplicate", "password456")
		assert.Error(t, err)
		assert.Equal(t, ErrUsernameTaken, err)
		assert.Nil(t, user)
	})
}

func TestService_Login(t *testing.T) {
	database := setupTestDB(t)
	service := NewService(database, testSessionDuration)

	// Create user first
	_, err := service.Register("loginuser", "password123")
	require.NoError(t, err)

	t.Run("valid credentials", func(t *testing.T) {
		session, err := service.Login("loginuser", "password123")
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "loginuser", session.Username)
		assert.NotEmpty(t, session.Token)
	})

	t.Run("invalid password", func(t *testing.T) {
		session, err := service.Login("loginuser", "wrongpassword")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		assert.Nil(t, session)
	})

	t.Run("nonexistent user", func(t *testing.T) {
		session, err := service.Login("nonexistent", "password123")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidCredentials, err)
		assert.Nil(t, session)
	})
}

func TestService_ValidateSession(t *testing.T) {
	database := setupTestDB(t)
	service := NewService(database, testSessionDuration)

	// Create and login user
	_, err := service.Register("sessionuser", "password123")
	require.NoError(t, err)

	session, err := service.Login("sessionuser", "password123")
	require.NoError(t, err)

	t.Run("valid session", func(t *testing.T) {
		validated, err := service.ValidateSession(session.Token)
		require.NoError(t, err)
		assert.Equal(t, session.UserID, validated.UserID)
		assert.Equal(t, session.Username, validated.Username)
	})

	t.Run("invalid token", func(t *testing.T) {
		validated, err := service.ValidateSession("invalid-token")
		assert.Error(t, err)
		assert.Equal(t, ErrSessionNotFound, err)
		assert.Nil(t, validated)
	})

	t.Run("logout invalidates session", func(t *testing.T) {
		err := service.Logout(session.Token)
		require.NoError(t, err)

		validated, err := service.ValidateSession(session.Token)
		assert.Error(t, err)
		assert.Equal(t, ErrSessionNotFound, err)
		assert.Nil(t, validated)
	})
}

func TestService_CleanupExpiredSessions(t *testing.T) {
	database := setupTestDB(t)
	// Use very short session duration for testing
	service := NewService(database, 0) // Already expired

	// Create and login user
	_, err := service.Register("cleanupuser", "password123")
	require.NoError(t, err)

	session, err := service.Login("cleanupuser", "password123")
	require.NoError(t, err)

	// Session should be immediately expired
	validated, err := service.ValidateSession(session.Token)
	assert.Error(t, err)
	assert.Equal(t, ErrSessionExpired, err)

	// Cleanup should remove it
	service.CleanupExpiredSessions()
	
	validated, err = service.ValidateSession(session.Token)
	assert.Error(t, err)
	assert.Equal(t, ErrSessionNotFound, err)
}
