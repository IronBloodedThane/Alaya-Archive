# Alaya Archive

Media collection catalog application (manga, movies, anime, books, games, etc.)

## Architecture

- **Frontend**: React 19 + Vite + Tailwind CSS 4 (PWA-enabled)
- **Backend**: Go + chi router + SQLite (via modernc.org/sqlite, pure Go)
- **Auth**: JWT (HS256) with access/refresh tokens
- **Deployment**: Firebase Hosting (frontend) + Cloud Run (API)

## Project Structure

```
frontend/          React SPA
  src/api/         Axios API clients
  src/pages/       Route components
  src/components/  Shared UI components
  src/contexts/    Auth, Theme contexts
  src/hooks/       Custom React hooks

backend-go/        Go API
  cmd/api/         Entry point + router
  internal/
    handler/       HTTP handlers
    repository/    Database access (SQL)
    auth/          JWT token creation/validation
    middleware/    Auth, CORS middleware
    database/      SQLite connection + migrations
    config/        Environment config
```

## Development

```bash
# Backend
cd backend-go && go run ./cmd/api/

# Frontend
cd frontend && npm install && npm run dev
```

## Key Decisions

- SQLite instead of PostgreSQL for simplicity and cost (single-file DB)
- Pure Go SQLite driver (modernc.org/sqlite) - no CGO needed
- Migrations embedded in Go code (no separate migration tool)
- WAL mode enabled for better concurrent read performance
- Friend system with request/accept flow (not just follow)
