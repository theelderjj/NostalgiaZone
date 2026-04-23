import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import './index.css';

// Lazy load components for better initial load time
import App from './App';
import LobbyBrowser from './lobby/LobbyBrowser';
import LobbyRoom from './lobby/LobbyRoom';
import CreateLobby from './lobby/CreateLobby';
import GameRoom from './game/GameRoom';
import Leaderboard from './leaderboard/Leaderboard';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/lobbies" element={<LobbyBrowser />} />
        <Route path="/lobby/create" element={<CreateLobby />} />
        <Route path="/lobby/:lobbyId" element={<LobbyRoom />} />
        <Route path="/game/:lobbyId" element={<GameRoom />} />
        <Route path="/leaderboard" element={<Leaderboard />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);
