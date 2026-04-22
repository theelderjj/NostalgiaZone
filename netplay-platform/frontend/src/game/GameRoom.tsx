import { useEffect, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useWebSocket } from '../hooks/useWebSocket';
import { useInput } from '../hooks/useInput';
import { useView } from '../hooks/useView';
import { useNetplayClient } from '../netplay/useNetplayClient';
import { createEmulatorBridge } from '../game/EmulatorBridge';
import { InputPacket } from '../types/protocol';

export default function GameRoom() {
  const { lobbyId } = useParams<{ lobbyId: string }>();
  const session = JSON.parse(localStorage.getItem('netplay_session') || '{}');
  
  const [latency, setLatency] = useState(0);
  const [fps, setFps] = useState(0);
  const [gameStarted, setGameStarted] = useState(false);
  
  // WebSocket connection
  const { isConnected, sendMessage, lastLatency } = useWebSocket({
    url: `ws://${window.location.host}/ws`,
    token: session?.token,
    onMessage: handleMessage,
  });
  
  // Netplay client
  const netplay = useNetplayClient({
    sendMessage,
    playerNum: 1, // Should be determined from lobby
    onFrame: (frame) => setFps(Math.round(frame / (performance.now() / 1000))),
  });
  
  // Input handling
  const { inputState, setFrame } = useInput({
    playerNum: 1,
    enabled: gameStarted,
    onInputChange: (input) => {
      setFrame(input.frame);
      netplay.sendInput(input);
    },
  });
  
  // View management
  const view = useView({ playerCount: 4, focusedPlayerNum: 1 });
  
  // Emulator canvas ref
  const canvasRef = useRef<HTMLCanvasElement>(null);
  
  // Initialize emulator on mount
  useEffect(() => {
    const initEmulator = async () => {
      const bridge = await createEmulatorBridge();
      netplay.setEmulator(bridge);
      
      if (canvasRef.current && bridge.getCanvas()) {
        canvasRef.current.replaceWith(bridge.getCanvas()!);
      }
    };
    
    initEmulator();
    
    return () => {
      netplay.setEmulator(null);
    };
  }, []);
  
  // Handle WebSocket messages
  function handleMessage(msg: any) {
    switch (msg.type) {
      case 'game_started':
        setGameStarted(true);
        netplay.start();
        break;
      case 'input_packet':
        netplay.handleInputPacket(msg.payload as InputPacket);
        break;
      case 'game_ended':
        setGameStarted(false);
        netplay.stop();
        break;
    }
  }
  
  // Update latency display
  useEffect(() => {
    setLatency(lastLatency);
  }, [lastLatency]);
  
  const qualityRating = latency < 50 ? 'excellent' : latency < 100 ? 'good' : latency < 200 ? 'fair' : 'poor';
  
  return (
    <div className="h-screen bg-black flex flex-col">
      {/* HUD */}
      <div className="bg-gray-900 px-4 py-2 flex items-center justify-between text-sm">
        <div className="flex items-center gap-4">
          <span className={`connection-quality quality-${qualityRating}`}>
            ● {latency}ms
          </span>
          <span className="text-gray-400">FPS: {fps}</span>
          <span className="text-gray-400">Frame: {netplay.currentFrame}</span>
        </div>
        
        <div className="flex items-center gap-2">
          <button
            onClick={view.toggleViewMode}
            className="px-3 py-1 bg-gray-800 hover:bg-gray-700 rounded transition"
          >
            {view.viewMode === 'split' ? 'Focus Mode' : 'Split View'}
          </button>
          <span className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
        </div>
      </div>
      
      {/* Game Container */}
      <div className={`flex-1 ${view.getContainerClass()} p-1`}>
        {[1, 2, 3, 4].map((playerNum) => (
          <div
            key={playerNum}
            className={`player-view ${!view.isPlayerVisible(playerNum) ? 'hidden' : ''}`}
          >
            <div className="player-label">Player {playerNum}</div>
            <canvas ref={playerNum === 1 ? canvasRef : undefined} className="w-full h-full" />
          </div>
        ))}
      </div>
      
      {/* Controls Help */}
      {!gameStarted && (
        <div className="absolute inset-0 bg-black/80 flex items-center justify-center">
          <div className="bg-dark p-8 rounded-xl max-w-md text-center">
            <h2 className="text-2xl font-bold mb-4">Waiting for Host...</h2>
            <p className="text-gray-400 mb-6">
              The host will start the game when all players are ready.
            </p>
            <div className="text-left text-sm text-gray-500">
              <h3 className="font-semibold mb-2">Controls:</h3>
              <ul className="space-y-1">
                <li>WASD / Arrow Keys - Move</li>
                <li>Z - A Button</li>
                <li>X - B Button</li>
                <li>C/V - X/Y Buttons</li>
                <li>A - Z Button</li>
                <li>S/D - L/R Triggers</li>
                <li>Enter - Start</li>
              </ul>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
