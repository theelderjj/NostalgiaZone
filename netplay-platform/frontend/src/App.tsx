import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';

interface Session {
  token: string;
  user_id: number;
  username: string;
}

export default function App() {
  const [session, setSession] = useState<Session | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    // Check for existing session on mount
    const stored = localStorage.getItem('netplay_session');
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        // Verify session is still valid (optional API call)
        setSession(parsed);
      } catch {
        localStorage.removeItem('netplay_session');
      }
    }
  }, []);

  const handleLogout = async () => {
    try {
      await fetch('/api/auth/logout', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${session?.token}`,
        },
      });
    } catch (err) {
      console.error('Logout error:', err);
    }
    
    localStorage.removeItem('netplay_session');
    setSession(null);
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-darker">
      {/* Header */}
      <header className="bg-dark border-b border-gray-800 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <Link to="/" className="text-2xl font-bold text-primary">
            🎮 Netplay Platform
          </Link>
          
          <nav className="flex items-center gap-6">
            <Link to="/lobbies" className="text-gray-300 hover:text-white transition">
              Browse Lobbies
            </Link>
            <Link to="/leaderboard" className="text-gray-300 hover:text-white transition">
              Leaderboard
            </Link>
            
            {session ? (
              <>
                <span className="text-gray-400">
                  Welcome, <span className="text-white font-medium">{session.username}</span>
                </span>
                <button
                  onClick={handleLogout}
                  className="px-4 py-2 bg-red-600 hover:bg-red-700 rounded-lg transition"
                >
                  Logout
                </button>
              </>
            ) : (
              <>
                <Link
                  to="/login"
                  className="px-4 py-2 text-gray-300 hover:text-white transition"
                >
                  Login
                </Link>
                <Link
                  to="/register"
                  className="px-4 py-2 bg-primary hover:bg-indigo-500 rounded-lg transition"
                >
                  Sign Up
                </Link>
              </>
            )}
          </nav>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-12">
        <div className="text-center">
          <h1 className="text-5xl font-bold mb-6 bg-gradient-to-r from-primary to-secondary bg-clip-text text-transparent">
            Play GameCube & N64 Online
          </h1>
          <p className="text-xl text-gray-400 mb-8 max-w-2xl mx-auto">
            Join lobbies, compete with friends, and climb the leaderboard. 
            Deterministic lockstep netplay ensures fair, synchronized gameplay.
          </p>
          
          <div className="flex justify-center gap-4">
            <Link
              to="/lobbies"
              className="px-8 py-4 bg-primary hover:bg-indigo-500 rounded-xl font-semibold text-lg transition"
            >
              Find a Game
            </Link>
            {!session && (
              <Link
                to="/register"
                className="px-8 py-4 bg-secondary hover:bg-purple-500 rounded-xl font-semibold text-lg transition"
              >
                Create Account
              </Link>
            )}
          </div>
        </div>

        {/* Features Grid */}
        <div className="grid md:grid-cols-3 gap-8 mt-20">
          <FeatureCard
            icon="🏆"
            title="Competitive Netplay"
            description="Deterministic lockstep synchronization ensures all players see the same game state."
          />
          <FeatureCard
            icon="👥"
            title="Up to 4 Players"
            description="Host or join lobbies with split-screen support for multiplayer mayhem."
          />
          <FeatureCard
            icon="⚡"
            title="Low Latency"
            description="Input buffering and prediction keep gameplay smooth even with network jitter."
          />
        </div>

        {/* Recent Games Placeholder */}
        <div className="mt-20">
          <h2 className="text-2xl font-bold mb-6">Popular Games</h2>
          <div className="grid md:grid-cols-4 gap-4">
            {['Super Smash Bros.', 'Mario Kart 64', 'GoldenEye 007', 'Melee'].map((game) => (
              <div
                key={game}
                className="bg-dark rounded-lg p-4 hover:bg-gray-800 transition cursor-pointer"
              >
                <div className="aspect-video bg-gray-700 rounded mb-3"></div>
                <h3 className="font-semibold">{game}</h3>
                <p className="text-sm text-gray-400">N64 / GameCube</p>
              </div>
            ))}
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-dark border-t border-gray-800 mt-20 py-8">
        <div className="max-w-7xl mx-auto px-6 text-center text-gray-400">
          <p>Netplay Platform © 2024. Built with Go + React.</p>
          <p className="text-sm mt-2">
            This platform requires you to own legal copies of any games played.
          </p>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ icon, title, description }: { icon: string; title: string; description: string }) {
  return (
    <div className="bg-dark rounded-xl p-6 border border-gray-800 hover:border-primary/50 transition">
      <div className="text-4xl mb-4">{icon}</div>
      <h3 className="text-xl font-bold mb-2">{title}</h3>
      <p className="text-gray-400">{description}</p>
    </div>
  );
}
