import { EmulatorBridge, Input } from '../types/protocol';

/**
 * Stub implementation of EmulatorBridge for development/testing
 * 
 * To use a real emulator:
 * 1. Load your WASM core (e.g., emulatorjs, mupen64plus)
 * 2. Implement the EmulatorBridge interface
 * 3. Replace this factory with your implementation
 * 
 * Example integration points:
 * - loadROM: Call emulator's ROM loading function
 * - sendInput: Forward inputs to emulator's controller API
 * - runFrame: Call emulator's step/frame function
 * - getCanvas: Return emulator's rendering canvas
 */
export async function createEmulatorBridge(): Promise<EmulatorBridge> {
  // In production, this would load and initialize the WASM core
  // For example:
  // const Module = await import('../lib/emulator-wasm');
  // await Module.default();
  
  console.log('[EmulatorBridge] Creating stub emulator bridge');
  console.warn('[EmulatorBridge] This is a stub. Integrate a real WASM emulator core.');
  
  let currentFrame = 0;
  let seed = 0;
  let canvas: HTMLCanvasElement | null = null;
  let isReady = false;

  // Create a placeholder canvas for development
  const initCanvas = () => {
    if (!canvas) {
      canvas = document.createElement('canvas');
      canvas.width = 640;
      canvas.height = 480;
      const ctx = canvas.getContext('2d');
      if (ctx) {
        ctx.fillStyle = '#000';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = '#fff';
        ctx.font = '24px Arial';
        ctx.textAlign = 'center';
        ctx.fillText('Emulator Not Loaded', canvas.width / 2, canvas.height / 2);
        ctx.font = '16px Arial';
        ctx.fillText('Integrate WASM core to play', canvas.width / 2, canvas.height / 2 + 30);
      }
    }
  };

  return {
    async loadROM(romData: ArrayBuffer): Promise<void> {
      console.log('[EmulatorBridge] Loading ROM...', romData.byteLength, 'bytes');
      
      // TODO: Integrate real emulator
      // Example for emulatorjs:
      // await EJS_player.startGame({
      //   core: 'mupen64plus',
      //   game: romData,
      // });
      
      initCanvas();
      isReady = true;
      console.log('[EmulatorBridge] ROM loaded (stub)');
    },

    sendInput(playerNum: number, frame: number, input: Input): void {
      // TODO: Forward input to emulator
      // Example:
      // emulator.setControllerState(playerNum - 1, {
      //   buttons: input.buttons,
      //   stickX: input.axis_x,
      //   stickY: input.axis_y,
      //   cStickX: input.c_axis_x,
      //   cStickY: input.c_axis_y,
      //   l: input.trigger_l,
      //   r: input.trigger_r,
      // });
      
      currentFrame = Math.max(currentFrame, frame);
    },

    runFrame(): boolean {
      if (!isReady) return false;
      
      // TODO: Run emulator frame
      // Example:
      // emulator.step();
      
      currentFrame++;
      
      // Update canvas display with frame info
      if (canvas) {
        const ctx = canvas.getContext('2d');
        if (ctx) {
          ctx.fillStyle = '#000';
          ctx.fillRect(0, 0, canvas.width, 20);
          ctx.fillStyle = '#0f0';
          ctx.font = '12px monospace';
          ctx.textAlign = 'left';
          ctx.fillText(`Frame: ${currentFrame}`, 10, 15);
        }
      }
      
      return true;
    },

    getCurrentFrame(): number {
      return currentFrame;
    },

    setSeed(newSeed: number): void {
      seed = newSeed;
      console.log('[EmulatorBridge] Seed set:', seed);
      // TODO: Pass seed to emulator for deterministic initialization
    },

    getCanvas(): HTMLCanvasElement | null {
      initCanvas();
      return canvas;
    },

    destroy(): void {
      console.log('[EmulatorBridge] Destroying emulator');
      isReady = false;
      currentFrame = 0;
      // TODO: Cleanup emulator resources
      // emulator.destroy();
    },

    isReady(): boolean {
      return isReady;
    },
  };
}

/**
 * Placeholder for where the real emulator WASM would be loaded
 * 
 * Integration guide:
 * 1. Install emulatorjs or build mupen64plus WASM
 * 2. Copy WASM files to frontend/public/emulator/
 * 3. Update this file to load and initialize the emulator
 * 4. Map emulator's API to the EmulatorBridge interface
 * 
 * Key considerations:
 * - Ensure deterministic execution (same inputs = same outputs)
 * - Handle async ROM loading
 * - Manage memory properly
 * - Support multiple controller inputs
 * - Synchronize audio/video
 */
export const EMULATOR_INTEGRATION_GUIDE = `
# Emulator Integration Guide

## Option 1: EmulatorJS

1. Install: npm install @emulatorjs/core
2. Import and initialize in createEmulatorBridge()
3. Map EJS_player API to EmulatorBridge methods

## Option 2: Mupen64Plus WASM

1. Build from source with Emscripten
2. Load WASM module dynamically
3. Implement SDL layer for canvas/audio
4. Expose controller input functions

## Required Features

- Deterministic lockstep support
- Multiple controller inputs (1-4 players)
- Frame-by-frame execution control
- Canvas rendering access
- ROM loading from ArrayBuffer
- Random seed initialization

## Testing

After integration, verify:
- Same ROM + same inputs = identical gameplay
- No desyncs in local testing
- Input latency is acceptable (< 3 frames)
`;
