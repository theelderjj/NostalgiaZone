package netplay

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ProtocolVersion is the current WebSocket protocol version
// Increment major version for breaking changes, minor for backwards-compatible additions
const ProtocolVersion = "1.0.0"

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Client -> Host messages
	MsgTypeJoinLobby      MessageType = "join_lobby"
	MsgTypeLeaveLobby     MessageType = "leave_lobby"
	MsgTypeReady          MessageType = "ready"
	MsgTypeInput          MessageType = "input"
	MsgTypeChat           MessageType = "chat"
	MsgTypeStartGame      MessageType = "start_game"
	MsgTypeEndGame        MessageType = "end_game"
	MsgTypeReconnect      MessageType = "reconnect"
	MsgTypeHeartbeat      MessageType = "heartbeat"

	// Host -> Client messages
	MsgTypeLobbyState     MessageType = "lobby_state"
	MsgTypePlayerJoined   MessageType = "player_joined"
	MsgTypePlayerLeft     MessageType = "player_left"
	MsgTypeGameStarted    MessageType = "game_started"
	MsgTypeGameEnded      MessageType = "game_ended"
	MsgTypeInputPacket    MessageType = "input_packet"
	MsgTypeError          MessageType = "error"
	MsgTypeSuccess        MessageType = "success"
	MsgTypeHostMigration  MessageType = "host_migration" // Placeholder for future feature
)

// Message is the base WebSocket message structure
type Message struct {
	Type            MessageType     `json:"type"`
	ProtocolVersion string          `json:"protocol_version"`
	Payload         json.RawMessage `json:"payload,omitempty"`
	Timestamp       int64           `json:"timestamp"` // Unix milliseconds
}

// Input represents controller input for a single frame
// This is a compact binary-friendly representation
type Input struct {
	PlayerNum int    `json:"player_num"` // 1-4
	Frame     int64  `json:"frame"`      // Frame number this input is for
	Buttons   uint32 `json:"buttons"`    // Bitmask of button states
	AxisX     int8   `json:"axis_x"`     // Left stick X (-127 to 127)
	AxisY     int8   `json:"axis_y"`     // Left stick Y (-127 to 127)
	CAxisX    int8   `json:"c_axis_x"`   // C-stick X (-127 to 127)
	CAxisY    int8   `json:"c_axis_y"`   // C-stick Y (-127 to 127)
	TriggerL  uint8  `json:"trigger_l"`  // L trigger (0-255)
	TriggerR  uint8  `json:"trigger_r"`  // R trigger (0-255)
}

// InputPacket contains aggregated inputs from all players for a frame
// Sent by host to all clients at tick rate
type InputPacket struct {
	Frame       int64           `json:"frame"`
	PlayerInputs map[int]Input  `json:"player_inputs"` // playerNum -> Input
	Checksum    uint32          `json:"checksum"`      // Simple checksum for validation
}

// LobbyStatePayload contains the current state of a lobby
type LobbyStatePayload struct {
	LobbyID     string           `json:"lobby_id"`
	GameID      string           `json:"game_id"`
	GameTitle   string           `json:"game_title"`
	HostID      int64            `json:"host_id"`
	Status      string           `json:"status"`
	Players     []PlayerState    `json:"players"`
	PlayerCount int              `json:"player_count"`
	MaxPlayers  int              `json:"max_players"`
}

// PlayerState represents a player's state in the lobby
type PlayerState struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	PlayerNum int    `json:"player_num"`
	Ready     bool   `json:"ready"`
	IsHost    bool   `json:"is_host"`
	Latency   int    `json:"latency_ms"` // Client-reported latency to host
}

// JoinLobbyPayload is sent by client to join a lobby
type JoinLobbyPayload struct {
	LobbyID string `json:"lobby_id"`
}

// ReadyPayload is sent by client to toggle ready status
type ReadyPayload struct {
	Ready bool `json:"ready"`
}

// InputPayload is sent by client to report input for a frame
type InputPayload struct {
	Input Input `json:"input"`
}

// ChatPayload is sent by client to send a chat message
type ChatPayload struct {
	Message string `json:"message"`
}

// StartGamePayload is sent by host to start the game
type StartGamePayload struct {
	Seed int64 `json:"seed"` // Random seed for deterministic initialization
}

// EndGamePayload is sent to end a match
type EndGamePayload struct {
	WinnerID *int64 `json:"winner_id,omitempty"` // NULL for draw
	Duration int64  `json:"duration"`            // Match duration in seconds
}

// ErrorPayload contains error information
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SuccessPayload contains success information
type SuccessPayload struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HeartbeatPayload for connection keepalive
type HeartbeatPayload struct {
	Timestamp int64 `json:"timestamp"`
	Latency   int   `json:"latency_ms,omitempty"`
}

// ReconnectPayload for reconnecting after disconnect
type ReconnectPayload struct {
	LobbyID       string `json:"lobby_id"`
	LastFrameRecv int64  `json:"last_frame_recv"` // Last frame successfully received
}

// ClientConnection represents a connected client in the netplay hub
type ClientConnection struct {
	Conn      *websocket.Conn
	UserID    int64
	Username  string
	PlayerNum int
	LobbyID   string
	mu        sync.Mutex
	lastPing  time.Time
	inputBuf  chan Input // Buffered inputs from this client
}

