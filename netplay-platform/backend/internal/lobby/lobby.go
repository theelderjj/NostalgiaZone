package lobby

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"netplay-platform/backend/internal/db"
)

var (
	ErrLobbyFull       = errors.New("lobby is full")
	ErrLobbyNotFound   = errors.New("lobby not found")
	ErrLobbyNotWaiting = errors.New("lobby is not in waiting status")
	ErrPlayerNotInLobby = errors.New("player not in lobby")
	ErrInvalidPlayerNum = errors.New("invalid player number")
)

// LobbyState represents the current state of a lobby
type LobbyState string

const (
	StateWaiting    LobbyState = "waiting"
	StateInProgress LobbyState = "in-progress"
	StateClosed     LobbyState = "closed"
)

// PlayerInfo contains information about a player in a lobby
type PlayerInfo struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	PlayerNum int    `json:"player_num"`
	Ready     bool   `json:"ready"`
}

// LobbyManager handles lobby lifecycle and state
type Manager struct {
	db           *db.Database
	lobbies      map[string]*LobbyInfo
	mu           sync.RWMutex
	hostSessions map[string]string // lobbyID -> sessionToken (for host validation)
}

// LobbyInfo extends db.Lobby with runtime information
type LobbyInfo struct {
	Lobby       *db.Lobby
	Players     map[int64]*PlayerInfo // userID -> PlayerInfo
	NextPlayerNum int
	mu          sync.RWMutex
}

// NewManager creates a new lobby manager
func NewManager(database *db.Database) *Manager {
	return &Manager{
		db:           database,
		lobbies:      make(map[string]*LobbyInfo),
		hostSessions: make(map[string]string),
	}
}

// CreateLobby creates a new lobby
func (m *Manager) CreateLobby(id, gameID, gameTitle string, hostID int64, hostUsername string) (*LobbyInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if lobby already exists
	if _, exists := m.lobbies[id]; exists {
		return nil, errors.New("lobby already exists")
	}

	// Create in database
	lobby, err := m.db.CreateLobby(id, gameID, gameTitle, hostID)
	if err != nil {
		return nil, err
	}

	// Create in-memory info
	info := &LobbyInfo{
		Lobby:         lobby,
		Players:       make(map[int64]*PlayerInfo),
		NextPlayerNum: 1,
	}

	// Add host as first player
	info.Players[hostID] = &PlayerInfo{
		UserID:    hostID,
		Username:  hostUsername,
		PlayerNum: 1,
		Ready:     true, // Host is automatically ready
	}
	info.NextPlayerNum = 2

	m.lobbies[id] = info
	return info, nil
}

// GetLobby retrieves a lobby by ID
func (m *Manager) GetLobby(id string) (*LobbyInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.lobbies[id]
	if !exists {
		return nil, ErrLobbyNotFound
	}

	return info, nil
}

// GetActiveLobbies returns all active lobbies
func (m *Manager) GetActiveLobbies() ([]*LobbyInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*LobbyInfo
	for _, info := range m.lobbies {
		if info.Lobby.Status != string(StateClosed) {
			result = append(result, info)
		}
	}
	return result, nil
}

// JoinLobby adds a player to a lobby
func (m *Manager) JoinLobby(lobbyID string, userID int64, username string) (*PlayerInfo, error) {
	m.mu.Lock()
	info, exists := m.lobbies[lobbyID]
	if !exists {
		m.mu.Unlock()
		return nil, ErrLobbyNotFound
	}
	m.mu.Unlock()

	info.mu.Lock()
	defer info.mu.Unlock()

	// Check if already in lobby
	if _, exists := info.Players[userID]; exists {
		return info.Players[userID], nil
	}

	// Check if full
	if len(info.Players) >= info.Lobby.MaxPlayers {
		return nil, ErrLobbyFull
	}

	// Check status
	if info.Lobby.Status != string(StateWaiting) {
		return nil, ErrLobbyNotWaiting
	}

	// Assign player number
	playerNum := info.NextPlayerNum
	info.NextPlayerNum++

	// Add player
	playerInfo := &PlayerInfo{
		UserID:    userID,
		Username:  username,
		PlayerNum: playerNum,
		Ready:     false,
	}
	info.Players[userID] = playerInfo

	// Update in database
	err := m.db.AddPlayerToLobby(lobbyID, userID, username, playerNum)
	if err != nil {
		delete(info.Players, userID)
		return nil, err
	}

	// Update lobby player count
	info.Lobby.PlayerCount = len(info.Players)

	return playerInfo, nil
}

