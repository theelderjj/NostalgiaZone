/**
 * EmulatorBridge.ts
 * 
 * Bridges the React Netplay Loop with the EmulatorJS Core.
 * 
 * INTEGRATION INSTRUCTIONS:
 * 1. Download EmulatorJS (N64 or GameCube core) from https://www.emulatorjs.org/
 *    OR use the pre-built package: npm install @emulatorjs/core
 * 2. Extract the `data` folder contents into `frontend/public/emulator/`
 *    - You should see: `loader.js`, `core.wasm`, `roms/`, etc.
 * 3. Ensure `public/emulator/loader.js` exists.
 * 
 * This bridge dynamically loads the EmulatorJS loader script and exposes
 * methods for the Netplay loop to inject deterministic inputs.
 */

import { EmulatorBridge, Input } from '../types/protocol';

export interface EmulatorInstance {
  pause(): void;
  play(): void;
  restart(): void;
  setVolume(volume: number): void;
  loadRom(file: File | string): Promise<void>;
  setInputState(playerNum: number, state: any): void;
  step(): void;
  _instance?: any; 
}

export interface BridgeConfig {
  coreType: 'n64' | 'gc'; // Determines which EmulatorJS core to load
  canvasId: string;
  width: number;
  height: number;
  onReady?: () => void;
  onFrame?: (frameData: ImageData) => void;
}

class EmulatorBridgeImpl {
  private config: BridgeConfig;
  private emulator: EmulatorInstance | null = null;
  private loaderScript: HTMLScriptElement | null = null;
  private isLoaded: boolean = false;
  private inputQueue: Uint8Array[] = [];
  private currentFrame: number = 0;
  private seed: number = 0;
  private canvas: HTMLCanvasElement | null = null;

  constructor(config: BridgeConfig) {
    this.config = config;
  }

  /**
   * Loads the EmulatorJS loader script dynamically.
   * Must be called before loadRom.
   */
  public async loadCore(): Promise<void> {
    if (this.isLoaded) return;

    return new Promise((resolve, reject) => {
      const scriptPath = '/emulator/loader.js';
      
      // Check if already injected
      if (document.querySelector(`script[src="${scriptPath}"]`)) {
        this.onCoreReady();
        resolve();
        return;
      }

      const script = document.createElement('script');
      script.src = scriptPath;
      script.type = 'text/javascript';
      script.onload = () => this.onCoreReady();
      script.onerror = () => reject(new Error(`Failed to load EmulatorJS from ${scriptPath}. Ensure you have downloaded EmulatorJS files to frontend/public/emulator/`));
      
      document.body.appendChild(script);
      this.loaderScript = script;
    });
  }

  private onCoreReady() {
    // EmulatorJS exposes a global 'Loader' object
    if (!(window as any).Loader) {
      console.error('EmulatorJS loaded but window.Loader not found.');
      return;
    }

    this.canvas = document.getElementById(this.config.canvasId) as HTMLCanvasElement;
    if (!this.canvas) {
      console.error(`Canvas with ID ${this.config.canvasId} not found.`);
      return;
    }

    // Initialize EmulatorJS with netplay-friendly settings
    const loader = new (window as any).Loader({
      core: this.config.coreType === 'n64' ? 'mupen64plus' : 'dolphin',
      rom: null,
      canvas: this.canvas,
      width: this.config.width,
      height: this.config.height,
      disableContextMenu: true,
      startOnLoad: false,
      pauseOnBlur: true,
      frameSkip: 0, // Disable frameskip for deterministic sync
      onGameStart: () => {
        console.log('Game started, netplay sync active.');
        this.isLoaded = true;
        if (this.config.onReady) this.config.onReady();
      },
      onFrame: () => {
        if (this.config.onFrame && this.canvas) {
           const ctx = this.canvas.getContext('2d');
           if (ctx) this.config.onFrame(ctx.getImageData(0, 0, this.canvas.width, this.canvas.height));
        }
        this.processInputQueue();
      }
    });

    this.emulator = {
      _instance: loader,
      pause: () => loader.pause(),
      play: () => loader.play(),
      restart: () => loader.restart(),
      setVolume: (vol) => loader.setVolume(vol),
      loadRom: (file) => this.loadRomIntoLoader(loader, file),
      setInputState: (playerNum, state) => loader.setInputState(playerNum, state),
      step: () => loader.step()
    };

    console.log('[EmulatorBridge] Core initialized successfully');
  }

