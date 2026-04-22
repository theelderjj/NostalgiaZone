package lobby

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_CreateLobby(t *testing.T) {
	database := setupTestDB(t)
	manager := NewManager(database)

	t.Run("create new lobby", func(t *testing.T) {
		info, err := manager.CreateLobby("test-lobby", "game123", "Test Game", 1, "hostuser")
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test-lobby", info.Lobby.ID)
		assert.Equal(t, "game123", info.Lobby.GameID)
		assert.Equal(t, "Test Game", info.Lobby.GameTitle)
		assert.Equal(t, int64(1), info.Lobby.HostID)
		assert.Equal(t, "waiting", info.Lobby.Status)
		assert.Equal(t, 1, info.Lobby.PlayerCount)
	})

	t.Run("duplicate lobby ID", func(t *testing.T) {
		_, err := manager.CreateLobby("dup-lobby", "game123", "Test Game", 1, "hostuser")
		require.NoError(t, err)

		_, err = manager.CreateLobby("dup-lobby", "game456", "Another Game", 2, "otheruser")
		assert.Error(t, err)
	})
}

func TestManager_JoinLobby(t *testing.T) {
	database := setupTestDB(t)
	manager := NewManager(database)

	// Create lobby
	_, err := manager.CreateLobby("join-test", "game123", "Test Game", 1, "hostuser")
	require.NoError(t, err)

	t.Run("join existing lobby", func(t *testing.T) {
		player, err := manager.JoinLobby("join-test", 2, "player2")
		require.NoError(t, err)
		assert.NotNil(t, player)
		assert.Equal(t, int64(2), player.UserID)
		assert.Equal(t, 2, player.PlayerNum)
		assert.False(t, player.Ready)
	})

	t.Run("lobby full", func(t *testing.T) {
		// Fill up the lobby
		for i := 2; i <= 4; i++ {
			if i == 2 {
				continue // Already joined
			}
			_, err := manager.JoinLobby("join-test", int64(i*10), "player"+string(rune('0'+i)))
			require.NoError(t, err)
		}

		// Try to join when full
		player, err := manager.JoinLobby("join-test", 99, "lateplayer")
		assert.Error(t, err)
		assert.Equal(t, ErrLobbyFull, err)
		assert.Nil(t, player)
	})

	t.Run("join nonexistent lobby", func(t *testing.T) {
		player, err := manager.JoinLobby("nonexistent", 2, "player2")
		assert.Error(t, err)
		assert.Equal(t, ErrLobbyNotFound, err)
		assert.Nil(t, player)
	})
}

func TestManager_LeaveLobby(t *testing.T) {
	database := setupTestDB(t)
	manager := NewManager(database)

	// Create lobby with host and one player
	_, err := manager.CreateLobby("leave-test", "game123", "Test Game", 1, "hostuser")
	require.NoError(t, err)
	_, err = manager.JoinLobby("leave-test", 2, "player2")
	require.NoError(t, err)

	t.Run("player leaves", func(t *testing.T) {
		err := manager.LeaveLobby("leave-test", 2)
		require.NoError(t, err)

		players, _ := manager.GetPlayers("leave-test")
		assert.Len(t, players, 1) // Only host remains
	})

	t.Run("host leaves closes lobby", func(t *testing.T) {
		err := manager.LeaveLobby("leave-test", 1)
		require.NoError(t, err)

		info, err := manager.GetLobby("leave-test")
		require.NoError(t, err)
		assert.Equal(t, "closed", info.Lobby.Status)
	})
}

func TestManager_SetPlayerReady(t *testing.T) {
	database := setupTestDB(t)
	manager := NewManager(database)

	_, err := manager.CreateLobby("ready-test", "game123", "Test Game", 1, "hostuser")
	require.NoError(t, err)

	t.Run("toggle ready status", func(t *testing.T) {
		err := manager.SetPlayerReady("ready-test", 1, true)
		require.NoError(t, err)

		players, _ := manager.GetPlayers("ready-test")
		assert.True(t, players[1].Ready)

		err = manager.SetPlayerReady("ready-test", 1, false)
		require.NoError(t, err)

		players, _ = manager.GetPlayers("ready-test")
		assert.False(t, players[1].Ready)
	})
}

func TestManager_StartLobby(t *testing.T) {
	database := setupTestDB(t)
	manager := NewManager(database)

	_, err := manager.CreateLobby("start-test", "game123", "Test Game", 1, "hostuser")
	require.NoError(t, err)
	_, err = manager.JoinLobby("start-test", 2, "player2")
	require.NoError(t, err)

	t.Run("cannot start with unready players", func(t *testing.T) {
		err := manager.StartLobby("start-test")
		assert.Error(t, err) // Player 2 not ready
	})

	t.Run("start with all ready", func(t *testing.T) {
		err = manager.SetPlayerReady("start-test", 2, true)
		require.NoError(t, err)

		err = manager.StartLobby("start-test")
		require.NoError(t, err)

		info, _ := manager.GetLobby("start-test")
		assert.Equal(t, "in-progress", info.Lobby.Status)
	})
}

func TestManager_AreAllPlayersReady(t *testing.T) {
	database := setupTestDB(t)
	manager := NewManager(database)

	_, err := manager.CreateLobby("allready-test", "game123", "Test Game", 1, "hostuser")
	require.NoError(t, err)
	_, err = manager.JoinLobby("allready-test", 2, "player2")
	require.NoError(t, err)

	t.Run("not all ready", func(t *testing.T) {
		ready, err := manager.AreAllPlayersReady("allready-test")
		require.NoError(t, err)
		assert.False(t, ready)
	})

	t.Run("all ready", func(t *testing.T) {
		err = manager.SetPlayerReady("allready-test", 2, true)
		require.NoError(t, err)

		ready, err := manager.AreAllPlayersReady("allready-test")
		require.NoError(t, err)
		assert.True(t, ready)
	})
}
