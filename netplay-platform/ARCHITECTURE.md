# Netplay Platform Architecture

## Overview
This is a production-grade web-based netplay platform for GameCube & N64 emulation with up to 4 players per lobby. The architecture uses **deterministic lockstep synchronization** where all clients run the same emulator core (WASM-based) and synchronize inputs rather than game states. The host validates and broadcasts aggregated input packets at a fixed tick rate (60Hz), while clients buffer inputs and execute frames deterministically.

## Key Design Decisions

1. **Deterministic Lockstep**: All clients simulate the same frames with identical inputs. If inputs match, game states remain synchronized. This minimizes bandwidth while ensuring consistency.

2. **Host-Authoritative Model**: The lobby host collects inputs from all players, validates them, and broadcasts the aggregated input packet. This prevents cheating and simplifies conflict resolution.

3. **Input Buffering**: Clients maintain a buffer of future inputs (typically 8-16 frames) to absorb network jitter. If the buffer runs low, the emulator stalls until more inputs arrive.

4. **WebSocket Hub**: Backend maintains persistent WebSocket connections for real-time communication. Uses goroutines with channels for concurrent message handling.

5. **Emulator Bridge Pattern**: Frontend exposes a clean interface (`EmulatorBridge`) that the WASM emulator core calls for input/output. This allows swapping emulator cores without changing netplay logic.

6. **View Management**: Client-side only feature. Each user can toggle between split-screen (2x2 grid) and focus mode (single player fullscreen) without affecting sync.

## Data Flow

```
[Client Input] → [Input Buffer] → [WS: Send to Host]
                                      ↓
[Host: Aggregate Inputs] → [WS: Broadcast to All]
                                      ↓
[Client: Receive Packet] → [Input Queue] → [Emulator Frame Execution]
```

## Protocol Versioning
All WebSocket messages include a `protocolVersion` field. Breaking changes increment the major version, allowing graceful degradation or rejection of incompatible clients.
