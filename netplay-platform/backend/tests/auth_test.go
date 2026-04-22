package tests

import (
	"testing"
	"time"

	"netplay-platform/backend/internal/auth"
	"netplay-platform/backend/internal/db"
)

func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"
	
	// Test that hashing works
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	// Test that verification works
	if !auth.CheckPasswordHash(password, hash) {
		t.Error("Password verification failed for correct password")
	}
	
	// Test that wrong password fails
	if auth.CheckPasswordHash("wrongpassword", hash) {
		t.Error("Password verification succeeded for wrong password")
	}
}

func TestSessionCreation(t *testing.T) {
	// Create in-memory database for testing
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()
	
	authService := auth.NewService(database, 24*time.Hour)
	
	// Register a test user
	user, err := authService.Register("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	
	// Login and get session
	session, err := authService.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	
	// Verify session properties
	if session.UserID != user.ID {
		t.Errorf("Session user ID mismatch: got %d, want %d", session.UserID, user.ID)
	}
	
	if session.Username != "testuser" {
		t.Errorf("Session username mismatch: got %s, want testuser", session.Username)
	}
	
	// Verify session expiration is set correctly
	expectedExpiry := time.Now().Add(24 * time.Hour)
	timeDiff := session.ExpiresAt.Sub(expectedExpiry)
	if timeDiff < -time.Second || timeDiff > time.Second {
		t.Errorf("Session expiry time incorrect: got %v, want ~%v", session.ExpiresAt, expectedExpiry)
	}
}

func TestInvalidCredentials(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()
	
	authService := auth.NewService(database, 24*time.Hour)
	
	// Register a user
	_, err = authService.Register("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	
	// Try login with wrong password
	_, err = authService.Login("testuser", "wrongpassword")
	if err == nil {
		t.Error("Login succeeded with wrong password")
	}
	
	// Try login with non-existent user
	_, err = authService.Login("nonexistent", "password123")
	if err == nil {
		t.Error("Login succeeded with non-existent user")
	}
}

func TestUsernameUniqueness(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()
	
	authService := auth.NewService(database, 24*time.Hour)
	
	// Register first user
	_, err = authService.Register("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to register first user: %v", err)
	}
	
	// Try to register same username
	_, err = authService.Register("testuser", "differentpassword")
	if err == nil {
		t.Error("Registration succeeded with duplicate username")
	}
}

func TestSessionValidation(t *testing.T) {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()
	
	authService := auth.NewService(database, 24*time.Hour)
	
	// Register and login
	_, err = authService.Register("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	
	session, err := authService.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	
	// Validate the session
	validSession, err := authService.ValidateSession(session.Token)
	if err != nil {
		t.Fatalf("Failed to validate valid session: %v", err)
	}
	
	if validSession.UserID != session.UserID {
		t.Error("Validated session has wrong user ID")
	}
	
	// Try to validate invalid token
	_, err = authService.ValidateSession("invalid-token")
	if err == nil {
		t.Error("Validation succeeded for invalid token")
	}
}
