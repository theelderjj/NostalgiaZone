package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// User represents a registered user
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	CreatedAt    time.Time `json:"created_at"`
	LastActive   time.Time `json:"last_active"`
}

// Lobby represents a game lobby
type Lobby struct {
	ID          string    `json:"id"`
	GameID      string    `json:"game_id"`
	GameTitle   string    `json:"game_title"`
	HostID      int64     `json:"host_id"`
	Status      string    `json:"status"` // waiting, full, in-progress, closed
	PlayerCount int       `json:"player_count"`
	MaxPlayers  int       `json:"max_players"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
}

// LobbyPlayer represents a player in a lobby
type LobbyPlayer struct {
	LobbyID   string    `json:"lobby_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	PlayerNum int       `json:"player_num"` // 1-4
	Ready     bool      `json:"ready"`
	JoinedAt  time.Time `json:"joined_at"`
}

// MatchResult represents the result of a completed match
type MatchResult struct {
	ID        int64     `json:"id"`
	LobbyID   string    `json:"lobby_id"`
	GameID    string    `json:"game_id"`
	WinnerID  *int64    `json:"winner_id,omitempty"` // NULL for draws
	Duration  int64     `json:"duration"`            // seconds
	EndedAt   time.Time `json:"ended_at"`
}

// LeaderboardEntry represents a user's stats
type LeaderboardEntry struct {
	UserID     int64  `json:"user_id"`
	Username   string `json:"username"`
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
	Draws      int    `json:"draws"`
	PlayTime   int64  `json:"play_time"` // total seconds
	GamesPlayed int   `json:"games_played"`
	LastGame   string `json:"last_game"`
}

// Database wraps the sql.DB connection
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection and initializes schema
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}
	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// initSchema creates the database tables if they don't exist
