package auth

import (
	"time"

	"netplay-platform/backend/internal/db"
)

const testSessionDuration = 24 * time.Hour

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t testing.TB) *db.Database {
	database, err := db.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return database
}
