# EmulatorJS Integration

This directory contains the EmulatorJS files required for the netplay platform.

## Required Files

For production use, you need to download EmulatorJS from https://www.emulatorjs.org/

### N64 Core
- `cores/mupen64plus.wasm` - N64 emulator core

### GameCube Core  
- `cores/dolphin.wasm` - GameCube emulator core

### Common Files
- `loader.js` - Main entry point (stub provided for testing)

## Current Setup

Currently, this directory contains a **stub loader** (`loader.js`) for integration testing. The stub:

1. Provides the `window.Loader` class that `EmulatorBridge.ts` expects
2. Simulates ROM loading and frame execution
3. Displays input states on the canvas for visualization
4. Does NOT actually emulate any games

## To Enable Real Emulation

1. Download EmulatorJS from https://www.emulatorjs.org/
2. Copy the following files to this directory:
   - `loader.js` (replace the stub)
   - `cores/mupen64plus.wasm` (for N64)
   - `cores/dolphin.wasm` (for GameCube)
3. Update `EmulatorBridge.ts` if the API differs from the stub

## File Structure

```
frontend/public/emulator/
├── README.md           # This file
├── loader.js           # EmulatorJS loader (stub or real)
└── cores/
    ├── mupen64plus.wasm  # N64 core (download separately)
    └── dolphin.wasm      # GameCube core (download separately)
```

## Testing with Stub

The stub loader allows you to:
- Test the EmulatorBridge integration
- Verify input handling works correctly
- Test the netplay synchronization logic
- Develop UI components without needing full WASM cores

When the stub is active, the canvas will show:
- "EmulatorJS Stub - Load a ROM to start" initially
- "RUNNING" or "PAUSED" status during operation
- Active button presses for each player