func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_active DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS lobbies (
		id TEXT PRIMARY KEY,
		game_id TEXT NOT NULL,
		game_title TEXT NOT NULL,
		host_id INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'waiting',
		player_count INTEGER NOT NULL DEFAULT 0,
		max_players INTEGER NOT NULL DEFAULT 4,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		started_at DATETIME,
		ended_at DATETIME,
		FOREIGN KEY (host_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS lobby_players (
		lobby_id TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		username TEXT NOT NULL,
		player_num INTEGER NOT NULL,
		ready BOOLEAN NOT NULL DEFAULT FALSE,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (lobby_id, user_id),
		FOREIGN KEY (lobby_id) REFERENCES lobbies(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS match_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		lobby_id TEXT NOT NULL,
		game_id TEXT NOT NULL,
		winner_id INTEGER,
		duration INTEGER NOT NULL,
		ended_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (lobby_id) REFERENCES lobbies(id)
	);

	CREATE TABLE IF NOT EXISTS user_stats (
		user_id INTEGER PRIMARY KEY,
		wins INTEGER NOT NULL DEFAULT 0,
		losses INTEGER NOT NULL DEFAULT 0,
		draws INTEGER NOT NULL DEFAULT 0,
		play_time INTEGER NOT NULL DEFAULT 0,
		games_played INTEGER NOT NULL DEFAULT 0,
		last_game TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_lobbies_status ON lobbies(status);
	CREATE INDEX IF NOT EXISTS idx_lobby_players_user ON lobby_players(user_id);
	CREATE INDEX IF NOT EXISTS idx_match_results_winner ON match_results(winner_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

// CreateUser creates a new user with hashed password
func (d *Database) CreateUser(username, passwordHash string) (*User, error) {
	result, err := d.db.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		username, passwordHash,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return d.GetUserByID(id)
}

// GetUserByUsername retrieves a user by username
func (d *Database) GetUserByUsername(username string) (*User, error) {
	row := d.db.QueryRow(
		"SELECT id, username, password_hash, created_at, last_active FROM users WHERE username = ?",
		username,
	)
	return scanUser(row)
}

// GetUserByID retrieves a user by ID
func (d *Database) GetUserByID(id int64) (*User, error) {
	row := d.db.QueryRow(
		"SELECT id, username, password_hash, created_at, last_active FROM users WHERE id = ?",
		id,
	)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*User, error) {
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.LastActive)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateLastActive updates the user's last active timestamp
func (d *Database) UpdateLastActive(userID int64) error {
	_, err := d.db.Exec(
		"UPDATE users SET last_active = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	)
	return err
}

// CreateLobby creates a new lobby
func (d *Database) CreateLobby(id, gameID, gameTitle string, hostID int64) (*Lobby, error) {
	_, err := d.db.Exec(
		`INSERT INTO lobbies (id, game_id, game_title, host_id, status, max_players) 
		 VALUES (?, ?, ?, ?, 'waiting', 4)`,
		id, gameID, gameTitle, hostID,
	)
	if err != nil {
		return nil, err
	}

	return d.GetLobby(id)
}

// GetLobby retrieves a lobby by ID
func (d *Database) GetLobby(id string) (*Lobby, error) {
	row := d.db.QueryRow(
		`SELECT id, game_id, game_title, host_id, status, player_count, max_players, 
		        created_at, started_at, ended_at 
		 FROM lobbies WHERE id = ?`,
		id,
	)
	return scanLobby(row)
}

func scanLobby(row *sql.Row) (*Lobby, error) {
	l := &Lobby{}
	var startedAt, endedAt sql.NullTime
	err := row.Scan(
		&l.ID, &l.GameID, &l.GameTitle, &l.HostID, &l.Status, &l.PlayerCount,
		&l.MaxPlayers, &l.CreatedAt, &startedAt, &endedAt,
	)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		l.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		l.EndedAt = &endedAt.Time
	}
	return l, nil
}

// GetActiveLobbies returns all lobbies that are not closed
func (d *Database) GetActiveLobbies() ([]*Lobby, error) {
	rows, err := d.db.Query(
		`SELECT id, game_id, game_title, host_id, status, player_count, max_players, 
		        created_at, started_at, ended_at 
		 FROM lobbies WHERE status != 'closed' ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lobbies []*Lobby
	for rows.Next() {
		lobby, err := scanLobbyRow(rows)
		if err != nil {
			return nil, err
		}
		lobbies = append(lobbies, lobby)
	}
	return lobbies, rows.Err()
}

func scanLobbyRow(rows *sql.Rows) (*Lobby, error) {
	l := &Lobby{}
	var startedAt, endedAt sql.NullTime
	err := rows.Scan(
		&l.ID, &l.GameID, &l.GameTitle, &l.HostID, &l.Status, &l.PlayerCount,
		&l.MaxPlayers, &l.CreatedAt, &startedAt, &endedAt,
	)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		l.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		l.EndedAt = &endedAt.Time
	}
	return l, nil
}

// AddPlayerToLobby adds a player to a lobby
func (d *Database) AddPlayerToLobby(lobbyID string, userID int64, username string, playerNum int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO lobby_players (lobby_id, user_id, username, player_num) 
		 VALUES (?, ?, ?, ?)`,
		lobbyID, userID, username, playerNum,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"UPDATE lobbies SET player_count = player_count + 1 WHERE id = ?",
		lobbyID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RemovePlayerFromLobby removes a player from a lobby
func (d *Database) RemovePlayerFromLobby(lobbyID string, userID int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"DELETE FROM lobby_players WHERE lobby_id = ? AND user_id = ?",
		lobbyID, userID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"UPDATE lobbies SET player_count = player_count - 1 WHERE id = ?",
		lobbyID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetLobbyPlayers returns all players in a lobby
func (d *Database) GetLobbyPlayers(lobbyID string) ([]*LobbyPlayer, error) {
	rows, err := d.db.Query(
		`SELECT lobby_id, user_id, username, player_num, ready, joined_at 
		 FROM lobby_players WHERE lobby_id = ? ORDER BY player_num`,
		lobbyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []*LobbyPlayer
	for rows.Next() {
		p := &LobbyPlayer{}
		err := rows.Scan(&p.LobbyID, &p.UserID, &p.Username, &p.PlayerNum, &p.Ready, &p.JoinedAt)
		if err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// SetPlayerReady sets a player's ready status
func (d *Database) SetPlayerReady(lobbyID string, userID int64, ready bool) error {
	_, err := d.db.Exec(
		"UPDATE lobby_players SET ready = ? WHERE lobby_id = ? AND user_id = ?",
		ready, lobbyID, userID,
	)
	return err
}

// UpdateLobbyStatus updates the lobby status
func (d *Database) UpdateLobbyStatus(lobbyID, status string) error {
	var query string
	if status == "in-progress" {
		query = "UPDATE lobbies SET status = ?, started_at = CURRENT_TIMESTAMP WHERE id = ?"
	} else {
		query = "UPDATE lobbies SET status = ? WHERE id = ?"
	}
	_, err := d.db.Exec(query, status, lobbyID)
	return err
}

// EndLobby marks a lobby as closed and records the end time
func (d *Database) EndLobby(lobbyID string) error {
	_, err := d.db.Exec(
		"UPDATE lobbies SET status = 'closed', ended_at = CURRENT_TIMESTAMP WHERE id = ?",
		lobbyID,
	)
	return err
}

// RecordMatchResult records the result of a match
func (d *Database) RecordMatchResult(lobbyID, gameID string, winnerID *int64, duration int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO match_results (lobby_id, game_id, winner_id, duration) VALUES (?, ?, ?, ?)",
		lobbyID, gameID, winnerID, duration,
	)
	if err != nil {
		return err
	}

	// Update user stats for each player in the lobby
	rows, err := tx.Query("SELECT user_id FROM lobby_players WHERE lobby_id = ?", lobbyID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return err
		}
		userIDs = append(userIDs, userID)
	}

	for _, userID := range userIDs {
		if err := d.updateUserStats(tx, userID, winnerID, duration, gameID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *Database) updateUserStats(tx *sql.Tx, userID int64, winnerID *int64, duration int64, gameID string) error {
	// Initialize or update stats
	_, err := tx.Exec(`
		INSERT INTO user_stats (user_id, wins, losses, draws, play_time, games_played, last_game)
		VALUES (?, 0, 0, 0, 0, 0, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			wins = user_stats.wins + CASE WHEN ? IS NOT NULL AND user_stats.user_id = ? THEN 1 ELSE 0 END,
			losses = user_stats.losses + CASE WHEN ? IS NOT NULL AND user_stats.user_id != ? THEN 1 ELSE 0 END,
			draws = user_stats.draws + CASE WHEN ? IS NULL THEN 1 ELSE 0 END,
			play_time = user_stats.play_time + ?,
			games_played = user_stats.games_played + 1,
			last_game = ?
	`, userID, gameID, winnerID, userID, winnerID, userID, winnerID, duration, gameID)
	return err
}

// GetLeaderboard returns the top users by wins
func (d *Database) GetLeaderboard(limit int) ([]*LeaderboardEntry, error) {
	rows, err := d.db.Query(`
		SELECT u.id, u.username, 
		       COALESCE(s.wins, 0), COALESCE(s.losses, 0), COALESCE(s.draws, 0),
		       COALESCE(s.play_time, 0), COALESCE(s.games_played, 0), COALESCE(s.last_game, '')
		FROM users u
		LEFT JOIN user_stats s ON u.id = s.user_id
		ORDER BY s.wins DESC, s.games_played DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*LeaderboardEntry
	for rows.Next() {
		e := &LeaderboardEntry{}
		err := rows.Scan(&e.UserID, &e.Username, &e.Wins, &e.Losses, &e.Draws, &e.PlayTime, &e.GamesPlayed, &e.LastGame)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetUserStats returns stats for a specific user
func (d *Database) GetUserStats(userID int64) (*LeaderboardEntry, error) {
	row := d.db.QueryRow(`
		SELECT u.id, u.username,
		       COALESCE(s.wins, 0), COALESCE(s.losses, 0), COALESCE(s.draws, 0),
		       COALESCE(s.play_time, 0), COALESCE(s.games_played, 0), COALESCE(s.last_game, '')
		FROM users u
		LEFT JOIN user_stats s ON u.id = s.user_id
		WHERE u.id = ?
	`, userID)

	e := &LeaderboardEntry{}
	err := row.Scan(&e.UserID, &e.Username, &e.Wins, &e.Losses, &e.Draws, &e.PlayTime, &e.GamesPlayed, &e.LastGame)
	if err != nil {
		return nil, err
	}
	return e, nil
}