// Hub manages WebSocket connections and message routing
type Hub struct {
	clients    map[*ClientConnection]bool
	broadcast  chan Message
	register   chan *ClientConnection
	unregister chan *ClientConnection
	mu         sync.RWMutex
	
	// Per-lobby state
	lobbyInputs   map[string]map[int64]Input // lobbyID -> userID -> latest input
	lobbyFrames   map[string]int64           // lobbyID -> current frame number
	lobbyTickRate int                        // frames per second
	mu_lobby      sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub(tickRate int) *Hub {
	return &Hub{
		clients:       make(map[*ClientConnection]bool),
		broadcast:     make(chan Message, 256),
		register:      make(chan *ClientConnection),
		unregister:    make(chan *ClientConnection),
		lobbyInputs:   make(map[string]map[int64]Input),
		lobbyFrames:   make(map[string]int64),
		lobbyTickRate: tickRate,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	ticker := time.NewTicker(time.Second / time.Duration(h.lobbyTickRate))
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.inputBuf)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if client.LobbyID == getMessageLobbyID(message) {
					select {
					case client.Conn.WriteJSON <- message:
					default:
						// Client connection is dead, queue for cleanup
						go h.unregister <- client
					}
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			// Tick: aggregate inputs and broadcast to all clients in active lobbies
			h.processTick()
		}
	}
}

// getMessageLobbyID extracts lobby ID from message payload
func getMessageLobbyID(msg Message) string {
	// This is a simplified implementation
	// In production, you'd properly parse the payload
	return ""
}

// processTick aggregates inputs and broadcasts input packets
// This is the core of deterministic lockstep synchronization
func (h *Hub) processTick() {
	h.mu_lobby.Lock()
	defer h.mu_lobby.Unlock()

	// For each active lobby
	for lobbyID, inputs := range h.lobbyInputs {
		currentFrame := h.lobbyFrames[lobbyID]
		
		// Create input packet for current frame
		packet := InputPacket{
			Frame:        currentFrame,
			PlayerInputs: make(map[int]Input),
		}

		// Collect inputs from all players
		// In a real implementation, you'd handle missing inputs with timeout/retry logic
		for userID, input := range inputs {
			if input.Frame == currentFrame {
				packet.PlayerInputs[input.PlayerNum] = input
			}
		}

		// Calculate simple checksum for validation
		packet.Checksum = calculateChecksum(packet)

		// Broadcast to all clients in this lobby
		msg := Message{
			Type:            MsgTypeInputPacket,
			ProtocolVersion: ProtocolVersion,
			Timestamp:       time.Now().UnixMilli(),
		}
		msg.Payload, _ = json.Marshal(packet)

		h.broadcast <- msg

		// Increment frame counter
		h.lobbyFrames[lobbyID]++
		
		// Clear inputs for next frame
		h.lobbyInputs[lobbyID] = make(map[int64]Input)
	}
}

// calculateChecksum computes a simple checksum for an input packet
func calculateChecksum(packet InputPacket) uint32 {
	var sum uint32
	sum += uint32(packet.Frame)
	for _, input := range packet.PlayerInputs {
		sum += uint32(input.Buttons)
		sum += uint32(input.AxisX) + uint32(input.AxisY)
		sum += uint32(input.CAxisX) + uint32(input.CAxisY)
		sum += uint32(input.TriggerL) + uint32(input.TriggerR)
	}
	return sum
}

// SendToLobby sends a message to all clients in a specific lobby
func (h *Hub) SendToLobby(lobbyID string, msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.LobbyID == lobbyID {
			client.mu.Lock()
			select {
			case client.Conn.WriteJSON <- msg:
			default:
				// Connection dead
			}
			client.mu.Unlock()
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetClientsInLobby returns all clients in a specific lobby
func (h *Hub) GetClientsInLobby(lobbyID string) []*ClientConnection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*ClientConnection
	for client := range h.clients {
		if client.LobbyID == lobbyID {
			clients = append(clients, client)
		}
	}
	return clients
}

// AddInputForFrame stores an input for a specific frame
func (h *Hub) AddInputForFrame(lobbyID string, userID int64, input Input) {
	h.mu_lobby.Lock()
	defer h.mu_lobby.Unlock()

	if _, exists := h.lobbyInputs[lobbyID]; !exists {
		h.lobbyInputs[lobbyID] = make(map[int64]Input)
	}
	h.lobbyInputs[lobbyID][userID] = input
}

// GetCurrentFrame returns the current frame number for a lobby
func (h *Hub) GetCurrentFrame(lobbyID string) int64 {
	h.mu_lobby.RLock()
	defer h.mu_lobby.RUnlock()

	if frame, exists := h.lobbyFrames[lobbyID]; exists {
		return frame
	}
	return 0
}

// InitializeLobby initializes frame tracking for a new lobby
func (h *Hub) InitializeLobby(lobbyID string) {
	h.mu_lobby.Lock()
	defer h.mu_lobby.Unlock()

	h.lobbyFrames[lobbyID] = 0
	h.lobbyInputs[lobbyID] = make(map[int64]Input)
}

// CleanupLobby removes lobby state
func (h *Hub) CleanupLobby(lobbyID string) {
	h.mu_lobby.Lock()
	defer h.mu_lobby.Unlock()

	delete(h.lobbyFrames, lobbyID)
	delete(h.lobbyInputs, lobbyID)
}
