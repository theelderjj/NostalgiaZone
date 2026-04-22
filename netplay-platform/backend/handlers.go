package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"netplay-platform/backend/internal/auth"
	"netplay-platform/backend/internal/config"
	"netplay-platform/backend/internal/db"
	"netplay-platform/backend/internal/lobby"
	"netplay-platform/backend/internal/netplay"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure appropriately for production
	},
}

// handleRegister handles user registration
func handleRegister(authService *auth.Service, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user, err := authService.Register(req.Username, req.Password)
		if err != nil {
			logger.Error("Registration failed", "error", err, "username", req.Username)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info("User registered", "user_id", user.ID, "username", user.Username)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
		})
	}
}

// handleLogin handles user login
func handleLogin(authService *auth.Service, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		session, err := authService.Login(req.Username, req.Password)
		if err != nil {
			logger.Error("Login failed", "error", err, "username", req.Username)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		logger.Info("User logged in", "user_id", session.UserID, "username", session.Username)
		json.NewEncoder(w).Encode(session)
	}
}

// handleLogout handles user logout
func handleLogout(authService *auth.Service, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		if err := authService.Logout(token); err != nil {
			http.Error(w, "Invalid session", http.StatusBadRequest)
			return
		}

		logger.Info("User logged out", "token", token[:8]+"...")
		w.WriteHeader(http.StatusOK)
	}
}

// handleGetMe returns the current user's info
func handleGetMe(authService *auth.Service, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		if session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":    session.UserID,
			"username":   session.Username,
			"expires_at": session.ExpiresAt,
		})
	}
}

