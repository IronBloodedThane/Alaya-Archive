# Alaya Archive Development Notes

Media collection catalog application for tracking manga, movies, anime, books, games, and other media. Built as a learning project with a focus on modern web development, mobile-friendly PWA design, and cloud deployment. Supports multiple users with social features (friend requests, follows, activity feed).

## Architecture

- Go API (chi router) is the sole backend service — handles all routes
- SQLite database with WAL mode for concurrent reads
- Pure Go SQLite driver (modernc.org/sqlite) — no CGO required
- Migrations embedded in Go code (no separate migration tool needed)
- Frontend is a React 19 PWA installable on mobile devices

## Overarching Software Concerns
- S.O.L.I.D. and D.R.Y. principles are important and should be adheared too.  Ask if they are going to be violated.  
- Automated testing is of utmost importance.  Protecting us from accidentally making breaking changes is key.
- All new feature should start with writing an automated test to cover the new functionality.  
- 80% or better unit test coverage
- Automated End-to-end testing should cover critical functionality flows

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

## Mobile (mobile-flutter/)

- Flutter app (Android + iOS), bundle id `com.dewees.alaya_archive`
- Targets the same Go API as the React frontend
- Initial focus is mobile-only features (barcode scanning, camera lookup)
- Key deps: `mobile_scanner` (barcodes), `dio` (HTTP)
- Run locally: `cd mobile-flutter && flutter pub get && flutter run`
- Test: `cd mobile-flutter && flutter test`

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
- SQLite persistence on Cloud Run: GCS bucket mounted as a volume at `/data`
  via the `--add-volume type=cloud-storage` flag in the deploy workflow
- Backups: GCS object versioning on the live bucket plus a daily external
  snapshot via `.github/workflows/backup.yml` to a separate backups bucket.
  Setup and restore steps in `BACKUP.md`.

## Local development with docker-compose

`docker-compose.yml` brings up the full stack:

- `api` (Go backend) on `:8080`, db file in a named volume
- `frontend` (Vite dev server) on `:5173`, hot reload via bind mount
- Frontend's `/api/*` proxy points at the `api` service via the
  `VITE_API_PROXY_TARGET` env var

`docker compose up --build`. Run without docker by starting each part
manually if you prefer (see Go API and Frontend sections above).

## Quality

- All code should have minimum 80% unit test coverage
- All APIs should have automated contract testing
- The UI should have automated testing around key functionality and workflows

## Testing

### Backend (Go)

- Integration tests in `backend-go/internal/handler/*_test.go`
- In-memory SQLite per test via `newTestEnv(t)`; pragma `foreign_keys=ON` is set explicitly so cascade tests exercise real behavior
- Email sends are captured by a `fakeMailer` implementing `email.Sender` — no network calls
- Run: `cd backend-go && go test ./...`

### End-to-end (Playwright)

- Tests in `frontend/e2e/*.spec.js`, fixtures in `frontend/e2e/fixtures.js`
- `createUser` fixture registers a user via the API and auto-deletes it in teardown; emails are namespaced `e2e-<uuid>@test.alaya-archive.com` so stray data is easy to spot
- For token-gated flows (verify-email, reset-password), tests mint their own JWTs with the shared `E2E_SECRET_KEY` rather than intercept real email — backend is configured with an empty `EMAIL_API_KEY` so no Resend calls are made during tests
- Playwright's `webServer` config auto-starts the Go backend (on a separate `e2e.db`) and the Vite dev server
- Run: `cd frontend && npm run test:e2e`
