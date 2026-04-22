# Netplay Platform Makefile

.PHONY: dev test build setup-emulator clean docker-up docker-down help

# Default target
help:
	@echo "Netplay Platform - Available Commands:"
	@echo ""
	@echo "  dev              - Start both backend and frontend in development mode"
	@echo "  dev-backend      - Start only the backend server"
	@echo "  dev-frontend     - Start only the frontend dev server"
	@echo "  test             - Run all tests (backend + frontend)"
	@echo "  test-backend     - Run backend Go tests"
	@echo "  test-frontend    - Run frontend Vitest tests"
	@echo "  test-e2e         - Run Playwright E2E tests"
	@echo "  build            - Build both backend and frontend for production"
	@echo "  setup-emulator   - Download EmulatorJS files to frontend/public/emulator/"
	@echo "  docker-up        - Start Docker Compose services"
	@echo "  docker-down      - Stop Docker Compose services"
	@echo "  clean            - Remove build artifacts and database"
	@echo ""

# Development
dev:
	@echo "Starting backend and frontend..."
	@cd backend && go run . &
	@cd frontend && npm run dev &
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:5173"
	@wait

dev-backend:
	@echo "Starting backend server..."
	@cd backend && go run .

dev-frontend:
	@echo "Starting frontend dev server..."
	@cd frontend && npm run dev

# Testing
test: test-backend test-frontend

test-backend:
	@echo "Running backend tests..."
	@cd backend && go test ./... -v

test-frontend:
	@echo "Running frontend tests..."
	@cd frontend && npm test

test-e2e:
	@echo "Running E2E tests..."
	@cd frontend && npx playwright test

# Building
build: build-backend build-frontend

build-backend:
	@echo "Building backend..."
	@cd backend && go build -o bin/server .

build-frontend:
	@echo "Building frontend..."
	@cd frontend && npm run build

# Emulator Setup
setup-emulator:
	@echo "Setting up EmulatorJS..."
	@echo "Downloading EmulatorJS core files..."
	@mkdir -p frontend/public/emulator
	@echo "Note: Please download EmulatorJS manually from https://www.emulatorjs.org/"
	@echo "Extract files to: frontend/public/emulator/"
	@echo "Required files: loader.js, cores/*.wasm"
	@# Alternative: Use npm package if available
	@# @cd frontend && npm install @emulatorjs/core && cp -r node_modules/@emulatorjs/core/data/* public/emulator/

# Docker
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

# Cleanup
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf backend/bin
	@rm -rf frontend/dist
	@rm -f backend/data/*.db
	@rm -rf backend/data/
	@echo "Clean complete."
