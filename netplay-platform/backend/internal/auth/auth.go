package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"netplay-platform/backend/internal/db"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username already exists")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
)

// Session represents an authenticated user session
type Session struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Service handles authentication logic
type Service struct {
	db       *db.Database
	sessions map[string]*Session
	mu       sync.RWMutex
	duration time.Duration
}

// NewService creates a new authentication service
func NewService(database *db.Database, sessionDuration time.Duration) *Service {
	return &Service{
		db:       database,
		sessions: make(map[string]*Session),
		duration: sessionDuration,
	}
}

// Register creates a new user account
func (s *Service) Register(username, password string) (*db.User, error) {
	// Validate input
	if len(username) < 3 || len(username) > 32 {
		return nil, errors.New("username must be between 3 and 32 characters")
	}
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Check if username exists
	existing, err := s.db.GetUserByUsername(username)
	if err == nil && existing != nil {
		return nil, ErrUsernameTaken
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := s.db.CreateUser(username, string(hash))
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and creates a session
func (s *Service) Login(username, password string) (*Session, error) {
	// Get user
	user, err := s.db.GetUserByUsername(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Create session
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	session := &Session{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.duration),
	}

	s.mu.Lock()
	s.sessions[token] = session
	s.mu.Unlock()

	// Update last active
	_ = s.db.UpdateLastActive(user.ID)

	return session, nil
}

// Logout invalidates a session
func (s *Service) Logout(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[token]; !exists {
		return ErrSessionNotFound
	}

	delete(s.sessions, token)
	return nil
}

// ValidateSession checks if a session is valid and returns it
func (s *Service) ValidateSession(token string) (*Session, error) {
	s.mu.RLock()
	session, exists := s.sessions[token]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		s.mu.Lock()
		delete(s.sessions, token)
		s.mu.Unlock()
		return nil, ErrSessionExpired
	}

	return session, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(userID int64) (*db.User, error) {
	return s.db.GetUserByID(userID)
}

// CleanupExpiredSessions removes expired sessions from memory
func (s *Service) CleanupExpiredSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, token)
		}
	}
}

// generateToken creates a random session token
func generateToken() (string, error) {

// HashPassword hashes a password using bcrypt
// Exported for testing purposes
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckPasswordHash compares a password with its hash
// Exported for testing purposes
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Middleware returns a function that can be used as HTTP middleware
// It extracts the session token from the Authorization header
func (s *Service) Middleware(next func(*Session) interface{}) func(token string) (interface{}, error) {
	return func(token string) (interface{}, error) {
		session, err := s.ValidateSession(token)
		if err != nil {
			return nil, err
		}
		return next(session)
	}
}
