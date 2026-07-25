# Alaya Archive — Master Roadmap & System Governance

## Project Leadership & Sign-Off Matrix
* **Project Owner / Lead Architect (James):** Final authority on product scope, UI/UX aesthetics, core feature selection, architectural design, and code merges into `main`. **Project Owner sign-off is required for any public release.**
* **Chief Technical Advisor / Consulting CTO (Tim):** Advisory oversight on backend stability, security baselines, and technical review.

## System Architecture Principle: Zero-Cloud Media Engine
* **Central Cloud Scope (`SaaS Tier`):** Account authentication, user profiles, social graph (friend lists), and ephemeral WebRTC signaling brokers. Central servers hold zero media files, local directory paths, or stream buffers.
* **Local Node Scope (`Self-Hosted Tier`):** Media storage, library indexing, local SQLite database, local metadata, and hardware transcoding.
* **Access Rule:** **No Server, No Stream.** A user account must be cryptographically bound to an active, running local server node to unlock media playback and watch party features.

---

## Phase 1: Infrastructure, Ticketing & System Governance
> **Goal:** Build a rock-solid, professional administrative foundation and communication pipeline before writing new consumer features.

* [ ] **1.1 Official Domain Email Pipeline:**
  * Configure SPF/DKIM/DMARC for official domain communications (`@alaya-archive.com`).
  * Establish dedicated routes for verified staff (`staff@`), platform updates (`newsletter@`), and customer support (`support@`).
* [ ] **1.2 Inbound Support Ticketing Engine:**
  * Build an internal administrative dashboard that ingests emails from `support@` and converts them into tracked tickets (Open / In Progress / Resolved).
* [ ] **1.3 Automated Client Update Pipeline:**
  * Implement an opt-in newsletter system to dispatch roadmap updates and platform announcements to registered users.
* [ ] **1.4 Local Dev & Infrastructure Parity:**
  * Standardize local development with a unified `docker-compose.yml` for backend and frontend environments.
  * Verify CI/CD automation executes linters, unit tests, and staging deployments on PR merge.

---

## Phase 2: Brand Identity, Artist Onboarding & UI/UX Design
> **Goal:** Eliminate developer-aesthetic placeholders and establish the true visual identity of Alaya Archive across web and mobile.

> **RELEASE GATE:** *No public release is permitted until visual branding and core UI components receive explicit Project Owner sign-off.*

* [ ] **2.1 Visual Identity & Asset Commissioning:**
  * Hire artists/designers for custom iconography, logo suites, and core branding assets.
  * Lock down a standardized Design System (color palette, typography, design tokens) for web and mobile components.
* [ ] **2.2 Public Vision & Goals Page:**
  * Create an accessible, public-facing landing page outlining the philosophy, roadmap, and privacy-first architecture of the Alaya Archive.
* [ ] **2.3 Profile Creation & Management Overhaul:**
  * Redesign user profiles with custom avatars, user bios, public showcase options, and cascading account deletion tools.

---

## Phase 3: Mobile PWA/GUI Refinement & Server Binding
> **Goal:** Ensure complete cross-platform library access and mobile UI polish while enforcing server-node validation.

* [ ] **3.1 Mobile UI/GUI Refinement:**
  * Overhaul the mobile PWA/app interface to match the Phase 2 design system with touch-optimized controls.
  * Verify PWA installation and UI rendering on iOS and Android.
* [ ] **3.2 Server Node Binding & Device Pairing:**
  * Enforce the "Must Have Server" check during authentication before granting playback UI access.
  * Implement QR-code pairing to establish a secure client-certificate link between mobile devices and a user's home server.
* [ ] **3.3 Public Launch Review Gate:**
  * Conduct a full product review. Opening the app to the public requires formal Project Owner sign-off.

---

## Phase 4: Local Storage, File Indexing & PWA / WebRTC Engine
> **Goal:** Build the local media indexing backend and private peer-to-peer streaming engine.

* [ ] **4.1 Local Media Parsing Engine (Golang):**
  * Build a lightweight local worker to extract metadata, codecs, and container headers from local MP4, MKV, and audio files.
  * Store index data locally in a user-hosted SQLite database.
* [ ] **4.2 "Glass Window" Direct-Peer Streaming (WebRTC):**
  * Implement WebRTC encrypted P2P tunnels for direct, host-to-guest video frame rendering.
  * Ensure zero disk persistence on guest clients (memory-only frame rendering to prevent binary extraction).
* [ ] **4.3 Host Control Matrix:**
  * Lock playback controls (Play/Pause/Seek, audio track selection) exclusively to the media owner's server session.

---

## Phase 5: Watch Parties & Synchronized Rooms
> **Goal:** Enable private, password-protected social viewing rooms between friends.

* [ ] **5.1 Synchronized WebSocket Control Broker:**
  * Build a low-latency signaling engine to synchronize playback timestamps across room participants (<200ms drift target).
  * Route timecode packets only; central cloud servers never touch video streams.
* [ ] **5.2 Password-Protected Room Invites:**
  * Allow room hosts to set access passwords and issue ephemeral, expiring room tokens to friends.
  * Enforce hard room capacity limits (e.g., maximum 5 connected peers) to maintain private social viewing compliance.
* [ ] **5.3 Federated Room Cataloging:**
  * Allow room participants with linked home servers to publish temporary session metadata showing their available titles (e.g., *"James offers: Gundam Wing; Mark offers: Die Hard 2"*).