// LeaveLobby removes a player from a lobby
func (m *Manager) LeaveLobby(lobbyID string, userID int64) error {
	m.mu.Lock()
	info, exists := m.lobbies[lobbyID]
	if !exists {
		m.mu.Unlock()
		return ErrLobbyNotFound
	}
	
	// Check if this is the host leaving
	isHost := info.Lobby.HostID == userID
	m.mu.Unlock()

	info.mu.Lock()
	defer info.mu.Unlock()

	// Check if player is in lobby
	if _, exists := info.Players[userID]; !exists {
		return ErrPlayerNotInLobby
	}

	// Remove player
	delete(info.Players, userID)
	info.Lobby.PlayerCount = len(info.Players)

	// If host left, close the lobby
	if isHost {
		info.Lobby.Status = string(StateClosed)
		_ = m.db.EndLobby(lobbyID)
		
		// Clean up from memory after a delay
		go func() {
			time.Sleep(5 * time.Minute)
			m.mu.Lock()
			delete(m.lobbies, lobbyID)
			m.mu.Unlock()
		}()
		
		return nil
	}

	// Remove from database
	err := m.db.RemovePlayerFromLobby(lobbyID, userID)
	if err != nil {
		return err
	}

	// If lobby is now empty and not started, close it
	if len(info.Players) == 0 && info.Lobby.Status == string(StateWaiting) {
		info.Lobby.Status = string(StateClosed)
		_ = m.db.EndLobby(lobbyID)
	}

	return nil
}

// SetPlayerReady sets a player's ready status
func (m *Manager) SetPlayerReady(lobbyID string, userID int64, ready bool) error {
	info, err := m.GetLobby(lobbyID)
	if err != nil {
		return err
	}

	info.mu.Lock()
	defer info.mu.Unlock()

	player, exists := info.Players[userID]
	if !exists {
		return ErrPlayerNotInLobby
	}

	player.Ready = ready
	return m.db.SetPlayerReady(lobbyID, userID, ready)
}

// AreAllPlayersReady checks if all players in a lobby are ready
func (m *Manager) AreAllPlayersReady(lobbyID string) (bool, error) {
	info, err := m.GetLobby(lobbyID)
	if err != nil {
		return false, err
	}

	info.mu.RLock()
	defer info.mu.RUnlock()

	for _, player := range info.Players {
		if !player.Ready {
			return false, nil
		}
	}
	return true, nil
}

// StartLobby transitions a lobby to in-progress state
func (m *Manager) StartLobby(lobbyID string) error {
	info, err := m.GetLobby(lobbyID)
	if err != nil {
		return err
	}

	info.mu.Lock()
	defer info.mu.Unlock()

	// Verify all players are ready
	for _, player := range info.Players {
		if !player.Ready {
			return fmt.Errorf("player %s is not ready", player.Username)
		}
	}

	// Update status
	info.Lobby.Status = string(StateInProgress)
	now := time.Now()
	info.Lobby.StartedAt = &now

	return m.db.UpdateLobbyStatus(lobbyID, string(StateInProgress))
}

// EndLobby ends a match and records results
func (m *Manager) EndLobby(lobbyID string, winnerID *int64, duration int64) error {
	info, err := m.GetLobby(lobbyID)
	if err != nil {
		return err
	}

	info.mu.Lock()
	defer info.mu.Unlock()

	// Record match result
	err = m.db.RecordMatchResult(lobbyID, info.Lobby.GameID, winnerID, duration)
	if err != nil {
		return err
	}

	// Update lobby status
	info.Lobby.Status = string(StateClosed)
	now := time.Now()
	info.Lobby.EndedAt = &now

	err = m.db.EndLobby(lobbyID)
	if err != nil {
		return err
	}

	// Clean up from memory after a delay
	go func() {
		time.Sleep(10 * time.Minute)
		m.mu.Lock()
		delete(m.lobbies, lobbyID)
		m.mu.Unlock()
	}()

	return nil
}

// GetPlayers returns all players in a lobby
func (m *Manager) GetPlayers(lobbyID string) (map[int64]*PlayerInfo, error) {
	info, err := m.GetLobby(lobbyID)
	if err != nil {
		return nil, err
	}

	info.mu.RLock()
	defer info.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[int64]*PlayerInfo)
	for k, v := range info.Players {
		result[k] = v
	}
	return result, nil
}

// IsHost checks if a user is the host of a lobby
func (m *Manager) IsHost(lobbyID string, userID int64) (bool, error) {
	info, err := m.GetLobby(lobbyID)
	if err != nil {
		return false, err
	}

	return info.Lobby.HostID == userID, nil
}

// GetPlayerCount returns the number of players in a lobby
func (m *Manager) GetPlayerCount(lobbyID string) (int, error) {
	info, err := m.GetLobby(lobbyID)
	if err != nil {
		return 0, err
	}

	info.mu.RLock()
	defer info.mu.RUnlock()

	return len(info.Players), nil
}
