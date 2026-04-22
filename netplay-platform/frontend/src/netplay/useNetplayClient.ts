import { useEffect, useRef, useCallback, useState } from 'react';
import { Input, InputPacket } from '../../types/protocol';
import { EmulatorBridge } from '../../types/protocol';

export interface UseNetplayClientOptions {
  sendMessage: (type: string, payload?: unknown) => void;
  playerNum: number;
  onFrame?: (frame: number) => void;
  onDesync?: () => void;
}

/**
 * Custom hook for netplay client logic
 * Handles input buffering, frame synchronization, and emulator integration
 */
export function useNetplayClient({
  sendMessage,
  playerNum,
  onFrame,
  onDesync,
}: UseNetplayClientOptions) {
  const [currentFrame, setCurrentFrame] = useState(0);
  const [isRunning, setIsRunning] = useState(false);
  const [bufferSize, setBufferSize] = useState(8); // Frames to buffer ahead
  
  // Input buffer: frame number -> aggregated inputs from all players
  const inputBufferRef = useRef<Map<number, InputPacket>>(new Map());
  
  // Local inputs waiting to be sent
  const pendingInputsRef = useRef<Input[]>([]);
  
  // Last received frame from host
  const lastRecvFrameRef = useRef<number>(-1);
  
  // Emulator bridge reference
  const emulatorRef = useRef<EmulatorBridge | null>(null);
  
  // Animation frame for game loop
  const animationFrameRef = useRef<number | null>(null);
  
  // Target frame rate (60 FPS)
  const targetFPS = 60;
  const frameDelay = 1000 / targetFPS;
  
  /**
   * Process incoming input packet from host
   * Adds packet to input buffer for frame execution
   */
  const handleInputPacket = useCallback((packet: InputPacket) => {
    // Validate checksum (basic anti-tamper)
    const calculatedChecksum = calculateChecksum(packet);
    if (calculatedChecksum !== packet.checksum) {
      console.warn('[Netplay] Checksum mismatch! Possible desync');
      onDesync?.();
      return;
    }
    
    // Store in buffer
    inputBufferRef.current.set(packet.frame, packet);
    lastRecvFrameRef.current = packet.frame;
    
    // Update buffer size metric
    setBufferSize(inputBufferRef.current.size);
  }, [onDesync]);
  
  /**
   * Send local input to host
   */
  const sendInput = useCallback((input: Input) => {
    pendingInputsRef.current.push(input);
    
    // Batch send inputs (could optimize further)
    if (pendingInputsRef.current.length >= 1) {
      const inputsToSend = [...pendingInputsRef.current];
      pendingInputsRef.current = [];
      
      for (const inp of inputsToSend) {
        sendMessage('input', { input: inp });
      }
    }
  }, [sendMessage]);
  
  /**
   * Get inputs for a specific frame
   * Returns null if frame not in buffer yet
   */
  const getInputsForFrame = useCallback((frame: number): InputPacket | null => {
    return inputBufferRef.current.get(frame) || null;
  }, []);
  
  /**
   * Main game loop
   * Runs at fixed 60 FPS, executing frames when inputs are available
   */
  const gameLoop = useCallback(() => {
    if (!isRunning || !emulatorRef.current) {
      animationFrameRef.current = requestAnimationFrame(gameLoop);
      return;
    }
    
    const emulator = emulatorRef.current;
    const nextFrame = emulator.getCurrentFrame() + 1;
    
    // Check if we have inputs for the next frame
    const inputs = getInputsForFrame(nextFrame);
    
    if (inputs) {
      // Apply inputs for all players
      for (const [pNum, input] of Object.entries(inputs.player_inputs)) {
        emulator.sendInput(parseInt(pNum), input.frame, input);
      }
      
      // Run emulator frame
      const success = emulator.runFrame();
      
      if (success) {
        setCurrentFrame(nextFrame);
        onFrame?.(nextFrame);
        
        // Remove processed frame from buffer (keep some history for rollback)
        if (nextFrame > 10) {
          inputBufferRef.current.delete(nextFrame - 10);
        }
      }
    } else {
      // Waiting for inputs - could implement frame delay/stall here
      // For now, just continue polling
    }
    
    animationFrameRef.current = requestAnimationFrame(gameLoop);
  }, [isRunning, getInputsForFrame, onFrame]);
  
  /**
   * Start the game loop
   */
  const start = useCallback(() => {
    console.log('[Netplay] Starting game loop');
    setIsRunning(true);
  }, []);
  
  /**
   * Stop the game loop
   */
  const stop = useCallback(() => {
    console.log('[Netplay] Stopping game loop');
    setIsRunning(false);
  }, []);
  
  /**
   * Set emulator bridge instance
   */
  const setEmulator = useCallback((emulator: EmulatorBridge | null) => {
    emulatorRef.current = emulator;
  }, []);
  
  /**
   * Reset state for new game
   */
  const reset = useCallback(() => {
    inputBufferRef.current.clear();
    pendingInputsRef.current = [];
    lastRecvFrameRef.current = -1;
    setCurrentFrame(0);
    setBufferSize(8);
  }, []);
  
  /**
   * Handle reconnection - request missing frames
   */
  const reconnect = useCallback((lastReceivedFrame: number) => {
    console.log('[Netplay] Reconnecting, last frame:', lastReceivedFrame);
    sendMessage('reconnect', {
      lobby_id: '', // Should be set from context
      last_frame_recv: lastReceivedFrame,
    });
  }, [sendMessage]);
  
  // Start game loop on mount
  useEffect(() => {
    animationFrameRef.current = requestAnimationFrame(gameLoop);
    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
    };
  }, [gameLoop]);
  
  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (emulatorRef.current) {
        emulatorRef.current.destroy();
      }
    };
  }, []);
  
  return {
    currentFrame,
    isRunning,
    bufferSize,
    lastRecvFrame: lastRecvFrameRef.current,
    handleInputPacket,
    sendInput,
    start,
    stop,
    setEmulator,
    reset,
    reconnect,
  };
}

/**
 * Calculate checksum for input packet validation
 * Must match backend implementation exactly
 */
function calculateChecksum(packet: InputPacket): number {
  let sum = packet.frame >>> 0;
  
  for (const [, input] of Object.entries(packet.player_inputs)) {
    sum += input.buttons >>> 0;
    sum += (input.axis_x as number) & 0xff;
    sum += (input.axis_y as number) & 0xff;
    sum += (input.c_axis_x as number) & 0xff;
    sum += (input.c_axis_y as number) & 0xff;
    sum += input.trigger_l;
    sum += input.trigger_r;
  }
  
  return sum >>> 0;
}
