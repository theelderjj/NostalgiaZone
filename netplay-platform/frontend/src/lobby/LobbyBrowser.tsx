import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

interface Lobby {
  id: string;
  game_id: string;
  game_title: string;
  host_id: number;
  status: string;
  player_count: number;
  max_players: number;
  created_at: string;
}

export default function LobbyBrowser() {
  const [lobbies, setLobbies] = useState<Lobby[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('');

  useEffect(() => {
    fetchLobbies();
  }, []);

  const fetchLobbies = async () => {
    try {
      const res = await fetch('/api/lobbies');
      if (res.ok) {
        const data = await res.json();
        setLobbies(data);
      }
    } catch (err) {
      console.error('Failed to fetch lobbies:', err);
    } finally {
      setLoading(false);
    }
  };

  const filteredLobbies = lobbies.filter(lobby =>
    lobby.game_title.toLowerCase().includes(filter.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-darker">
      <div className="max-w-7xl mx-auto px-6 py-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-3xl font-bold">Browse Lobbies</h1>
          
          <Link
            to="/lobby/create"
            className="px-6 py-3 bg-primary hover:bg-indigo-500 rounded-lg font-semibold transition"
          >
            Create Lobby
          </Link>
        </div>

        {/* Search/Filter */}
        <div className="mb-6">
          <input
            type="text"
            placeholder="Search games..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="w-full max-w-md px-4 py-3 bg-dark border border-gray-700 rounded-lg focus:outline-none focus:border-primary transition"
          />
        </div>

        {/* Lobby List */}
        {loading ? (
          <div className="text-center py-12 text-gray-400">Loading lobbies...</div>
        ) : filteredLobbies.length === 0 ? (
          <div className="text-center py-12 text-gray-400">
            No lobbies found. Create one!
          </div>
        ) : (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredLobbies.map((lobby) => (
              <LobbyCard key={lobby.id} lobby={lobby} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function LobbyCard({ lobby }: { lobby: Lobby }) {
  const handleJoin = async () => {
    try {
      const res = await fetch(`/api/lobbies/${lobby.id}/join`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (res.ok) {
        window.location.href = `/lobby/${lobby.id}`;
      } else {
        alert('Failed to join lobby');
      }
    } catch (err) {
      console.error('Join error:', err);
      alert('Failed to join lobby');
    }
  };

  const isFull = lobby.player_count >= lobby.max_players;
  const isOpen = lobby.status === 'waiting';

  return (
    <div className="bg-dark rounded-xl p-6 border border-gray-800 hover:border-primary/50 transition">
      <div className="aspect-video bg-gray-800 rounded-lg mb-4 flex items-center justify-center">
        <span className="text-4xl">🎮</span>
      </div>
      
      <h3 className="text-xl font-bold mb-2">{lobby.game_title}</h3>
      
      <div className="flex items-center justify-between mb-4">
        <span className={`px-3 py-1 rounded-full text-sm font-medium ${
          lobby.status === 'waiting' ? 'bg-green-900/50 text-green-400' :
          lobby.status === 'in-progress' ? 'bg-yellow-900/50 text-yellow-400' :
          'bg-gray-700 text-gray-400'
        }`}>
          {lobby.status}
        </span>
        
        <span className="text-gray-400 text-sm">
          {lobby.player_count}/{lobby.max_players} players
        </span>
      </div>

      {isOpen && !isFull ? (
        <button
          onClick={handleJoin}
          className="w-full py-2 bg-primary hover:bg-indigo-500 rounded-lg font-semibold transition"
        >
          Join Lobby
        </button>
      ) : (
        <Link
          to={`/lobby/${lobby.id}`}
          className="block w-full py-2 bg-gray-700 hover:bg-gray-600 rounded-lg font-semibold text-center transition"
        >
          View Details
        </Link>
      )}
    </div>
  );
}
