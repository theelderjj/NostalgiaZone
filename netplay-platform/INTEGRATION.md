# EmulatorJS Integration Guide

This document explains how the emulator files have been integrated into the netplay platform.

## What Was Done

### 1. Created Emulator Directory Structure

```
frontend/public/emulator/
├── .gitkeep              # Placeholder for git tracking
├── README.md             # Documentation for this directory
├── loader.js             # Stub loader for testing
└── cores/
    └── .gitkeep          # Placeholder for WASM cores
```

### 2. Implemented Stub Loader (`loader.js`)

A functional stub loader has been created that:

- Exposes the `window.Loader` class that `EmulatorBridge.ts` expects
- Simulates ROM loading with callbacks
- Provides pause/play/restart functionality
- Handles input state via `setInputState()`
- Executes frames via `step()` method
- Visualizes active inputs on the canvas

The stub allows you to:
- Test the EmulatorBridge integration without real WASM files
- Verify the netplay synchronization logic works
- Develop UI components and test input handling
- Debug the React component lifecycle

### 3. Updated Project Configuration

The `.gitignore` already excludes:
- `*.wasm` files (large binaries)
- `public/emulator/cores/` directory contents

But keeps `.gitkeep` files to track the directory structure.

## How It Works

### Loading Flow

1. **GameRoom.tsx** mounts and calls `createEmulatorBridge()`
2. **EmulatorBridge.ts** loads `/emulator/loader.js` dynamically
3. The loader script creates `window.Loader` class
4. EmulatorBridge instantiates the Loader with config:
   - Core type: 'n64' or 'gc'
   - Canvas element
   - Callbacks for game start and frame events
5. When a ROM is loaded, the emulator starts

### Input Flow

1. User presses keys → **useInput.ts** captures input
2. **NetplayClient** sends input to WebSocket
3. Host aggregates inputs from all players
4. Broadcast aggregated input packet back to clients
5. **EmulatorBridge** receives input packet
6. Converts protocol bytes to button states
7. Calls `loader.setInputState(playerNum, buttons)`
8. On next frame, emulator uses the input state

### Frame Execution Flow

1. **NetplayClient** runs at 60Hz tick rate
2. Each tick, calls `emulator.runFrame()`
3. Which calls `loader.step()`
4. Loader executes one frame of emulation
5. Triggers `onFrame()` callback
6. Netplay updates frame counter and FPS display

## Testing the Integration

### With Stub Loader (Current Setup)

```bash
cd frontend
npm install
npm run dev
```

Navigate to `http://localhost:5173`, create a lobby, and enter game room.

You should see:
- Canvas appears with "EmulatorJS Stub - Load a ROM to start"
- Pressing keys shows button states on canvas
- FPS counter updates as frames are "executed"
- WebSocket connection status indicator

### With Real EmulatorJS

1. Download EmulatorJS from https://www.emulatorjs.org/
2. Copy files to `frontend/public/emulator/`:
   ```
   frontend/public/emulator/
   ├── loader.js           # Replace stub with real one
   └── cores/
       ├── mupen64plus.wasm
       └── dolphin.wasm
   ```
3. Restart the frontend dev server
4. Load a ROM through the UI

## Architecture Reference

See these files for implementation details:

- `frontend/src/game/EmulatorBridge.ts` - Bridge between React and emulator
- `frontend/src/game/GameRoom.tsx` - Game room component
- `frontend/src/netplay/useNetplayClient.ts` - Netplay synchronization loop
- `frontend/src/hooks/useInput.ts` - Keyboard input handling
- `frontend/src/types/protocol.ts` - Message and input types

## Next Steps

1. **Download Real EmulatorJS**: Get actual WASM cores from emulatorjs.org
2. **Test ROM Loading**: Verify ROMs load and games run
3. **Verify Determinism**: Ensure same inputs produce same frames on all clients
4. **Tune Input Buffer**: Adjust buffer size for latency compensation
5. **Add ROM Browser**: Implement file picker for loading ROMs
6. **Implement Save States**: Add save/load state functionality

## Troubleshooting

### "Failed to load EmulatorJS"
- Check that `loader.js` exists in `frontend/public/emulator/`
- Verify the dev server is serving static files from `public/`

### "window.Loader not found"
- The loader script may not have executed yet
- Check browser console for script loading errors
- Ensure `loader.js` properly exposes `window.Loader`

### Desyncs Between Clients
- All clients must use identical ROM files (same hash)
- All clients must use same emulator version
- Check that input packets are being received in order

## Production Deployment

For production builds:

1. Build frontend: `npm run build`
2. Ensure `dist/emulator/` contains all required files
3. Configure web server to serve WASM files with correct MIME types:
   - `.wasm` → `application/wasm`
   - `.js` → `application/javascript`
4. Enable CORS if serving emulator files from CDN

## License Note

EmulatorJS has its own license. Ensure compliance when distributing:
- Users must provide their own ROM files
- Do not bundle copyrighted BIOS files
- Check EmulatorJS license for redistribution terms
