# Netplay Platform

A production-grade web-based netplay platform for GameCube & N64 emulation with up to 4 players per lobby.

## Architecture

- **Backend:** Go 1.21+ with chi router, gorilla/websocket, SQLite
- **Frontend:** TypeScript, React 18, Vite, TailwindCSS
- **Netplay:** Deterministic lockstep synchronization at 60Hz tick rate

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker (optional, for containerized dev)

### Backend Setup

```bash
cd backend
go mod download
go run .
```

Server runs on `http://localhost:8080`

### Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

Frontend runs on `http://localhost:5173`

### Using Docker

```bash
docker-compose up --build
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| SERVER_PORT | 8080 | HTTP server port |
| DB_PATH | ./data/netplay.db | SQLite database path |
| TICK_RATE | 60 | Netplay tick rate (FPS) |
| CORS_ORIGINS | http://localhost:5173 | Allowed CORS origins |
| SESSION_SECRET | auto-generated | Session signing secret |

## Testing

### Backend Tests
```bash
cd backend
go test ./...
```

### Frontend Tests
```bash
cd frontend
npm test
```

### E2E Tests (Playwright)
```bash
cd frontend
npx playwright install
npm run test:e2e
```

## WebSocket Protocol

Messages follow this structure:
```json
{
  "type": "input_packet",
  "protocol_version": "1.0.0",
  "payload": {...},
  "timestamp": 1703001234567
}
```

See `frontend/src/types/protocol.ts` for full type definitions.

## Emulator Integration

The platform uses an `EmulatorBridge` interface to connect to WASM-based emulator cores.

### To integrate a real emulator:

1. Place WASM files in `frontend/public/emulator/`
2. Update `frontend/src/game/EmulatorBridge.ts` to load your core
3. Map emulator API to the bridge interface methods

Supported cores:
- EmulatorJS
- Mupen64Plus WASM
- Any core supporting deterministic execution

## Project Structure

```
netplay-platform/
├── backend/
│   ├── main.go              # Entry point
│   ├── handlers.go          # HTTP handlers
│   └── internal/
│       ├── auth/            # Authentication
│       ├── lobby/           # Lobby management
│       ├── netplay/         # WebSocket & sync
│       ├── db/              # Database layer
│       └── config/          # Configuration
├── frontend/
│   ├── src/
│   │   ├── types/           # TypeScript types
│   │   ├── hooks/           # React hooks
│   │   ├── auth/            # Auth components
│   │   ├── lobby/           # Lobby components
│   │   ├── game/            # Game room & emulator
│   │   └── leaderboard/     # Stats display
│   └── tests/               # Test files
└── docker-compose.yml
```

## Key Features

- ✅ User authentication (bcrypt passwords)
- ✅ Lobby creation/joining (max 4 players)
- ✅ Ready check system
- ✅ Deterministic lockstep netplay
- ✅ Split-screen & focus view modes
- ✅ Leaderboard tracking
- ✅ Auto-reconnect with exponential backoff
- ✅ Input buffering for latency compensation

## Known Limitations

1. **No rollback netcode** - Uses basic lockstep; high latency causes frame stalls
2. **No host migration** - If host disconnects, lobby closes
3. **Basic anti-cheat** - Checksum validation only; no input verification
4. **In-memory sessions** - Sessions lost on restart (use Redis for production)
5. **Stub emulator** - Requires WASM core integration for actual gameplay

## Next Steps

1. Integrate real WASM emulator core
2. Implement rollback netcode for better latency handling
3. Add host migration support
4. Implement proper anti-cheat (input validation, replay verification)
5. Add voice chat via WebRTC
6. Deploy to production with PostgreSQL and Redis

## License

MIT License - See LICENSE file for details

**Note:** Users must own legal copies of any games played on this platform.
