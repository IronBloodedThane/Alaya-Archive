# Alaya Archive

Media collection catalog application. Track your manga volumes, movies, anime series, books, games, and more. Rate and tag items, track reading/watching progress, connect with friends, and share your collection. Free, privacy-focused, and installable as a PWA on mobile devices.

## Features

### Media Catalog
- Support for 8 media types: manga, anime, movies, books, games, TV shows, music, and other
- Track status: planned, in progress, completed, on hold, dropped
- Rate items on a 1-10 scale
- Custom tags for organizing your collection
- Original title field for Japanese/foreign language titles
- Cover images via URL
- Per-type progress tracking:
  - Manga/books: volumes owned, chapters read
  - Anime/TV: episodes watched
- Notes field for personal thoughts
- Public/private visibility per item
- Search and filter by type, status, tags

### Social
- Friend request system with accept/reject flow
- Follow other users
- Activity feed showing friends' and followed users' activity
- Public user profiles with visible collection

### Account & Privacy
- Email/username + password registration
- JWT authentication with auto-refreshing tokens
- Email verification
- Password reset via email
- Full account deletion with cascade
- No tracking, no ads, no data selling

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **API** | Go 1.23, chi router, raw SQL |
| **Database** | SQLite (WAL mode, pure Go driver) |
| **Auth** | JWT (access + refresh tokens), bcrypt |
| **Frontend** | React 19, React Router 6, Tailwind CSS 4, Vite 6 |
| **State** | TanStack React Query, React Context |
| **PWA** | vite-plugin-pwa, Workbox |
| **HTTP Client** | Axios |
| **CI/CD** | GitHub Actions |
| **Hosting** | Firebase Hosting (frontend), Google Cloud Run (API) |

## Project Structure

```
alaya-archive/
├── backend-go/
│   ├── cmd/api/
│   │   ├── main.go                # Server entry point, DB init, migrations
│   │   └── router.go              # Route registration, middleware, handler wiring
│   ├── internal/
│   │   ├── auth/jwt.go            # JWT creation/validation (access, refresh, verify, reset)
│   │   ├── config/config.go       # Environment variable configuration
│   │   ├── database/
│   │   │   ├── sqlite.go          # SQLite connection (WAL, foreign keys, busy timeout)
│   │   │   └── migrations.go      # Embedded schema migrations (6 versions)
│   │   ├── handler/
│   │   │   ├── auth.go            # Register, login, refresh, verify, password reset, delete
│   │   │   ├── user.go            # Profile CRUD, avatar upload, public profiles
│   │   │   ├── media.go           # Media CRUD, rating, tags, stats, search
│   │   │   └── social.go          # Follow/unfollow, friend requests, feed
│   │   ├── middleware/
│   │   │   ├── auth.go            # RequireAuth, OptionalAuth, context user ID
│   │   │   └── cors.go            # CORS with configurable origins
│   │   └── repository/
│   │       ├── user.go            # User CRUD, lookup by email/username
│   │       ├── media.go           # Media CRUD, filtered listing, stats, tags
│   │       └── social.go          # Follows, friend requests, friends, feed items
│   ├── go.mod
│   ├── Dockerfile
│   └── .dockerignore
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   │   ├── client.js          # Axios instance with JWT refresh interceptor
│   │   │   ├── auth.js            # Auth API calls
│   │   │   ├── media.js           # Media API calls
│   │   │   └── social.js          # Social/friends API calls
│   │   ├── pages/
│   │   │   ├── Landing.jsx        # Public landing page
│   │   │   ├── Login.jsx          # Sign in
│   │   │   ├── Register.jsx       # Create account
│   │   │   ├── Dashboard.jsx      # Collection stats overview
│   │   │   ├── Collection.jsx     # Browse/filter/search collection
│   │   │   ├── AddMedia.jsx       # Add or edit media item
│   │   │   ├── MediaDetail.jsx    # View media item details
│   │   │   ├── Profile.jsx        # Profile settings
│   │   │   ├── PublicProfile.jsx  # Public user profile
│   │   │   ├── Friends.jsx        # Friends list and requests
│   │   │   ├── Feed.jsx           # Activity feed
│   │   │   └── NotFound.jsx       # 404 page
│   │   ├── components/
│   │   │   └── Layout.jsx         # App shell with nav (desktop + mobile)
│   │   ├── contexts/
│   │   │   ├── AuthContext.jsx    # Auth state, login/register/logout
│   │   │   └── ThemeContext.jsx   # Dark/light theme toggle
│   │   └── hooks/
│   │       ├── useAuth.js         # Auth context hook
│   │       └── useTheme.js        # Theme context hook
│   ├── index.html
│   ├── vite.config.js             # Vite + React + Tailwind + PWA config
│   ├── firebase.json              # Firebase Hosting with Cloud Run API proxy
│   ├── .firebaserc                # Firebase project reference
│   └── package.json
├── docker-compose.yml              # Local development (API with mounted volume)
├── .env.example                    # Required environment variables
├── .gitignore
├── CLAUDE.md                       # Development notes
└── .github/workflows/
    ├── ci.yml                      # Test + lint + build on PR and push
    └── deploy.yml                  # Build + deploy to Cloud Run + Firebase on merge to main
```

