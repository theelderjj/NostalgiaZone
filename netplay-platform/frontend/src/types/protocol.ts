/**
 * WebSocket Protocol Types for Netplay Platform
 * These types must match the backend Go structs exactly
 */

export const PROTOCOL_VERSION = '1.0.0';

export type MessageType =
  // Client -> Host
  | 'join_lobby'
  | 'leave_lobby'
  | 'ready'
  | 'input'
  | 'chat'
  | 'start_game'
  | 'end_game'
  | 'reconnect'
  | 'heartbeat'
  // Host -> Client
  | 'lobby_state'
  | 'player_joined'
  | 'player_left'
  | 'game_started'
  | 'game_ended'
  | 'input_packet'
  | 'error'
  | 'success'
  | 'host_migration';

/**
 * Base WebSocket message structure
 */
export interface WSMessage {
  type: MessageType;
  protocol_version: string;
  payload?: unknown;
  timestamp: number;
}

/**
 * Controller input for a single frame
 * Matches the Go Input struct
 */
export interface Input {
  player_num: number;    // 1-4
  frame: number;         // Frame number
  buttons: number;       // Bitmask of button states
  axis_x: number;        // Left stick X (-127 to 127)
  axis_y: number;        // Left stick Y (-127 to 127)
  c_axis_x: number;      // C-stick X (-127 to 127)
  c_axis_y: number;      // C-stick Y (-127 to 127)
  trigger_l: number;     // L trigger (0-255)
  trigger_r: number;     // R trigger (0-255)
}

/**
 * Aggregated inputs from all players for a frame
 */
export interface InputPacket {
  frame: number;
  player_inputs: Record<number, Input>;
  checksum: number;
}

/**
 * Player state in lobby
 */
export interface PlayerState {
  user_id: number;
  username: string;
  player_num: number;
  ready: boolean;
  is_host: boolean;
  latency_ms?: number;
}

/**
 * Lobby state payload
 */
export interface LobbyStatePayload {
  lobby_id: string;
  game_id: string;
  game_title: string;
  host_id: number;
  status: 'waiting' | 'in-progress' | 'closed';
  players: PlayerState[];
  player_count: number;
  max_players: number;
}

/**
 * User session from auth
 */
export interface Session {
  token: string;
  user_id: number;
  username: string;
  created_at: string;
  expires_at: string;
}

/**
 * User info
 */
export interface User {
  id: number;
  username: string;
}

/**
 * Lobby info
 */
export interface Lobby {
  id: string;
  game_id: string;
  game_title: string;
  host_id: number;
  status: string;
  player_count: number;
  max_players: number;
  created_at?: string;
  started_at?: string;
  ended_at?: string;
}

/**
 * Leaderboard entry
 */
export interface LeaderboardEntry {
  user_id: number;
  username: string;
  wins: number;
  losses: number;
  draws: number;
  play_time: number;
  games_played: number;
  last_game: string;
}

/**
 * Button bitmask constants for GameCube/N64 controller
 */
export const BUTTONS = {
  A: 1 << 0,
  B: 1 << 1,
  X: 1 << 2,
  Y: 1 << 3,
  Z: 1 << 4,
  L: 1 << 5,
  R: 1 << 6,
  START: 1 << 7,
  UP: 1 << 8,
  DOWN: 1 << 9,
  LEFT: 1 << 10,
  RIGHT: 1 << 11,
} as const;

/**
 * Game metadata for browser
 */
export interface GameInfo {
  id: string;
  title: string;
  platform: 'N64' | 'GameCube';
  thumbnail?: string;
  tags: string[];
  description?: string;
}

/**
 * Chat message in lobby
 */
export interface ChatMessage {
  user_id: number;
  username: string;
  message: string;
  timestamp: number;
}

/**
 * Connection quality metrics
 */
export interface ConnectionQuality {
  latency_ms: number;
  packet_loss: number;
  jitter_ms: number;
  rating: 'excellent' | 'good' | 'fair' | 'poor';
}

/**
 * Emulator bridge interface - implemented by WASM core
 * This is the contract between the netplay client and emulator
 */
export interface EmulatorBridge {
  /**
   * Initialize the emulator with game ROM
   */
  loadROM(romData: ArrayBuffer): Promise<void>;
  
  /**
   * Send controller input for a specific frame
   * Called by netplay client when input packet arrives
   */
  sendInput(playerNum: number, frame: number, input: Input): void;
  
  /**
   * Run emulator for one frame with buffered inputs
   * Returns true if frame was rendered, false if waiting for inputs
   */
  runFrame(): boolean;
  
  /**
   * Get current frame number
   */
  getCurrentFrame(): number;
  
  /**
   * Set random seed for deterministic initialization
   */
  setSeed(seed: number): void;
  
  /**
   * Get canvas element for rendering
   */
  getCanvas(): HTMLCanvasElement | null;
  
  /**
   * Cleanup and unload emulator
   */
  destroy(): void;
  
  /**
   * Check if emulator is ready
   */
  isReady(): boolean;
}

/**
 * Factory function type for creating emulator bridges
 */
export type EmulatorBridgeFactory = () => Promise<EmulatorBridge>;
