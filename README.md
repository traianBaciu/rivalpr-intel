# RivalPR Intel

A PR Agency CRM and Outreach Tracker. Manage clients, track media contacts, score journalist relationships, and generate personalised pitch emails using Google Gemini AI — all scoped to your account.

---

## Quick Start

### Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) running
- A `.env` file in the project root (see [Environment Variables](#environment-variables))

### 1. Create your `.env`

```bash
cp .env.example .env
```

Open `.env` and set `JWT_SECRET` to any random string of at least 32 characters. Everything else works with the defaults.

### 2. Start the stack

```bash
docker compose up --build
```

| Service  | URL                   |
| -------- | --------------------- |
| Backend  | http://localhost:8080 |
| Database | localhost:**5433**    |

The backend auto-creates all database tables on first boot. Verify it's running:

```bash
curl http://localhost:8080/health
# {"status":"ok","timestamp":"..."}
```

### 3. Stop the stack

```bash
docker compose down
```

### Everyday workflow

```bash
docker compose up       # start (no rebuild)
docker compose up -d    # start in background
docker compose logs -f  # tail logs
```

The backend uses `air` for hot-reload — saving any `.go` file automatically recompiles and restarts the server inside the container.

---

## Environment Variables

| Variable              | Required | Default                 | Description                                                                                     |
| --------------------- | -------- | ----------------------- | ----------------------------------------------------------------------------------------------- |
| `DATABASE_URL`        | Yes      | (see .env.example)      | PostgreSQL connection string                                                                    |
| `JWT_SECRET`          | Yes      | —                       | Token signing key — **min 32 characters**                                                       |
| `POSTGRES_USER`       | Yes      | `rivalpr`               | Database user (used by postgres container)                                                      |
| `POSTGRES_PASSWORD`   | Yes      | `rivalpr_secret`        | Database password                                                                               |
| `POSTGRES_DB`         | Yes      | `rivalpr`               | Database name                                                                                   |
| `BCRYPT_COST`         | No       | `12`                    | bcrypt work factor (min 10)                                                                     |
| `AI_RATE_LIMIT_RPM`   | No       | `20`                    | AI generation requests per user per minute                                                      |
| `GEMINI_API_KEY`      | No       | —                       | Required for real AI generation; get free at [aistudio.google.com](https://aistudio.google.com) |
| `NEXT_PUBLIC_API_URL` | No       | `http://localhost:8080` | Backend URL for the frontend                                                                    |

---

## API

Import the Postman collection for a ready-to-use test suite:

- `api-collection/RivalPR-Intel.postman_collection.json`
- `api-collection/RivalPR-Intel.postman_environment.json`

The collection auto-saves the JWT token and all resource IDs — run requests top to bottom and everything chains together.

### Endpoints

| Method                | Path                               | Description                          |
| --------------------- | ---------------------------------- | ------------------------------------ |
| `POST`                | `/auth/register`                   | Create account                       |
| `POST`                | `/auth/login`                      | Get JWT token                        |
| `GET/POST/PUT/DELETE` | `/api/clients`                     | Client management                    |
| `GET/POST/PUT/DELETE` | `/api/outlets`                     | Media outlet management              |
| `GET/POST/PUT/DELETE` | `/api/journalists`                 | Journalist database (paginated)      |
| `GET`                 | `/api/journalists/agency`          | Shared agency journalist pool        |
| `GET/POST/PUT/DELETE` | `/api/crm/relationships`           | Per-journalist relationship scores   |
| `GET/POST/PUT/DELETE` | `/api/campaigns`                   | Campaign management                  |
| `GET/POST/PUT/DELETE` | `/api/prompt-templates`            | Reusable AI generation presets       |
| `GET/POST/PUT/DELETE` | `/api/pitches`                     | Pitch management                     |
| `GET`                 | `/api/pitches/:id/versions`        | List AI-generated versions           |
| `GET`                 | `/api/pitches/:id/versions/:verId` | Get specific version                 |
| `POST`                | `/api/pitches/:id/generate`        | Generate a new pitch version (async) |
| `GET`                 | `/health`                          | Health check                         |

All `/api/*` routes require `Authorization: Bearer <token>`.

---

## User Scenarios

### Epic 1 — Authentication

A PR professional registers with their email, logs in, and receives a JWT. All their data — campaigns, pitches, relationships — is scoped to their account. No one else can see or modify their records.

```
Register → Login → receive token → use for all subsequent requests
```

---

### Epic 2 — Client & Campaign Management

The user creates a client (e.g. "Acme Corp") representing a brand they manage. Under that client, they create a campaign with a title and a press release text. The campaign becomes the top-level context for all outreach activity.

```
Create Client (Acme Corp, industry: Technology)
  └── Create Campaign (title: "Q2 AI Platform Launch", press_release_text: "...")
```

---

### Epic 3 — Journalist CRM

The user builds a personal media contact database. They add journalists linked to outlets (TechCrunch, Forbes, etc.) and maintain a private `relationship_score` (1–10) and notes for each — stored in `crm_relationships`, invisible to other users.

> _"I've worked with Sarah at TechCrunch three times — let me score her as 8/10 and note that she responds best to exclusive data angles."_

```
Create Outlet (TechCrunch)
  └── Create Journalist (Sarah Chen, niche: Enterprise SaaS)
        └── Create CRM Relationship (score: 8, notes: "warm contact, prefers data-first pitches")
```

---

### Epic 4 — AI-Powered Pitch Generation

The user selects a campaign and a journalist, then hits **Generate**. The system uses the campaign's press release, the journalist's niche and outlet name, and their CRM score as context — calls the AI asynchronously, and stores the result as a new `pitch_version`. Each regeneration increments the version number.

**Generation parameters:**

| Parameter           | Options                                                                  |
| ------------------- | ------------------------------------------------------------------------ |
| Tone                | `formal` · `conversational` · `urgent` · `enthusiastic`                  |
| Length              | `concise` · `standard` · `detailed`                                      |
| Angle               | `news_hook` · `exclusive` · `follow_up` · `thought_leadership` · `event` |
| Custom Instructions | Free-text (e.g. "mention their recent AI article")                       |

These parameters can be saved as **Prompt Templates** for reuse across pitches.

> _"Generate a pitch for Sarah at TechCrunch for the Acme AI launch with a conversational tone and exclusive angle, then refine as v2."_

```
Create Pitch (client: Acme, journalist: Sarah Chen, campaign: Q2 Launch)
  └── POST /pitches/:id/generate  →  202 Accepted (async worker)
  └── GET  /pitches/:id/versions  →  poll until v1 appears
  └── POST /pitches/:id/generate  →  refine with different params → v2
  └── GET  /pitches/:id/versions  →  compare v1 and v2 side-by-side
```

The frontend polls `GET /pitches/:id/versions` for up to 60 seconds after generation is requested. A side-by-side comparison view helps users pick the best version.

> **Note:** AI generation returns a mock template response when no key is configured. Real Gemini integration requires `GEMINI_API_KEY` (free tier available at [aistudio.google.com](https://aistudio.google.com)).

---

### Epic 5 — Pitch Tracking & Status

After sending a pitch, the user updates its status to reflect real-world progress. Valid states: `draft → sent → opened → replied`. The pitch list can be filtered by status or campaign.

```
PUT /api/pitches/:id  { "status": "sent" }
GET /api/pitches?status=replied       ← all journalists who responded
GET /api/pitches?campaign_id=<id>     ← all pitches for a specific campaign
```

---

### Epic 6 — Outreach Dashboard

The Next.js frontend provides a full-featured dashboard and management UI:

- **Dashboard Home** — Summary cards (total Clients, Campaigns, Pitches) with pitch status breakdown (draft / sent / opened / replied) and quick-action links.
- **Clients** — Table view with CRUD; create/edit via modal dialogs.
- **Campaigns** — Campaign list tied to clients with press release preview; full CRUD.
- **Journalists** — Tabbed interface:
  - _Journalists tab:_ paginated list, import from shared agency pool, CRM relationship scoring (1–10), private notes.
  - _Outlets tab:_ CRUD for media publications.
- **Pitches** — Filter by status/campaign, create pitches, trigger AI generation with configurable parameters, view version history, compare versions side-by-side, update status.
- **Prompt Templates** — Save/load generation parameter presets for reuse.

---

## Seed Data

Populate the database with realistic demo data for development and testing:

```bash
# Seed (5 users, 30 outlets, 200 journalists, 20 clients, 20 campaigns, 1000+ pitches)
docker compose exec backend go run ./cmd/seed --action=seed

# Clear all data (preserves schema)
docker compose exec backend go run ./cmd/seed --action=clear
```

**Demo credentials** (password for all: `Demo1234!`):

| Email             | Name          |
| ----------------- | ------------- |
| alice@rivalpr.com | Alice Morgan  |
| bob@rivalpr.com   | Bob Singh     |
| carol@rivalpr.com | Carol Tanaka  |
| dave@rivalpr.com  | Dave Okoye    |
| eve@rivalpr.com   | Eve Lindström |

The seed is deterministic (fixed random seed) — running it multiple times is safe (uses `FirstOrCreate`).

---

## Project Structure

```
.
├── backend/
│   ├── cmd/
│   │   ├── server/main.go          # entrypoint: DB, router, middleware, routes
│   │   └── seed/main.go            # database seeder & clearer
│   ├── internal/
│   │   ├── config/config.go        # env var loader
│   │   ├── models/                 # 9 GORM models
│   │   ├── handlers/               # HTTP handlers — one file per resource
│   │   ├── middleware/             # JWT auth, CORS, rate limiting
│   │   ├── services/               # auth (register/login/JWT), gemini (AI client)
│   │   └── worker/worker.go        # async AI generation goroutine pool
│   ├── Dockerfile
│   └── .air.toml                   # hot-reload config
├── frontend/
│   ├── src/
│   │   ├── app/                    # Next.js App Router pages
│   │   │   ├── (dashboard)/        # Protected: home, clients, campaigns, journalists, pitches
│   │   │   ├── login/              # Public: login page
│   │   │   └── register/           # Public: registration page
│   │   ├── components/             # Sidebar, shadcn/ui primitives
│   │   └── lib/                    # API client, auth context, types, utils
│   └── Dockerfile
├── api-collection/                 # Postman collection + environment
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Tech Stack

| Layer     | Technology                                      |
| --------- | ----------------------------------------------- |
| Backend   | Go 1.26 · Gin · GORM                            |
| Database  | PostgreSQL 16 (alpine)                          |
| Auth      | JWT (HS256) · bcrypt cost ≥ 12                  |
| AI        | Google Gemini 1.5 Flash                         |
| Frontend  | Next.js (App Router) · shadcn/ui · Tailwind CSS |
| Container | Docker · docker-compose                         |

---

## Connect to the Database

Using DBeaver, TablePlus, or `psql`:

```
Host:     localhost
Port:     5433
User:     rivalpr
Password: rivalpr_secret
Database: rivalpr
```

```bash
psql postgres://rivalpr:rivalpr_secret@localhost:5433/rivalpr
```

---

<details>
<summary><strong>Architecture & Design Decisions</strong></summary>

## System Architecture

---

## 2. System Architecture

The application follows a standard modern **decoupled architecture**, containerized via Docker for seamless deployment.

```
┌──────────────────────────────────────────────────────────┐
│                     CLIENT BROWSER                        │
└───────────────────────┬──────────────────────────────────┘
                        │ HTTPS
┌───────────────────────▼──────────────────────────────────┐
│              FRONTEND  (Container: frontend)              │
│         Next.js App Router · Tailwind CSS · shadcn/ui     │
│   Handles UI, state management, and user interactions     │
└───────────────────────┬──────────────────────────────────┘
                        │ REST / JSON
┌───────────────────────▼──────────────────────────────────┐
│               BACKEND  (Container: backend)               │
│            Go (Golang) · Gin HTTP Framework               │
│  Routing · JWT Auth · Validation · Rate-Limiting · GORM  │
│                        │                                  │
│         ┌──────────────▼──────────────┐                  │
│         │   Async Worker (goroutine)  │                   │
│         │   + Job Channel             │                   │
│         │          │                  │                   │
│         │  ┌───────▼────────┐        │                   │
│         │  │  Google Gemini  │        │                   │
│         │  │ gemini-1.5-flash         │                   │
│         │  └────────────────┘        │                   │
│         └────────────────────────────┘                   │
└──────────┬────────────────────────────────────────────────┘
           │ SQL (GORM)
┌──────────▼───────────────────────────────────────────────┐
│                DATABASE  (Container: db)                  │
│              PostgreSQL (alpine image)                    │
│              Persistent Docker Volume                     │
└──────────────────────────────────────────────────────────┘
```

---

## 3. Technical Stack

| Layer         | Technology               | Version / Notes                    |
| ------------- | ------------------------ | ---------------------------------- |
| Frontend      | Next.js (App Router)     | Latest stable                      |
| UI Components | shadcn/ui + Tailwind CSS | Utility-first styling              |
| Backend       | Go (Golang) + Gin        | Stateless REST API                 |
| ORM           | GORM                     | PostgreSQL driver                  |
| Database      | PostgreSQL               | Alpine image, persistent volume    |
| AI            | Google Gemini API        | gemini-1.5-flash                   |
| Auth          | JWT                      | Gin middleware, bcrypt (cost ≥ 12) |
| Container     | Docker + docker-compose  | 3 isolated containers              |

---

## 4. Technical Tradeoffs & Decisions

### Tradeoff 1 — Go (Golang) vs. Node.js / Python for Backend

**Decision: Go.**

Go serves as a viable Proof of Concept for the company's planned V3 backend rewrite. Building in Go here lets the team evaluate concurrency performance, deployment footprint (single compiled binary inside Docker), and developer ergonomics before committing at the enterprise level.

### Tradeoff 2 — Relational (PostgreSQL) vs. NoSQL (MongoDB)

**Decision: PostgreSQL.**

The PR domain is inherently relational. A `Pitch` requires strict foreign-key constraints to a `Client`, an optional `Campaign`, and a `Journalist`. PostgreSQL guarantees referential integrity and prevents the data duplication that would emerge in a document-based store.

### Tradeoff 3 — Next.js App Router + shadcn/ui

**Decision: Adopted App Router and component libraries.**

For a solo developer, this combination maximises frontend iteration speed — freeing time to focus on backend architecture, AI prompt engineering, and database schema validation rather than UI primitives.

---

## 5. Database Schema (ERD)

### Entity Overview

| Category          | Tables                                       | Purpose                                 |
| ----------------- | -------------------------------------------- | --------------------------------------- |
| **Core**          | `users`, `clients`, `journalists`, `outlets` | Foundation — agency-wide shared data    |
| **Workflow**      | `campaigns`, `crm_relationships`             | Press releases & personalised scoring   |
| **Action**        | `pitches`, `pitch_versions`                  | AI-generated outreach + version history |
| **Configuration** | `prompt_templates`                           | Reusable AI generation presets          |

### Improvements applied to v2.0

- **`outlets` table added** — journalists now link to a publication entity; a journalist's outlet is essential for media targeting.
- **`pitches.status` uses CHECK constraint** — restricted to `('draft', 'sent', 'opened', 'replied')` to prevent invalid states.
- **`crm_relationships` gains `created_at` / `updated_at`** — relationship scores change over time; timestamps preserve auditability.
- **Soft-delete on `clients`, `journalists`, `campaigns`** — `deleted_at` (nullable) replaces hard deletes, preserving historical pitch records.
- **`pitch_versions` table** — each AI generation attempt is stored separately so users can compare outputs before sending.
- **Erroneous self-referential FK removed** — `Ref: campaigns.user_id < campaigns.title` from v1 was invalid and has been dropped.
- **`password_hash` uses bcrypt (cost ≥ 12)** — documented and enforced at the Go service layer.

### Schema (DBML)

```dbml
// ==========================================
// 1. CORE ENTITIES
// ==========================================

Table users {
  id            uuid      [primary key]
  email         varchar   [unique, not null]
  password_hash varchar   [not null, note: 'bcrypt cost >= 12']
  created_at    timestamp [default: `now()`]
  Note: 'PR professionals inside the agency'
}

Table outlets {
  id         uuid      [primary key]
  added_by   uuid      [null, note: 'user who created this record']
  name       varchar   [unique, not null]
  website    varchar
  country    varchar
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  deleted_at timestamp [null, note: 'soft delete']
  Note: 'Publications / media outlets — referenced by journalists'
}

Table clients {
  id         uuid      [primary key]
  user_id    uuid      [not null, note: 'owner of this client record']
  name       varchar   [not null]
  industry   varchar
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  deleted_at timestamp [null, note: 'soft delete']
  Note: 'Brands represented by the agency — user-scoped'
}

Table journalists {
  id         uuid      [primary key]
  added_by   uuid      [null, note: 'user who created this record']
  outlet_id  uuid      [not null, note: 'REQUIRED: which publication?']
  name       varchar   [not null]
  email      varchar   [unique, not null]
  niche      varchar
  created_at timestamp [default: `now()`]
  updated_at timestamp [default: `now()`]
  deleted_at timestamp [null, note: 'soft delete']
  Note: 'Shared agency media contact database'
}

// ==========================================
// 2. WORKFLOW ENTITIES
// ==========================================

Table crm_relationships {
  id                 uuid      [primary key]
  user_id            uuid      [not null]
  journalist_id      uuid      [not null]
  relationship_score int       [not null, default: 0, note: 'CHECK: 0-10']
  private_notes      text
  created_at         timestamp [default: `now()`]
  updated_at         timestamp [default: `now()`]
  Note: 'Personalised contact book per PR rep'

  indexes {
    (user_id, journalist_id) [unique, name: 'uq_crm_user_journalist']
  }
}

Table campaigns {
  id                 uuid      [primary key]
  user_id            uuid      [not null]
  client_id          uuid      [not null]
  title              varchar   [not null]
  press_release_text text
  created_at         timestamp [default: `now()`]
  updated_at         timestamp [default: `now()`]
  deleted_at         timestamp [null, note: 'soft delete']
  Note: 'Major press launches'
}

// ==========================================
// 3. ACTION ENTITIES
// ==========================================

Table pitches {
  id            uuid      [primary key]
  user_id       uuid      [not null]
  client_id     uuid      [not null]
  campaign_id   uuid      [null, note: 'OPTIONAL: part of a campaign?']
  journalist_id uuid      [not null]
  context_brief text      [note: 'AI source when no campaign is linked']
  status        varchar   [not null, default: 'draft',
                           note: "CHECK (status IN ('draft','sent','opened','replied'))"]
  created_at    timestamp [default: `now()`]
  updated_at    timestamp [default: `now()`]
  Note: 'Actual outreach messages — ad-hoc or campaign-linked'
}

Table pitch_versions {
  id                uuid      [primary key]
  pitch_id          uuid      [not null]
  version_number    int       [not null]
  ai_generated_body text      [not null]
  prompt_snapshot   text      [note: 'exact prompt sent to Gemini — for debugging & audit']
  generation_params text      [note: 'JSON: tone, length, angle, custom_instructions']
  created_at        timestamp [default: `now()`]
  Note: 'Each AI generation attempt; user picks the best before sending'

  indexes {
    (pitch_id, version_number) [unique, name: 'uq_pitch_version']
  }
}

// ==========================================
// 4. CONFIGURATION ENTITIES
// ==========================================

Table prompt_templates {
  id                  uuid      [primary key]
  user_id             uuid      [not null]
  name                varchar   [not null]
  tone                varchar   [note: 'formal/conversational/urgent/enthusiastic']
  length              varchar   [note: 'concise/standard/detailed']
  angle               varchar   [note: 'news_hook/exclusive/follow_up/thought_leadership/event']
  custom_instructions text
  created_at          timestamp [default: `now()`]
  updated_at          timestamp [default: `now()`]
  deleted_at          timestamp [null, note: 'soft delete']
  Note: 'Reusable AI generation presets per user'

  indexes {
    (user_id, name) [unique, name: 'uq_template_user_name']
  }
}

// ==========================================
// 5. FOREIGN KEYS
// ==========================================

Ref: outlets.added_by > users.id
Ref: journalists.added_by > users.id
Ref: journalists.outlet_id > outlets.id

Ref: clients.user_id > users.id

Ref: crm_relationships.user_id       > users.id
Ref: crm_relationships.journalist_id > journalists.id

Ref: campaigns.user_id   > users.id
Ref: campaigns.client_id > clients.id

Ref: pitches.user_id       > users.id
Ref: pitches.client_id     > clients.id
Ref: pitches.campaign_id   > campaigns.id
Ref: pitches.journalist_id > journalists.id

Ref: pitch_versions.pitch_id > pitches.id

Ref: prompt_templates.user_id > users.id
```

### Relationship Summary

```
users ──< clients ──< campaigns ──< pitches >── journalists >── outlets
  │                                    │
  ├──< crm_relationships >── journalists
  │
  ├──< prompt_templates
  │
  └──< pitches ──< pitch_versions
```

---

## 6. AI Integration & Cost Control

**Model:** `gemini-1.5-flash` via Google Gemini REST API.

### Context Injection pattern

```
[Go Backend — HTTP Handler]
  │
  ├── fetch Campaign.press_release_text   (full press release)
  ├── fetch Journalist.niche + Outlet.name (beat + publication)
  ├── fetch CRM.relationship_score         (1-10 warmth)
  │
  └── push GenerateJob onto channel
        │
        ▼
  [Async Worker (goroutine pool)]
        │
        ├── compile structured prompt
        ├── POST → Google Gemini API
        │
        └── on success: INSERT pitch_versions (version_number++)
                        UPDATE pitches.status = 'draft'
```

The Go backend is the sole keeper of the Gemini API key — it is never exposed to the frontend. The HTTP handler returns `202 Accepted` immediately; the frontend polls `GET /pitches/:id/versions` for results.

### Rate Limiting

A Gin middleware restricts AI generation calls per authenticated user:

| Endpoint group           | Limit                                 |
| ------------------------ | ------------------------------------- |
| `POST /auth/*`           | 10 req / min (brute-force protection) |
| `POST /pitches/generate` | 20 req / min / user (cost control)    |
| All other endpoints      | 120 req / min / user                  |

---

## 7. Authentication Flow

```
Client                   Backend (Gin)              PostgreSQL
  │                           │                          │
  │──── POST /auth/login ─────▶                          │
  │                           │──── SELECT user ────────▶│
  │                           │◀─── user row ────────────│
  │                           │  verify bcrypt hash      │
  │◀─── 200 { jwt_token } ────│                          │
  │                           │                          │
  │──── GET /pitches ─────────▶                          │
  │     Authorization: Bearer │                          │
  │                     validate JWT (middleware)        │
  │                           │──── SELECT pitches ─────▶│
  │◀─── 200 [pitches] ────────│                          │
```

**Password storage:** bcrypt with a minimum cost factor of 12. The algorithm choice is enforced at the `users` service layer in Go — no raw SHA/MD5 hashing is permitted.

---

## 8. Soft Delete Strategy

Tables with `deleted_at timestamp` use a **soft-delete pattern** — rows are never physically removed. GORM's `gorm.DeletedAt` type handles this automatically when the model embeds `gorm.Model`.

```go
// Example Go model
type Campaign struct {
    gorm.Model           // embeds ID, CreatedAt, UpdatedAt, DeletedAt
    UserID   uuid.UUID
    ClientID uuid.UUID
    Title    string
    // ...
}
```

All queries must filter `WHERE deleted_at IS NULL` (GORM does this automatically). This preserves historical pitch records even if a client or campaign is "deleted" by a user.

---

## 9. Pitch Versioning Flow

```
User clicks "Regenerate"
        │
        ▼
POST /pitches/:id/generate
        │
        ▼
  Worker generates new body
        │
        ▼
INSERT pitch_versions
  pitch_id       = :id
  version_number = MAX(version_number) + 1
  ai_generated_body = <new output>
  prompt_snapshot   = <exact prompt>
        │
        ▼
GET /pitches/:id/versions  →  user compares v1, v2, v3
        │
        ▼
POST /pitches/:id/select { version_id }  →  marks preferred version
```

---

## 10. Local Environment & Docker

The full stack is orchestrated via `docker-compose.yml` — a single `docker compose up` starts everything.

```yaml
# docker-compose.yml (overview)
services:
  frontend: # Next.js dev server         → localhost:3000
  backend: # Go Gin (air hot-reload)     → localhost:8080
  db: # PostgreSQL alpine            → localhost:5432
    volumes:
      - pg_data:/var/lib/postgresql/data # persistent data

volumes:
  pg_data:
```

### Environment Variables

| Variable              | Container | Description                                   |
| --------------------- | --------- | --------------------------------------------- |
| `DATABASE_URL`        | backend   | PostgreSQL connection string                  |
| `GEMINI_API_KEY`      | backend   | Never passed to frontend                      |
| `JWT_SECRET`          | backend   | Token signing key (min 32 chars)              |
| `BCRYPT_COST`         | backend   | bcrypt work factor (default: 12)              |
| `AI_RATE_LIMIT_RPM`   | backend   | Requests per minute per user for AI endpoints |
| `NEXT_PUBLIC_API_URL` | frontend  | Backend base URL                              |

---

## 11. Frontend UX Improvements

- **Optimistic UI for pitch status** — `status` updates locally on "Mark as Sent" before the API confirms, for a snappier experience.
- **Version picker UI** — after AI generation, a side-by-side diff view lets users compare `pitch_versions` before selecting one to send.
- **Cursor-based pagination on `/journalists`** — the global journalist DB can grow large; offset pagination is added from day one to prevent slow full-table scans.

</details>