  private async loadRomIntoLoader(loader: any, file: File | string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        if (typeof file === 'string') {
          loader.loadRom(file);
        } else {
          const url = URL.createObjectURL(file);
          loader.loadRom(url);
        }
        resolve();
      } catch (e) {
        reject(e);
      }
    });
  }

  /**
   * Queues an input packet for the next frame.
   */
  public queueInput(inputPayload: Uint8Array) {
    this.inputQueue.push(inputPayload);
  }

  /**
   * Called internally by the EmulatorJS onFrame hook.
   */
  private processInputQueue() {
    if (this.inputQueue.length === 0) {
      // In strict lockstep, we might want to pause if buffer is empty
      // For now, repeat last input or send neutral
      return;
    }

    const nextInput = this.inputQueue.shift();
    if (nextInput && this.emulator?._instance) {
      this.applyInputToEmulator(nextInput);
      this.currentFrame++;
    }
  }

  /**
   * Maps the netplay input payload to EmulatorJS specific key presses.
   * Protocol: Byte 0 = Player 1 buttons bitmask
   */
  private applyInputToEmulator(input: Uint8Array) {
    const instance = this.emulator?._instance;
    if (!instance) return;

    const player1Input = input[0]; 
    
    const buttons = {
      up: (player1Input & 0x01) !== 0,
      down: (player1Input & 0x02) !== 0,
      left: (player1Input & 0x04) !== 0,
      right: (player1Input & 0x08) !== 0,
      a: (player1Input & 0x10) !== 0,
      b: (player1Input & 0x20) !== 0,
      z: (player1Input & 0x40) !== 0,
      start: (player1Input & 0x80) !== 0,
    };

    instance.setInputState(1, buttons);
  }

  public getEmulator(): EmulatorInstance | null {
    return this.emulator;
  }

  public destroy() {
    if (this.emulator) {
      this.emulator.pause();
      this.emulator = null;
    }
    if (this.loaderScript) {
      this.loaderScript.remove();
      this.loaderScript = null;
    }
    this.isLoaded = false;
    this.inputQueue = [];
    this.currentFrame = 0;
  }
}

/**
 * Factory function to create the EmulatorBridge instance
 * Compatible with the existing EmulatorBridge interface from protocol.ts
 */
export async function createEmulatorBridge(coreType: 'n64' | 'gc' = 'n64'): Promise<EmulatorBridge> {
  const canvasId = 'emulator-canvas';
  
  // Create canvas if it doesn't exist
  if (!document.getElementById(canvasId)) {
    const canvas = document.createElement('canvas');
    canvas.id = canvasId;
    canvas.width = 640;
    canvas.height = 480;
    canvas.style.display = 'none'; // Hidden until game starts
    document.body.appendChild(canvas);
  }

  const bridge = new EmulatorBridgeImpl({
    coreType,
    canvasId,
    width: 640,
    height: 480,
  });

  await bridge.loadCore();

  return {
    async loadROM(romData: ArrayBuffer): Promise<void> {
      const emulator = bridge.getEmulator();
      if (!emulator) throw new Error('Emulator not initialized');
      
      const blob = new Blob([romData], { type: 'application/octet-stream' });
      const file = new File([blob], 'game.rom', { type: 'application/octet-stream' });
      await emulator.loadRom(file);
    },

    sendInput(playerNum: number, frame: number, input: Input): void {
      // Convert Input to Uint8Array based on protocol
      const payload = new Uint8Array(4); // Support up to 4 players
      
      // Simple mapping: pack button states into bytes
      // Adjust based on your actual Input interface
      let playerByte = 0;
      if (input.buttons & 0x0001) playerByte |= 0x01; // A
      if (input.buttons & 0x0002) playerByte |= 0x02; // B
      if (input.buttons & 0x0004) playerByte |= 0x04; // Z
      if (input.buttons & 0x0008) playerByte |= 0x08; // Start
      if (input.buttons & 0x0010) playerByte |= 0x10; // Up
      if (input.buttons & 0x0020) playerByte |= 0x20; // Down
      if (input.buttons & 0x0040) playerByte |= 0x40; // Left
      if (input.buttons & 0x0080) playerByte |= 0x80; // Right
      
      payload[playerNum - 1] = playerByte;
      
      bridge.queueInput(payload);
    },

    runFrame(): boolean {
      const emulator = bridge.getEmulator();
      if (!emulator) return false;
      
      emulator.step();
      return true;
    },

    getCurrentFrame(): number {
      return bridge['currentFrame'];
    },

    setSeed(newSeed: number): void {
      bridge['seed'] = newSeed;
      console.log('[EmulatorBridge] Seed set:', newSeed);
    },

    getCanvas(): HTMLCanvasElement | null {
      const element = document.getElementById(canvasId);
      if (element && element instanceof HTMLCanvasElement) {
        element.style.display = 'block';
        return element;
      }
      return null;
    },

    destroy(): void {
      bridge.destroy();
    },

    isReady(): boolean {
      return bridge['isLoaded'];
    },
    
    // Additional EmulatorJS-specific methods if needed
    pause: () => bridge.getEmulator()?.pause(),
    play: () => bridge.getEmulator()?.play(),
  };
}

export const EMULATOR_INTEGRATION_GUIDE = `
# Emulator Integration Guide

## Setup Steps

1. Download EmulatorJS from https://www.emulatorjs.org/
   OR install via npm: npm install @emulatorjs/core

2. Copy emulator files to frontend/public/emulator/:
   - loader.js (main entry point)
   - cores/mupen64plus.wasm (for N64)
   - cores/dolphin.wasm (for GameCube)

3. The bridge automatically loads these files at runtime.

## Protocol Mapping

The Input.buttons bitmask maps to emulator buttons as follows:
- Bit 0 (0x0001): A Button
- Bit 1 (0x0002): B Button  
- Bit 2 (0x0004): Z Button
- Bit 3 (0x0008): Start Button
- Bit 4 (0x0010): D-Pad Up
- Bit 5 (0x0020): D-Pad Down
- Bit 6 (0x0040): D-Pad Left
- Bit 7 (0x0080): D-Pad Right

Modify the sendInput() method above if your protocol differs.

## Troubleshooting

- "Failed to load EmulatorJS": Check that files exist in public/emulator/
- "window.Loader not found": Verify loader.js exports the Loader class
- Desyncs: Ensure all clients use identical ROM files and emulator versions
`;