// handleGetLobbies returns all active lobbies
func handleGetLobbies(lobbyManager *lobby.Manager, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lobbies, err := lobbyManager.GetActiveLobbies()
		if err != nil {
			logger.Error("Failed to get lobbies", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		result := make([]map[string]interface{}, 0, len(lobbies))
		for _, l := range lobbies {
			result = append(result, map[string]interface{}{
				"id":           l.Lobby.ID,
				"game_id":      l.Lobby.GameID,
				"game_title":   l.Lobby.GameTitle,
				"host_id":      l.Lobby.HostID,
				"status":       l.Lobby.Status,
				"player_count": l.Lobby.PlayerCount,
				"max_players":  l.Lobby.MaxPlayers,
				"created_at":   l.Lobby.CreatedAt,
			})
		}

		json.NewEncoder(w).Encode(result)
	}
}

// handleCreateLobby creates a new lobby
func handleCreateLobby(lobbyManager *lobby.Manager, authService *auth.Service, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		if session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			GameID    string `json:"game_id"`
			GameTitle string `json:"game_title"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Generate unique lobby ID
		lobbyID := time.Now().Format("20060102150405") + "-" + session.Username

		info, err := lobbyManager.CreateLobby(lobbyID, req.GameID, req.GameTitle, session.UserID, session.Username)
		if err != nil {
			logger.Error("Failed to create lobby", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Info("Lobby created", "lobby_id", lobbyID, "host", session.Username)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           info.Lobby.ID,
			"game_id":      info.Lobby.GameID,
			"game_title":   info.Lobby.GameTitle,
			"host_id":      info.Lobby.HostID,
			"status":       info.Lobby.Status,
			"player_count": info.Lobby.PlayerCount,
		})
	}
}

// handleGetLobby returns details for a specific lobby
func handleGetLobby(lobbyManager *lobby.Manager, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lobbyID := chi.URLParam(r, "lobbyID")

		info, err := lobbyManager.GetLobby(lobbyID)
		if err != nil {
			http.Error(w, "Lobby not found", http.StatusNotFound)
			return
		}

		players, _ := lobbyManager.GetPlayers(lobbyID)
		playerList := make([]map[string]interface{}, 0, len(players))
		for _, p := range players {
			isHost, _ := lobbyManager.IsHost(lobbyID, p.UserID)
			playerList = append(playerList, map[string]interface{}{
				"user_id":    p.UserID,
				"username":   p.Username,
				"player_num": p.PlayerNum,
				"ready":      p.Ready,
				"is_host":    isHost,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           info.Lobby.ID,
			"game_id":      info.Lobby.GameID,
			"game_title":   info.Lobby.GameTitle,
			"host_id":      info.Lobby.HostID,
			"status":       info.Lobby.Status,
			"player_count": info.Lobby.PlayerCount,
			"max_players":  info.Lobby.MaxPlayers,
			"players":      playerList,
		})
	}
}

// handleJoinLobby allows a user to join a lobby
func handleJoinLobby(lobbyManager *lobby.Manager, hub *netplay.Hub, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		if session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		lobbyID := chi.URLParam(r, "lobbyID")

		playerInfo, err := lobbyManager.JoinLobby(lobbyID, session.UserID, session.Username)
		if err != nil {
			logger.Error("Failed to join lobby", "error", err, "user", session.Username)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Initialize lobby in netplay hub
		hub.InitializeLobby(lobbyID)

		logger.Info("Player joined lobby", "lobby_id", lobbyID, "user", session.Username, "player_num", playerInfo.PlayerNum)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"lobby_id":   lobbyID,
			"user_id":    session.UserID,
			"username":   session.Username,
			"player_num": playerInfo.PlayerNum,
			"ready":      playerInfo.Ready,
		})
	}
}

// handleLeaveLobby allows a user to leave a lobby
func handleLeaveLobby(lobbyManager *lobby.Manager, hub *netplay.Hub, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		if session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		lobbyID := chi.URLParam(r, "lobbyID")

		if err := lobbyManager.LeaveLobby(lobbyID, session.UserID); err != nil {
			logger.Error("Failed to leave lobby", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Cleanup lobby from hub if empty
		count, _ := lobbyManager.GetPlayerCount(lobbyID)
		if count == 0 {
			hub.CleanupLobby(lobbyID)
		}

		logger.Info("Player left lobby", "lobby_id", lobbyID, "user", session.Username)
		w.WriteHeader(http.StatusOK)
	}
}

// handleSetReady sets a player's ready status
func handleSetReady(lobbyManager *lobby.Manager, hub *netplay.Hub, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		if session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		lobbyID := chi.URLParam(r, "lobbyID")

		var req struct {
			Ready bool `json:"ready"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := lobbyManager.SetPlayerReady(lobbyID, session.UserID, req.Ready); err != nil {
			logger.Error("Failed to set ready", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info("Player ready status updated", "lobby_id", lobbyID, "user", session.Username, "ready", req.Ready)
		w.WriteHeader(http.StatusOK)
	}
}

// handleStartLobby starts a game in a lobby (host only)
func handleStartLobby(lobbyManager *lobby.Manager, hub *netplay.Hub, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		if session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		lobbyID := chi.URLParam(r, "lobbyID")

		// Verify host
		isHost, err := lobbyManager.IsHost(lobbyID, session.UserID)
		if err != nil || !isHost {
			http.Error(w, "Only the host can start the game", http.StatusForbidden)
			return
		}

		if err := lobbyManager.StartLobby(lobbyID); err != nil {
			logger.Error("Failed to start lobby", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info("Lobby started", "lobby_id", lobbyID)
		w.WriteHeader(http.StatusOK)
	}
}

// handleGetLeaderboard returns the top players
func handleGetLeaderboard(database *db.Database, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50 // Default limit
		entries, err := database.GetLeaderboard(limit)
		if err != nil {
			logger.Error("Failed to get leaderboard", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(entries)
	}
}

// handleGetMyStats returns the current user's stats
func handleGetMyStats(database *db.Database, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := getSession(r)
		if session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		stats, err := database.GetUserStats(session.UserID)
		if err != nil {
			logger.Error("Failed to get user stats", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(stats)
	}
}

// handleWebSocket upgrades HTTP connection to WebSocket and handles netplay communication
func handleWebSocket(hub *netplay.Hub, authService *auth.Service, lobbyManager *lobby.Manager, logger *config.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from query param (for WebSocket connections)
		token := r.URL.Query().Get("token")
		if token == "" {
			token = r.Header.Get("Authorization")
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
		}

		session, err := authService.ValidateSession(token)
		if err != nil {
			logger.Error("WebSocket auth failed", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("WebSocket upgrade failed", "error", err)
			return
		}

		client := &netplay.ClientConnection{
			Conn:      conn,
			UserID:    session.UserID,
			Username:  session.Username,
			PlayerNum: 0, // Will be set when joining lobby
			LobbyID:   "",
			lastPing:  time.Now(),
			inputBuf:  make(chan netplay.Input, 64),
		}

		// Register client with hub
		hub.Register <- client

		logger.Info("WebSocket connected", "user", session.Username)

		// Handle messages
		go handleWebSocketMessages(client, hub, lobbyManager, logger)

		// Handle disconnection
		defer func() {
			hub.Unregister <- client
			conn.Close()
			logger.Info("WebSocket disconnected", "user", session.Username)
		}()
	}
}

// handleWebSocketMessages reads and processes WebSocket messages
func handleWebSocketMessages(client *netplay.ClientConnection, hub *netplay.Hub, lobbyManager *lobby.Manager, logger *config.Logger) {
	for {
		var msg netplay.Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		// Process message based on type
		switch msg.Type {
		case netplay.MsgTypeInput:
			var payload netplay.InputPayload
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				hub.AddInputForFrame(client.LobbyID, client.UserID, payload.Input)
			}
		case netplay.MsgTypeHeartbeat:
			client.lastPing = time.Now()
		}
	}
}
