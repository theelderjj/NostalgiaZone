import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

export default function LobbyRoom() {
  const { lobbyId } = useParams<{ lobbyId: string }>();
  const navigate = useNavigate();
  const [lobby, setLobby] = useState<any>(null);
  const [ready, setReady] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchLobby();
  }, [lobbyId]);

  const fetchLobby = async () => {
    try {
      const res = await fetch(`/api/lobbies/${lobbyId}`);
      if (res.ok) {
        const data = await res.json();
        setLobby(data);
        
        // Check if current user is in lobby and ready
        const player = data.players?.find((p: any) => p.username === localStorage.getItem('player_username'));
        if (player) setReady(player.ready);
      }
    } catch (err) {
      console.error('Failed to fetch lobby:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleReady = async () => {
    try {
      await fetch(`/api/lobbies/${lobbyId}/ready`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ ready: !ready }),
      });
      setReady(!ready);
      fetchLobby();
    } catch (err) {
      console.error('Failed to toggle ready:', err);
    }
  };

  const handleStart = async () => {
    try {
      await fetch(`/api/lobbies/${lobbyId}/start`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      navigate(`/game/${lobbyId}`);
    } catch (err) {
      console.error('Failed to start:', err);
      alert('Make sure all players are ready!');
    }
  };

  const handleLeave = async () => {
    try {
      await fetch(`/api/lobbies/${lobbyId}/leave`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      navigate('/lobbies');
    } catch (err) {
      console.error('Failed to leave:', err);
    }
  };

  if (loading) return <div className="min-h-screen bg-darker flex items-center justify-center">Loading...</div>;
  if (!lobby) return <div className="min-h-screen bg-darker flex items-center justify-center">Lobby not found</div>;

  const isHost = lobby.players?.some((p: any) => p.username === localStorage.getItem('player_username') && p.is_host);
  const allReady = lobby.players?.every((p: any) => p.ready);

  return (
    <div className="min-h-screen bg-darker">
      <div className="max-w-4xl mx-auto px-6 py-8">
        <h1 className="text-3xl font-bold mb-2">{lobby.game_title}</h1>
        <p className="text-gray-400 mb-6">Lobby ID: {lobby.id}</p>

        {/* Players Grid */}
        <div className="grid grid-cols-2 gap-4 mb-8">
          {[1, 2, 3, 4].map((num) => {
            const player = lobby.players?.find((p: any) => p.player_num === num);
            return (
              <div
                key={num}
                className={`p-4 rounded-lg border ${
                  player ? 'bg-dark border-gray-700' : 'bg-gray-900 border-gray-800 border-dashed'
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="font-semibold">Player {num}</span>
                  {player && (
                    <span className={`px-2 py-1 rounded text-xs ${player.ready ? 'bg-green-900 text-green-400' : 'bg-yellow-900 text-yellow-400'}`}>
                      {player.ready ? 'READY' : 'NOT READY'}
                    </span>
                  )}
                </div>
                {player ? (
                  <p className="text-gray-400 mt-1">{player.username} {player.is_host && '(Host)'}</p>
                ) : (
                  <p className="text-gray-600 mt-1">Empty</p>
                )}
              </div>
            );
          })}
        </div>

        {/* Actions */}
        <div className="flex gap-4">
          {!isHost && session && (
            <button
              onClick={handleToggleReady}
              className={`px-6 py-3 rounded-lg font-semibold transition ${
                ready ? 'bg-red-600 hover:bg-red-700' : 'bg-green-600 hover:bg-green-700'
              }`}
            >
              {ready ? 'Not Ready' : 'Ready Up'}
            </button>
          )}
          
          {isHost && allReady && (
            <button
              onClick={handleStart}
              className="px-6 py-3 bg-primary hover:bg-indigo-500 rounded-lg font-semibold transition"
            >
              Start Game
            </button>
          )}
          
          <button
            onClick={handleLeave}
            className="px-6 py-3 bg-gray-700 hover:bg-gray-600 rounded-lg font-semibold transition"
          >
            Leave Lobby
          </button>
        </div>
      </div>
    </div>
  );
}
