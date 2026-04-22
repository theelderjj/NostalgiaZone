import { useState, useEffect } from 'react';

interface LeaderboardEntry {
  user_id: number;
  username: string;
  wins: number;
  losses: number;
  draws: number;
  play_time: number;
  games_played: number;
  last_game: string;
}

export default function Leaderboard() {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  const fetchLeaderboard = async () => {
    try {
      const res = await fetch('/api/leaderboard');
      if (res.ok) {
        const data = await res.json();
        setEntries(data);
      }
    } catch (err) {
      console.error('Failed to fetch leaderboard:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatTime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  return (
    <div className="min-h-screen bg-darker">
      <div className="max-w-4xl mx-auto px-6 py-8">
        <h1 className="text-3xl font-bold mb-8">Leaderboard</h1>

        {loading ? (
          <div className="text-center py-12 text-gray-400">Loading...</div>
        ) : (
          <div className="bg-dark rounded-xl border border-gray-800 overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-900">
                <tr>
                  <th className="px-6 py-4 text-left text-sm font-semibold">#</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold">Player</th>
                  <th className="px-6 py-4 text-center text-sm font-semibold">W-D-L</th>
                  <th className="px-6 py-4 text-center text-sm font-semibold">Win Rate</th>
                  <th className="px-6 py-4 text-center text-sm font-semibold">Games</th>
                  <th className="px-6 py-4 text-right text-sm font-semibold">Play Time</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800">
                {entries.map((entry, index) => {
                  const totalGames = entry.wins + entry.losses + entry.draws;
                  const winRate = totalGames > 0 
                    ? Math.round((entry.wins / totalGames) * 100) 
                    : 0;

                  return (
                    <tr key={entry.user_id} className="hover:bg-gray-800/50 transition">
                      <td className="px-6 py-4">
                        <span className={`font-bold ${
                          index === 0 ? 'text-yellow-400' :
                          index === 1 ? 'text-gray-400' :
                          index === 2 ? 'text-amber-600' :
                          'text-gray-500'
                        }`}>
                          #{index + 1}
                        </span>
                      </td>
                      <td className="px-6 py-4 font-medium">{entry.username}</td>
                      <td className="px-6 py-4 text-center">
                        <span className="text-green-400">{entry.wins}</span>-
                        <span className="text-gray-400">{entry.draws}</span>-
                        <span className="text-red-400">{entry.losses}</span>
                      </td>
                      <td className="px-6 py-4 text-center">
                        <span className={`px-2 py-1 rounded text-sm ${
                          winRate >= 60 ? 'bg-green-900/50 text-green-400' :
                          winRate >= 40 ? 'bg-yellow-900/50 text-yellow-400' :
                          'bg-red-900/50 text-red-400'
                        }`}>
                          {winRate}%
                        </span>
                      </td>
                      <td className="px-6 py-4 text-center text-gray-400">
                        {entry.games_played}
                      </td>
                      <td className="px-6 py-4 text-right text-gray-400">
                        {formatTime(entry.play_time)}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>

            {entries.length === 0 && (
              <div className="text-center py-12 text-gray-400">
                No matches played yet. Be the first to climb the leaderboard!
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
