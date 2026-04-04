# Alaya Archive Development Notes

Media collection catalog application for tracking manga, movies, anime, books, games, and other media. Built as a learning project with a focus on modern web development, mobile-friendly PWA design, and cloud deployment. Supports multiple users with social features (friend requests, follows, activity feed).

## Architecture

- Go API (chi router) is the sole backend service — handles all routes
- SQLite database with WAL mode for concurrent reads
- Pure Go SQLite driver (modernc.org/sqlite) — no CGO required
- Migrations embedded in Go code (no separate migration tool needed)
- Frontend is a React 19 PWA installable on mobile devices

## Go API (backend-go/)

- Router: chi, handlers in `internal/handler/`, one file per resource
- Each handler follows: struct with repos+config → handler methods
- Auth: `middleware.RequireAuth(secretKey)` injects user ID into context
- `middleware.OptionalAuth(secretKey)` for public routes that optionally use auth
- SQL via modernc.org/sqlite (no ORM), raw SQL queries in repository layer
- **Repository pattern**: SQL queries live in `internal/repository/`, handlers call repository methods for all database access
- ID generation: crypto/rand hex-encoded 16-byte strings
- Run locally: `cd backend-go && go run ./cmd/api/`

## Frontend (frontend/)

- React 19 + Vite + Tailwind CSS 4
- State: React Context (Auth, Theme) + TanStack React Query for server state
- HTTP client: Axios with JWT auto-refresh interceptor
- Routing: React Router v6 with ProtectedRoute/GuestRoute wrappers
- PWA: vite-plugin-pwa with Workbox caching (NetworkFirst for API, 7-day cache)
- Dev server proxies `/api` to localhost:8080
- Run locally: `cd frontend && npm install && npm run dev`

## Database

- SQLite with WAL journal mode, foreign keys enforced, 5s busy timeout
- Single connection (MaxOpenConns=1) for write safety
- Migrations tracked in `schema_migrations` table, applied on startup
- Data path configurable via `DATABASE_PATH` env var (default: `./data/alaya-archive.db`)

## Media Types

manga, anime, movie, book, game, tv_show, music, other

## Media Status Values

planned, in_progress, completed, on_hold, dropped

## Deployment

- Frontend: Firebase Hosting (CDN, SPA rewrites, Cloud Run API proxy)
- Backend: Google Cloud Run (scale-to-zero)
- CI/CD: GitHub Actions (test + lint on PR, deploy on merge to main)
- SQLite persistence on Cloud Run: TBD (Litestream + GCS recommended)

## Quality

- All code should have minimum 80% unit test coverage
- All APIs should have automated contract testing
- The UI should have automated testing around key functionality and workflows
