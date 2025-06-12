# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this Baseline Full-Stack Template repository.

## CRITICAL: Request Clarification Protocol

**If you cannot complete a request for ANY reason, STOP immediately and ask for clarification.**

- Don't make assumptions about unclear requirements
- Don't proceed with partial implementations  
- Don't guess what the user wants
- Simply state what you don't understand and ask for specific clarification

This prevents wasted time and ensures accurate implementation.

## Template Customization Guide

This is a baseline template meant to be customized for specific projects. When working with this template:

1. **Project Setup**: Guide users to copy `.env.template` to `.env` and update values
2. **Naming Updates**: Help update project names in `package.json`, README, and other configs
3. **Database Path**: Current default is `data/app.db` - can be customized per project
4. **Security Values**: Always remind users to generate secure secrets for production
5. **Feature Development**: The baseline includes auth, WebSocket, and basic UI - build upon these

## Common Development Commands

### Development Environment
- **Start development**: `tilt up` (starts all services with hot reloading)
- **Stop development**: `tilt down`
- **View logs**: `tilt up --stream`
- **Tilt dashboard**: http://localhost:10350

### Testing & Linting
- **Server tests**: `tilt trigger server-tests` or `cd server && go test ./...`
- **Server test coverage**: `cd server && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html`
- **Server linting**: `tilt trigger server-lint` or `cd server && golangci-lint run`
- **Client linting**: `tilt trigger client-lint` or `cd client && npm run lint:check`
- **Client tests**: `tilt trigger client-tests` or `cd client && npm run test`

### Database Operations
- **Run migrations**: `tilt trigger migrate-up`
- **Rollback migration**: `tilt trigger migrate-down`
- **Seed database**: `tilt trigger migrate-seed`
- **Valkey info**: `tilt trigger valkey-info`

### Manual Development (without Tilt)
- **Server**: `cd server && go run cmd/api/main.go`
- **Client**: `cd client && npm run dev`
- **Full stack**: `docker compose -f docker-compose.dev.yml up --build`

### Important Note: cd Command Aliasing
The `cd` command is aliased. When using bash commands, use absolute paths instead:
- Instead of: `cd server && go test ./...`
- Use: `go test /home/bobparsons/Development/bobb/baseline/server/...`
- Or use the working directory parameter in bash commands

## Architecture Overview

### High-Level Structure
Full-stack application with Go backend, SolidJS frontend, and Valkey cache:
- **Backend**: Fiber framework with SQLite + GORM, JWT auth, WebSockets
- **Frontend**: SolidJS with TypeScript, Vite, CSS Modules, Solid Query
- **Cache**: Valkey (Redis-compatible) for sessions and caching
- **Orchestration**: Docker Compose + Tilt for development

### Key Ports
- Server API: http://localhost:8280 (WebSocket: ws://localhost:8280/ws)
- Client App: http://localhost:3010
- Valkey DB: localhost:6379

### Backend Architecture (Go)
- **Dependency Injection**: App struct (`internal/app/app.go`) contains all services
- **Controllers**: Interface-based design (`internal/interfaces/`)
- **Database**: Dual database setup - SQLite (primary) + Valkey (cache)
- **Auth**: JWT tokens with bcrypt, middleware-based protection
- **WebSockets**: Manager pattern with hub for real-time communication
- **Routing**: Fiber router with middleware chain

### Frontend Architecture (SolidJS)
- **State Management**: AuthContext + Solid Query for server state
- **API Layer**: Axios with interceptors for token management (`services/api/`)
- **WebSocket**: Auto-connecting WebSocket context with auth token header
- **Routing**: @solidjs/router with protected routes
- **Styling**: SCSS with CSS Modules pattern

### Database Layer
- **Primary**: SQLite with GORM (migrations in `cmd/migration/`)
- **Cache**: Valkey client for sessions and temporary data
- **Models**: GORM models with methods (`internal/models/`)

### Authentication Flow
1. Login via `/users/login` returns JWT
2. Token stored in HTTP-only cookie and sent via `X-Auth-Token` header
3. AuthContext manages client state and API interceptors
4. WebSocket auth uses same token in connection headers
5. Middleware validates JWT on protected routes

### WebSocket Architecture
- Hub pattern managing client connections
- Auth token required in connection headers
- Real-time communication between authenticated clients
- Automatic reconnection and auth token refresh on client

## Development Notes

### Key Files to Understand
- `server/internal/app/app.go` - Main dependency injection container
- `client/src/context/AuthContext.tsx` - Auth state management
- `server/internal/routes/router.go` - API route definitions
- `client/src/services/api/api.service.ts` - API client with interceptors
- `Tiltfile` - Development environment configuration

### Environment Configuration
All environment variables in `.env` at project root, shared between services.

### Testing Strategy
- Go tests: Standard `go test` with table-driven tests
- Frontend: ESLint for linting, test framework ready for setup
- Manual testing via Tilt dashboard utilities

### Database Migrations
- Migrations in `server/cmd/migration/migrations/`
- Use `tilt trigger migrate-up` for applying migrations
- Seed data available via `tilt trigger migrate-seed`