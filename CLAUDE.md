# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

**Full build (frontend + backend):**
```bash
make build
```

**Frontend only:**
```bash
npm install
npm run build          # production
npm run watch          # development with hot reload
```

**Backend only (from src/):**
```bash
cd src && go build -o ../factorio-server-manager/factorio-server-manager .
```

**Cross-platform release:**
```bash
make gen_release       # builds Linux and Windows binaries
```

**Clean:**
```bash
make clean
```

## Testing

**Backend tests:**
```bash
cd src && go test ./... -v
cd src && go test ./... -v -test.short   # faster, skips integration tests
```

For mod-related tests, set environment variables in `src/.env`:
- `factorio_username` / `factorio_password` - Factorio portal credentials

**Frontend:** No dedicated test suite; relies on webpack build success.

## Architecture

This is a full-stack web application for managing Factorio game servers.

### Backend (Go) - `/src`
- `api/` - REST API handlers and WebSocket endpoints
- `factorio/` - Factorio server control (saves, mods, config, RCON)
- `bootstrap/` - Configuration and user initialization
- Key dependencies: gorilla/mux (router), gorilla/websocket, gorm+sqlite3 (users DB)

### Frontend (React) - `/ui`
- `App/components/` - Reusable UI components
- `App/views/` - Page components (Controls, Saves, Mods, Logs, Console, Settings)
- `api/` - API client layer with resource-specific modules
- `api/socket.js` - WebSocket for real-time log streaming
- Built with Webpack, outputs to `app/bundle.js` and `app/style.css`
- Uses Tailwind CSS for styling

### Key Integration Points
- Backend serves frontend static files from `app/`
- WebSocket provides real-time game console output
- RCON protocol used for Factorio server communication
- Mod portal integration for downloading mods from official Factorio API

## Docker

Primary deployment method. See `docker/README.md` for details.
- `Dockerfile` - Production (downloads pre-built binaries)
- `Dockerfile-local` - Development (builds from source)
- Docker Hub: `ofsm/ofsm`