## Getting Started

### Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Node.js 22+](https://nodejs.org/)
- [Docker](https://docs.docker.com/get-docker/) (optional, for containerized development)

### Local Development

1. **Clone the repo**

   ```bash
   git clone https://github.com/your-org/alaya-archive.git
   cd alaya-archive
   ```

2. **Start the backend**

   ```bash
   cd backend-go
   go run ./cmd/api/
   ```

   The API is now running at `http://localhost:8080`. The SQLite database is created automatically at `./data/alaya-archive.db` with all migrations applied on startup.

3. **Start the frontend**

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   The frontend is now running at `http://localhost:5173` with Vite's dev server proxying `/api` requests to the Go API.

4. **Run backend tests**

   ```bash
   cd backend-go
   go test ./...
   ```

### Using Docker

```bash
# Start the API in a container with persistent volume
docker compose up -d

# API available at http://localhost:8080
# Then start the frontend separately with npm run dev
```

### Database Migrations

Migrations are embedded in Go code and run automatically on startup. To add a new migration:

1. Add a new entry to the `migrations` slice in `backend-go/internal/database/migrations.go`
2. Increment the version number
3. Restart the API — the new migration applies automatically

## Production Deployment (Google Cloud)

Frontend is served by Firebase Hosting; the API runs on Cloud Run with scale-to-zero. The SQLite database file lives in a GCS bucket mounted into the Cloud Run container as a FUSE volume, which means state survives restarts without a separate database service.

### Architecture

- **Firebase Hosting** serves the built frontend and rewrites `/api/**` to the Cloud Run service, so the browser only ever talks to the Firebase origin (no CORS headaches).
- **Cloud Run** runs the Go API as a gen2 service, pinned to `--max-instances=1` (SQLite is single-writer — multiple instances would corrupt the DB).
- **GCS bucket** (`alaya-archive-prod-db`) is mounted at `/data` via Cloud Run's Cloud Storage volume support. `DATABASE_PATH=/data/alaya-archive.db`.
- **Journal mode**: `DB_JOURNAL_MODE=DELETE` in production (GCS FUSE has no POSIX file locks, so WAL is unsafe). Local dev still defaults to WAL.
- **Artifact Registry** holds the Docker images.

### One-time GCP setup

The setup commands are documented in this repo's git history under the "Setup GCP deployment" conversation — the short version:

1. Create the project, enable `run`, `artifactregistry`, `iam`, `cloudbuild`, and `storage` APIs, set region.
2. Create Artifact Registry repo `alaya-archive` in `us-central1`.
3. Create the DB bucket with uniform bucket-level access.
4. Create two service accounts:
   - `ci-deployer@$PROJECT.iam.gserviceaccount.com` — used by GitHub Actions. Needs `roles/run.admin`, `roles/artifactregistry.writer`, `roles/iam.serviceAccountUser`, `roles/firebasehosting.admin`.
   - `api-runtime@$PROJECT.iam.gserviceaccount.com` — used by the Cloud Run service at runtime. Needs `roles/storage.objectAdmin` **scoped to the DB bucket** (not project-wide).
5. Download a JSON key for `ci-deployer` and add it as the `GCP_SA_KEY` GitHub secret.
6. Link Firebase to the GCP project (`firebase projects:addfirebase $PROJECT`), then `firebase use --add` inside `frontend/` with alias `default`.

### GitHub secrets

Add these in **Settings → Secrets and variables → Actions**:

| Secret | Value / Example |
|--------|-----------------|
| `GCP_SA_KEY` | Full JSON contents of the `ci-deployer` key file |
| `GCP_PROJECT_ID` | `alaya-archive-prod` |
| `GCP_REGION` | `us-central1` |
| `DB_BUCKET` | `alaya-archive-prod-db` |
| `SECRET_KEY` | Output of `openssl rand -base64 48` |
| `CORS_ORIGINS` | `https://alaya-archive-prod.web.app` |
| `FRONTEND_URL` | `https://alaya-archive-prod.web.app` |
| `SENDGRID_API_KEY` | `disabled` until email is wired up |
| `SENDGRID_FROM_EMAIL` | Placeholder until email is wired up |

### Deploying

Fully automated. Pushes to `main` trigger `.github/workflows/deploy.yml`, which:

1. Builds the API Docker image and pushes it to Artifact Registry (tagged with the commit SHA).
2. Deploys to Cloud Run with the GCS volume mount, `api-runtime` SA, gen2 execution environment, and all env vars from secrets.
3. Builds the frontend and deploys to Firebase Hosting.

After a successful deploy, the app is live at `https://<project-id>.web.app`.

### Operational notes

- **Never increase `--max-instances` above 1** for Cloud Run while SQLite remains the database. Concurrent writers on GCS FUSE will corrupt the file.
- **Backups**: GCS bucket versioning is worth enabling — accidental deletes of the DB file are otherwise unrecoverable.
- **Cold starts**: first request after idle pays the Cloud Run cold start + GCS FUSE mount time (a few seconds). Acceptable for a personal catalog; consider `--min-instances=1` if that becomes annoying (costs ~$5/mo).
- **Migrations**: run automatically on startup; there's no separate deploy step.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API server port |
| `DATABASE_PATH` | `./data/alaya-archive.db` | Path to SQLite database file |
| `DB_JOURNAL_MODE` | `WAL` | SQLite journal mode (set to `DELETE` on GCS FUSE) |
| `SECRET_KEY` | `change-me-in-production` | JWT signing key — **change in production** |
| `CORS_ORIGINS` | `http://localhost:5173` | Comma-separated allowed origins |
| `FRONTEND_URL` | `http://localhost:5173` | Frontend URL for email links |
| `ACCESS_TOKEN_EXPIRE_MINUTES` | `15` | Access token TTL |
| `REFRESH_TOKEN_EXPIRE_DAYS` | `30` | Refresh token TTL |
| `SENDGRID_API_KEY` | — | SendGrid API key (optional) |
| `SENDGRID_FROM_EMAIL` | — | Sender email address (optional) |
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |

## API Endpoints

### Auth
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login (returns access + refresh tokens) |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/verify-email` | Verify email with token |
| POST | `/api/v1/auth/forgot-password` | Request password reset email |
| POST | `/api/v1/auth/reset-password` | Reset password with token |
| POST | `/api/v1/auth/change-password` | Change password (auth required) |
| POST | `/api/v1/auth/delete-account` | Delete account (auth required) |
| POST | `/api/v1/auth/resend-verification` | Resend verification email (auth required) |

### Users
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/users/me` | Get current user profile |
| PATCH | `/api/v1/users/me` | Update profile |
| POST | `/api/v1/users/me/avatar` | Upload avatar (multipart) |
| DELETE | `/api/v1/users/me/avatar` | Delete avatar |
| GET | `/api/v1/users/{username}` | Get public profile |

### Media
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/media` | List media (filterable by type, status, search, tag) |
| POST | `/api/v1/media` | Add media to collection |
| GET | `/api/v1/media/stats` | Collection statistics |
| GET | `/api/v1/media/search` | Search media |
| GET | `/api/v1/media/{id}` | Get media details |
| PATCH | `/api/v1/media/{id}` | Update media |
| DELETE | `/api/v1/media/{id}` | Delete media |
| POST | `/api/v1/media/{id}/rating` | Rate media (1-10) |
| POST | `/api/v1/media/{id}/tags` | Set tags |

### Social
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/social/follow/{userId}` | Follow user |
| DELETE | `/api/v1/social/follow/{userId}` | Unfollow user |
| GET | `/api/v1/social/followers` | List followers |
| GET | `/api/v1/social/following` | List following |
| GET | `/api/v1/social/feed` | Activity feed |

### Friends
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/friends/request/{userId}` | Send friend request |
| POST | `/api/v1/friends/accept/{requestId}` | Accept friend request |
| POST | `/api/v1/friends/reject/{requestId}` | Reject friend request |
| GET | `/api/v1/friends` | List friends |
| GET | `/api/v1/friends/requests` | List pending friend requests |
| DELETE | `/api/v1/friends/{friendId}` | Remove friend |

## License

This project is licensed under the Apache License 2.0 — see [LICENSE](LICENSE) for details.
