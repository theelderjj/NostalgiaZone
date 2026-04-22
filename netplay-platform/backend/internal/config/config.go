package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	ServerPort    int
	DBPath        string
	TickRate      int // Frames per second for netplay sync
	CORSOrigins   []string
	LogLevel      string
	SessionSecret string
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	cfg := &Config{
		ServerPort:    8080,
		DBPath:        "./data/netplay.db",
		TickRate:      60,
		CORSOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		LogLevel:      "info",
		SessionSecret: generateRandomSecret(),
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.ServerPort = p
		}
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}

	if tickRate := os.Getenv("TICK_RATE"); tickRate != "" {
		if t, err := strconv.Atoi(tickRate); err == nil && t > 0 {
			cfg.TickRate = t
		}
	}

	if cors := os.Getenv("CORS_ORIGINS"); cors != "" {
		cfg.CORSOrigins = splitString(cors, ",")
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	if secret := os.Getenv("SESSION_SECRET"); secret != "" {
		cfg.SessionSecret = secret
	}

	return cfg
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	return result
}

func generateRandomSecret() string {
	// In production, use crypto/rand for better randomness
	return "change-me-in-production-" + os.Getenv("HOSTNAME")
}

// Logger is a simple structured logger
type Logger struct {
	level string
}

// NewLogger creates a new logger instance
func NewLogger(level string) *Logger {
	return &Logger{level: level}
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...interface{}) {
	if l.level != "debug" && l.level != "info" {
		return
	}
	log.Printf("[INFO] "+msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...interface{}) {
	log.Printf("[ERROR] "+msg, fields...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...interface{}) {
	if l.level != "debug" {
		return
	}
	log.Printf("[DEBUG] "+msg, fields...)
}
