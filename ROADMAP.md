# Alaya Archive Roadmap

## What We Have Now

- **Auth**: Register, login, JWT tokens, password reset/change, email verification (emails not yet sending)
- **Media Catalog**: Full CRUD — add/edit/delete media with ratings, progress tracking, tags, cover images, status
- **Search & Filter**: By type, status, tags, text search with pagination
- **User Profiles**: Display name, bio, avatar upload, public profiles
- **Social**: Follow/unfollow, friend requests, friends list, activity feed
- **CI/CD**: Tests + lint on PR, deploy to Cloud Run + Firebase on merge (needs verification)
- **Docker**: Local dev with docker-compose

---

## Phase 1: Polish & Fix (Current)

Get the existing features solid before adding new ones.

- [ ] Fix email sending (SendGrid integration for verification & password reset)
- [x] Implement public collections endpoint
- [ ] Bug fixes and UI polish
- [ ] Mobile PWA testing and fixes
- [ ] Verify CI/CD pipeline deploys correctly end-to-end
- [ ] Add frontend to docker-compose for full local dev setup

## Phase 2: Better Catalog Experience

Make the core media tracking more useful and fun.

- [ ] Import from external APIs (MyAnimeList, TMDB, IGDB, OpenLibrary)
- [ ] Bulk add / quick add media
- [ ] Custom lists / shelves (e.g. "Top 10 Anime", "Watch with Dad")
- [ ] Review / notes field for each media entry
- [ ] Sort collection by date added, rating, title, status
- [ ] Media recommendations based on collection

## Phase 3: Social Features

Make it fun to use with friends.

- [ ] Activity feed improvements (likes, comments on activity)
- [ ] Share a list or collection publicly via link
- [ ] Compare collections with a friend ("we both liked...")
- [ ] User search / discovery
- [ ] Notifications (in-app and/or push)
- [ ] Profile customization (themes, favorites showcase)

## Phase 4: Infrastructure & Quality

Harden everything before going bigger.

- [ ] SQLite persistence strategy for Cloud Run (Litestream + GCS)
- [ ] Automated test coverage to 80%+ (backend and frontend)
- [ ] API contract tests
- [ ] Rate limiting and abuse prevention
- [ ] Monitoring and error tracking
- [ ] Backup and restore strategy

## Phase 5: Long-Term — Media Server

*Prerequisite: Phases 1-4 substantially complete.*

Turn Alaya Archive into a platform for sharing actual media files with friends.

- [ ] File upload and storage (local + cloud)
- [ ] Media streaming / download for shared files
- [ ] Sharing permissions (friends only, specific users, link sharing)
- [ ] Storage quotas and management
- [ ] Content organization (folders, playlists)
- [ ] Transcoding for video/audio compatibility

---

*This roadmap is a living document. Priorities can shift — update it as we go.*
