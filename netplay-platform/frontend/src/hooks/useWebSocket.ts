import { useState, useEffect, useCallback, useRef } from 'react';
import { PROTOCOL_VERSION, Input, InputPacket, WSMessage, MessageType } from '../types/protocol';

export interface UseWebSocketOptions {
  url: string;
  token: string;
  onMessage?: (message: WSMessage) => void;
  onError?: (error: Event) => void;
  onClose?: () => void;
}

export interface UseWebSocketReturn {
  ws: WebSocket | null;
  isConnected: boolean;
  sendMessage: (type: MessageType, payload?: unknown) => void;
  lastLatency: number;
  reconnect: () => void;
}

/**
 * Custom hook for managing WebSocket connections with auto-reconnect
 */
export function useWebSocket({
  url,
  token,
  onMessage,
  onError,
  onClose,
}: UseWebSocketOptions): UseWebSocketReturn {
  const wsRef = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [lastLatency, setLastLatency] = useState(0);
  const heartbeatIntervalRef = useRef<number | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;
  const reconnectDelay = 1000;

  const cleanup = useCallback(() => {
    if (heartbeatIntervalRef.current) {
      clearInterval(heartbeatIntervalRef.current);
      heartbeatIntervalRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  const connect = useCallback(() => {
    cleanup();

    // Add token to URL query params for WebSocket auth
    const wsUrl = `${url}?token=${encodeURIComponent(token)}`;
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('[WS] Connected');
      setIsConnected(true);
      reconnectAttemptsRef.current = 0;

      // Start heartbeat
      heartbeatIntervalRef.current = window.setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          const msg: WSMessage = {
            type: 'heartbeat',
            protocol_version: PROTOCOL_VERSION,
            timestamp: Date.now(),
          };
          ws.send(JSON.stringify(msg));
        }
      }, 5000);
    };

    ws.onmessage = (event) => {
      try {
        const message: WSMessage = JSON.parse(event.data);
        
        // Handle latency measurement from heartbeat response
        if (message.type === 'heartbeat' && typeof message.payload === 'object' && message.payload !== null) {
          const payload = message.payload as { timestamp?: number };
          if (payload.timestamp) {
            const latency = Date.now() - payload.timestamp;
            setLastLatency(latency);
          }
        }

        onMessage?.(message);
      } catch (err) {
        console.error('[WS] Failed to parse message:', err);
      }
    };

    ws.onerror = (error) => {
      console.error('[WS] Error:', error);
      onError?.(error);
    };

    ws.onclose = () => {
      console.log('[WS] Disconnected');
      setIsConnected(false);
      cleanup();
      onClose?.();

      // Auto-reconnect with exponential backoff
      if (reconnectAttemptsRef.current < maxReconnectAttempts) {
        reconnectAttemptsRef.current++;
        const delay = reconnectDelay * Math.pow(2, reconnectAttemptsRef.current - 1);
        console.log(`[WS] Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current}/${maxReconnectAttempts})`);
        setTimeout(connect, delay);
      } else {
        console.error('[WS] Max reconnect attempts reached');
      }
    };

    wsRef.current = ws;
  }, [url, token, onMessage, onError, onClose, cleanup]);

  const sendMessage = useCallback((type: MessageType, payload?: unknown) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const message: WSMessage = {
        type,
        protocol_version: PROTOCOL_VERSION,
        payload,
        timestamp: Date.now(),
      };
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.warn('[WS] Cannot send message - not connected');
    }
  }, []);

  const reconnect = useCallback(() => {
    reconnectAttemptsRef.current = 0;
    connect();
  }, [connect]);

  useEffect(() => {
    connect();
    return cleanup;
  }, [connect, cleanup]);

  return {
    ws: wsRef.current,
    isConnected,
    sendMessage,
    lastLatency,
    reconnect,
  };
}
