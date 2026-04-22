package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	"netplay-platform/backend/internal/auth"
	"netplay-platform/backend/internal/config"
	"netplay-platform/backend/internal/db"
	"netplay-platform/backend/internal/lobby"
	"netplay-platform/backend/internal/netplay"
)

func main() {
	// Load configuration
	cfg := config.Load()
	logger := config.NewLogger(cfg.LogLevel)

	logger.Info("Starting netplay server", "port", cfg.ServerPort, "tick_rate", cfg.TickRate)

	// Initialize database
	database, err := db.NewDatabase(cfg.DBPath)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize services
	authService := auth.NewService(database, 24*time.Hour)
	lobbyManager := lobby.NewManager(database)
	netplayHub := netplay.NewHub(cfg.TickRate)

	// Start netplay hub in background
	go netplayHub.Run()

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", handleRegister(authService, logger))
		r.Post("/login", handleLogin(authService, logger))
		r.Post("/logout", handleLogout(authService, logger))
		r.Get("/me", handleGetMe(authService, logger))
	})

	// Lobby routes (require auth)
	r.Route("/api/lobbies", func(r chi.Router) {
		r.Use(authMiddleware(authService))
		r.Get("/", handleGetLobbies(lobbyManager, logger))
		r.Post("/", handleCreateLobby(lobbyManager, authService, logger))
		r.Get("/{lobbyID}", handleGetLobby(lobbyManager, logger))
		r.Post("/{lobbyID}/join", handleJoinLobby(lobbyManager, netplayHub, logger))
		r.Post("/{lobbyID}/leave", handleLeaveLobby(lobbyManager, netplayHub, logger))
		r.Post("/{lobbyID}/ready", handleSetReady(lobbyManager, netplayHub, logger))
		r.Post("/{lobbyID}/start", handleStartLobby(lobbyManager, netplayHub, logger))
	})

	// Leaderboard routes
	r.Route("/api/leaderboard", func(r chi.Router) {
		r.Get("/", handleGetLeaderboard(database, logger))
		r.Get("/me", authMiddleware(authService), handleGetMyStats(database, logger))
	})

	// WebSocket endpoint
	r.HandleFunc("/ws", handleWebSocket(netplayHub, authService, lobbyManager, logger))

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error", "error", err)
		}
	}()

	// Start session cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			authService.CleanupExpiredSessions()
		}
	}()

	// Start server
	logger.Info("Server listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped")
}

// authMiddleware extracts and validates the session token from the Authorization header
func authMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Remove "Bearer " prefix if present
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			session, err := authService.ValidateSession(token)
			if err != nil {
				http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
				return
			}

			// Add session to context
			ctx := context.WithValue(r.Context(), "session", session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getSession retrieves the session from context
func getSession(r *http.Request) *auth.Session {
	session, ok := r.Context().Value("session").(*auth.Session)
	if !ok {
		return nil
	}
	return session
}

// Handler functions are defined in handlers.go
// Suppress unused import warning for sqlite3 driver
var _ = "github.com/mattn/go-sqlite3"
