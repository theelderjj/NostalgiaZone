/**
 * EmulatorJS Loader Stub
 * 
 * This is a stub loader for integration testing.
 * Replace with actual EmulatorJS loader from https://www.emulatorjs.org/
 * 
 * Required files for production:
 * - loader.js (this file)
 * - cores/mupen64plus.wasm (N64 core)
 * - cores/dolphin.wasm (GameCube core)
 */

(function() {
  'use strict';

  // Global Loader class that EmulatorBridge expects
  window.Loader = class Loader {
    constructor(config) {
      this.config = config;
      this.canvas = config.canvas;
      this.core = config.core;
      this.paused = true;
      this.currentRom = null;
      
      console.log('[EmulatorJS Loader] Initialized with config:', config);
      
      // Set up canvas context for rendering
      this.ctx = this.canvas.getContext('2d');
      if (this.ctx) {
        this.ctx.fillStyle = '#000';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
        this.ctx.fillStyle = '#fff';
        this.ctx.font = '20px Arial';
        this.ctx.textAlign = 'center';
        this.ctx.fillText('EmulatorJS Stub - Load a ROM to start', this.canvas.width / 2, this.canvas.height / 2);
      }
    }

    loadRom(romPath) {
      console.log('[EmulatorJS Loader] Loading ROM:', romPath);
      this.currentRom = romPath;
      
      // Simulate game start callback
      if (this.config.onGameStart) {
        setTimeout(() => {
          console.log('[EmulatorJS Loader] Game started');
          this.config.onGameStart();
        }, 100);
      }
      
      return Promise.resolve();
    }

    pause() {
      console.log('[EmulatorJS Loader] Paused');
      this.paused = true;
    }

    play() {
      console.log('[EmulatorJS Loader] Playing');
      this.paused = false;
    }

    restart() {
      console.log('[EmulatorJS Loader] Restarting');
      this.loadRom(this.currentRom);
    }

    setVolume(volume) {
      console.log('[EmulatorJS Loader] Volume:', volume);
    }

    setInputState(playerNum, state) {
      // Store input state for the player
      if (!this.inputStates) this.inputStates = {};
      this.inputStates[playerNum] = state;
      // In a real implementation, this would update the emulator's input state
    }

    step() {
      // Execute one frame
      if (this.config.onFrame && !this.paused) {
        this.config.onFrame();
      }
      
      // Clear and redraw canvas with current input state for visualization
      if (this.ctx) {
        this.ctx.fillStyle = '#000';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
        
        let statusText = this.paused ? 'PAUSED' : 'RUNNING';
        if (this.currentRom) {
          statusText += ' - ROM: ' + this.currentRom;
        }
        
        this.ctx.fillStyle = '#fff';
        this.ctx.font = '16px Arial';
        this.ctx.textAlign = 'center';
        this.ctx.fillText(statusText, this.canvas.width / 2, 30);
        
        // Display input states
        if (this.inputStates) {
          this.ctx.font = '12px Arial';
          let y = 60;
          for (const [playerNum, state] of Object.entries(this.inputStates)) {
            const buttons = [];
            for (const [btn, pressed] of Object.entries(state)) {
              if (pressed) buttons.push(btn.toUpperCase());
            }
            if (buttons.length > 0) {
              this.ctx.fillText(`P${playerNum}: ${buttons.join(' ')}`, this.canvas.width / 2, y);
              y += 20;
            }
          }
        }
      }
    }
  };

  console.log('[EmulatorJS Loader] Stub loader loaded successfully');
})();
