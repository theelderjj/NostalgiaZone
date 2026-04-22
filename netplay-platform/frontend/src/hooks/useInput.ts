import { useState, useEffect, useCallback, useRef } from 'react';
import { Input, BUTTONS } from '../types/protocol';

export interface InputState {
  buttons: number;
  axisX: number;
  axisY: number;
  cAxisX: number;
  cAxisY: number;
  triggerL: number;
  triggerR: number;
}

export interface UseInputOptions {
  playerNum: number;
  enabled?: boolean;
  onInputChange?: (input: Input) => void;
}

/**
 * Custom hook for handling keyboard and gamepad input
 * Maps controller inputs to the netplay Input format
 */
export function useInput({
  playerNum,
  enabled = true,
  onInputChange,
}: UseInputOptions) {
  const [inputState, setInputState] = useState<InputState>({
    buttons: 0,
    axisX: 0,
    axisY: 0,
    cAxisX: 0,
    cAxisY: 0,
    triggerL: 0,
    triggerR: 0,
  });

  const keysPressed = useRef<Set<string>>(new Set());
  const gamepadIndex = useRef<number | null>(null);
  const animationFrameRef = useRef<number | null>(null);
  const lastFrameRef = useRef<number>(0);

  // Key mapping for keyboard controls
  const keyMap = useRef<Record<string, keyof typeof BUTTONS>>({
    // GameCube/N64 style layout
    'z': 'A',
    'x': 'B',
    'c': 'X',
    'v': 'Y',
    'a': 'Z',
    's': 'L',
    'd': 'R',
    'Enter': 'START',
    'ArrowUp': 'UP',
    'ArrowDown': 'DOWN',
    'ArrowLeft': 'LEFT',
    'ArrowRight': 'RIGHT',
    // WASD alternative
    'w': 'UP',
    'a': 'LEFT',
    's': 'DOWN',
    'd': 'RIGHT',
  }).current;

  // Update input state from current key/gamepad state
  const updateInputState = useCallback(() => {
    if (!enabled) return;

    let buttons = 0;
    let axisX = 0;
    let axisY = 0;
    let cAxisX = 0;
    let cAxisY = 0;
    let triggerL = 0;
    let triggerR = 0;

    // Process keyboard input
    for (const key of keysPressed.current) {
      const button = keyMap[key];
      if (button) {
        buttons |= BUTTONS[button];
      }
    }

    // Process gamepad input
    if (gamepadIndex.current !== null) {
      const gamepads = navigator.getGamepads();
      const gamepad = gamepads[gamepadIndex.current];

      if (gamepad) {
        // Button mappings (standard gamepad layout)
        if (gamepad.buttons[0].pressed) buttons |= BUTTONS.A;      // Cross/A
        if (gamepad.buttons[1].pressed) buttons |= BUTTONS.B;      // Circle/B
        if (gamepad.buttons[2].pressed) buttons |= BUTTONS.X;      // Square/X
        if (gamepad.buttons[3].pressed) buttons |= BUTTONS.Y;      // Triangle/Y
        if (gamepad.buttons[4].pressed) buttons |= BUTTONS.L;      // L1
        if (gamepad.buttons[5].pressed) buttons |= BUTTONS.R;      // R1
        if (gamepad.buttons[6].pressed) buttons |= BUTTONS.Z;      // L2
        if (gamepad.buttons[7].pressed) buttons |= BUTTONS.START;  // Start

        // D-pad
        if (gamepad.buttons[12].pressed) buttons |= BUTTONS.UP;
        if (gamepad.buttons[13].pressed) buttons |= BUTTONS.DOWN;
        if (gamepad.buttons[14].pressed) buttons |= BUTTONS.LEFT;
        if (gamepad.buttons[15].pressed) buttons |= BUTTONS.RIGHT;

        // Left stick
        if (Math.abs(gamepad.axes[0]) > 0.1) axisX = Math.round(gamepad.axes[0] * 127);
        if (Math.abs(gamepad.axes[1]) > 0.1) axisY = Math.round(-gamepad.axes[1] * 127);

        // Right stick (C-stick)
        if (gamepad.axes.length >= 4) {
          if (Math.abs(gamepad.axes[2]) > 0.1) cAxisX = Math.round(gamepad.axes[2] * 127);
          if (Math.abs(gamepad.axes[3]) > 0.1) cAxisY = Math.round(-gamepad.axes[3] * 127);
        }

        // Triggers (analog)
        if (gamepad.buttons[6].value !== undefined) {
          triggerL = Math.round(gamepad.buttons[6].value * 255);
        }
        if (gamepad.buttons[7].value !== undefined) {
          triggerR = Math.round(gamepad.buttons[7].value * 255);
        }
      }
    }

    const newState: InputState = {
      buttons,
      axisX,
      axisY,
      cAxisX,
      cAxisY,
      triggerL,
      triggerR,
    };

    setInputState(newState);

    // Notify parent of input change
    if (onInputChange && JSON.stringify(newState) !== JSON.stringify(inputState)) {
      const input: Input = {
        player_num: playerNum,
        frame: lastFrameRef.current,
        buttons: newState.buttons,
        axis_x: newState.axisX,
        axis_y: newState.axisY,
        c_axis_x: newState.cAxisX,
        c_axis_y: newState.cAxisY,
        trigger_l: newState.triggerL,
        trigger_r: newState.triggerR,
      };
      onInputChange(input);
    }
  }, [enabled, playerNum, onInputChange, inputState, keyMap]);

  // Keyboard event handlers
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      keysPressed.current.add(e.key);
      updateInputState();
    };

    const handleKeyUp = (e: KeyboardEvent) => {
      keysPressed.current.delete(e.key);
      updateInputState();
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
    };
  }, [enabled, updateInputState]);

  // Gamepad polling
  useEffect(() => {
    if (!enabled) return;

    // Try to connect gamepad on any button press
    const handleGamepadConnected = () => {
      const gamepads = navigator.getGamepads();
      for (let i = 0; i < gamepads.length; i++) {
        if (gamepads[i]) {
          gamepadIndex.current = i;
          console.log(`[Input] Gamepad connected: ${gamepads[i].id}`);
          break;
        }
      }
    };

    window.addEventListener('gamepadconnected', handleGamepadConnected);
    window.addEventListener('gamepaddisconnected', () => {
      gamepadIndex.current = null;
      console.log('[Input] Gamepad disconnected');
    });

    // Initial check
    handleGamepadConnected();

    // Poll gamepad at 60fps
    const pollGamepad = () => {
      updateInputState();
      animationFrameRef.current = requestAnimationFrame(pollGamepad);
    };
    animationFrameRef.current = requestAnimationFrame(pollGamepad);

    return () => {
      window.removeEventListener('gamepadconnected', handleGamepadConnected);
      window.removeEventListener('gamepaddisconnected', () => {});
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
    };
  }, [enabled, updateInputState]);

  // Update frame counter
  const setFrame = useCallback((frame: number) => {
    lastFrameRef.current = frame;
  }, []);

  // Clear all inputs
  const clearInputs = useCallback(() => {
    keysPressed.current.clear();
    setInputState({
      buttons: 0,
      axisX: 0,
      axisY: 0,
      cAxisX: 0,
      cAxisY: 0,
      triggerL: 0,
      triggerR: 0,
    });
  }, []);

  return {
    inputState,
    setFrame,
    clearInputs,
    hasGamepad: gamepadIndex.current !== null,
  };
}
