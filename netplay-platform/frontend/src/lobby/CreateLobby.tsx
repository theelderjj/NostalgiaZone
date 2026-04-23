import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

export default function CreateLobby() {
  const navigate = useNavigate();
  const [gameTitle, setGameTitle] = useState('');
  const [gameId, setGameId] = useState('n64');
  const [username, setUsername] = useState('');
  const [loading, setLoading] = useState(false);

  const handleCreate = async () => {
    if (!gameTitle.trim()) {
      alert('Please enter a game title');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch('/api/lobbies', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          game_title: gameTitle,
          game_id: gameId,
          username: username || undefined,
        }),
      });

      if (res.ok) {
        const data = await res.json();
        // Store username for session
        if (data.username) {
          localStorage.setItem('player_username', data.username);
        }
        navigate(`/lobby/${data.id}`);
      } else {
        const error = await res.text();
        alert('Failed to create lobby: ' + error);
      }
    } catch (err) {
      console.error('Create error:', err);
      alert('Failed to create lobby');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-darker">
      <div className="max-w-2xl mx-auto px-6 py-12">
        <h1 className="text-3xl font-bold mb-8">Create Lobby</h1>

        <div className="bg-dark rounded-xl p-6 border border-gray-800">
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Username (optional)
            </label>
            <input
              type="text"
              placeholder="Enter your username or leave blank for random"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-4 py-3 bg-darker border border-gray-700 rounded-lg focus:outline-none focus:border-primary transition"
            />
          </div>

          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Game Title
            </label>
            <input
              type="text"
              placeholder="e.g., Super Mario 64"
              value={gameTitle}
              onChange={(e) => setGameTitle(e.target.value)}
              className="w-full px-4 py-3 bg-darker border border-gray-700 rounded-lg focus:outline-none focus:border-primary transition"
            />
          </div>

          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Platform
            </label>
            <select
              value={gameId}
              onChange={(e) => setGameId(e.target.value)}
              className="w-full px-4 py-3 bg-darker border border-gray-700 rounded-lg focus:outline-none focus:border-primary transition"
            >
              <option value="n64">Nintendo 64</option>
              <option value="gc">GameCube</option>
            </select>
          </div>

          <div className="flex gap-4">
            <button
              onClick={handleCreate}
              disabled={loading}
              className="flex-1 py-3 bg-primary hover:bg-indigo-500 disabled:bg-gray-600 rounded-lg font-semibold transition"
            >
              {loading ? 'Creating...' : 'Create Lobby'}
            </button>
            <button
              onClick={() => navigate('/lobbies')}
              className="px-6 py-3 bg-gray-700 hover:bg-gray-600 rounded-lg font-semibold transition"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
