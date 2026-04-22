import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import './index.css';

// Lazy load components for better initial load time
import App from './App';
import LoginPage from './auth/LoginPage';
import RegisterPage from './auth/RegisterPage';
import LobbyBrowser from './lobby/LobbyBrowser';
import LobbyRoom from './lobby/LobbyRoom';
import GameRoom from './game/GameRoom';
import Leaderboard from './leaderboard/Leaderboard';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/lobbies" element={<LobbyBrowser />} />
        <Route path="/lobby/:lobbyId" element={<LobbyRoom />} />
        <Route path="/game/:lobbyId" element={<GameRoom />} />
        <Route path="/leaderboard" element={<Leaderboard />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);
