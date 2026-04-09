# ARCHITECTURE.md — RivalPR Intel

> System Design Document · April 2025 · v2.0

---

## 1. Overview & Purpose

| Field            | Value                                                                                                                                                                                                                                                                     |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Project Name** | RivalPR Intel                                                                                                                                                                                                                                                             |
| **Description**  | A streamlined PR Agency CRM and Outreach Tracker. Allows PR professionals to manage clients, track media contacts in a shared database, score personal relationships, and generate hyper-personalized pitch emails using Anthropic's Claude AI based on campaign context. |
| **Developer**    | [Your Name]                                                                                                                                                                                                                                                               |
| **Timeline**     | April 1–30, 2025                                                                                                                                                                                                                                                          |
| **Doc Version**  | v2.0 — includes schema hardening, async AI generation, soft-delete, pitch versioning                                                                                                                                                                                      |

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
│         │  │ Anthropic Claude│        │                   │
│         │  │ claude-3-5-sonnet        │                   │
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
| AI            | Anthropic Claude API     | claude-3-5-sonnet                  |
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

| Category     | Tables                                       | Purpose                                 |
| ------------ | -------------------------------------------- | --------------------------------------- |
| **Core**     | `users`, `clients`, `journalists`, `outlets` | Foundation — agency-wide shared data    |
| **Workflow** | `campaigns`, `crm_relationships`             | Press releases & personalised scoring   |
| **Action**   | `pitches`, `pitch_versions`                  | AI-generated outreach + version history |

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
  id         uuid    [primary key]
  name       varchar [unique, not null]  // e.g. TechCrunch, Forbes Romania
  website    varchar
  country    varchar
  created_at timestamp [default: `now()`]
  deleted_at timestamp [null, note: 'soft delete']
  Note: 'Publications / media outlets — referenced by journalists'
}

Table clients {
  id         uuid      [primary key]
  name       varchar   [not null]
  industry   varchar   // e.g. Tech, Auto, FMCG
  created_at timestamp [default: `now()`]
  deleted_at timestamp [null, note: 'soft delete']
  Note: 'Brands represented by the agency'
}

Table journalists {
  id         uuid    [primary key]
  outlet_id  uuid    [not null, note: 'REQUIRED: which publication?']
  name       varchar [not null]
  email      varchar [unique, not null]
  niche      varchar // e.g. Software, Lifestyle
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
  relationship_score int       [not null, note: 'Score 1-10']
  private_notes      text
  created_at         timestamp [default: `now()`]
  updated_at         timestamp [default: `now()`, note: 'updated on every score change']
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
  deleted_at         timestamp [null, note: 'soft delete']
  Note: 'Major press launches'
}

// ==========================================
// 3. ACTION ENTITIES
// ==========================================

Table pitches {
  id            uuid    [primary key]
  user_id       uuid    [not null]
  client_id     uuid    [not null,  note: 'REQUIRED: which client?']
  campaign_id   uuid    [null,      note: 'OPTIONAL: part of a campaign?']
  journalist_id uuid    [not null,  note: 'REQUIRED: target journalist']
  context_brief text    [note: 'AI source when no campaign is linked']
  status        varchar [not null, default: 'draft',
                         note: "CHECK (status IN ('draft','sent','opened','replied'))"]
  created_at    timestamp [default: `now()`]
  Note: 'Actual outreach messages — ad-hoc or campaign-linked'
}

Table pitch_versions {
  id                uuid      [primary key]
  pitch_id          uuid      [not null]
  version_number    int       [not null, note: 'increments per pitch']
  ai_generated_body text      [not null]
  prompt_snapshot   text      [note: 'exact prompt sent to Claude — for debugging & audit']
  created_at        timestamp [default: `now()`]
  Note: 'Stores each AI generation attempt; user picks the best before sending'

  indexes {
    (pitch_id, version_number) [unique, name: 'uq_pitch_version']
  }
}

// ==========================================
// 4. FOREIGN KEYS
// ==========================================

Ref: journalists.outlet_id > outlets.id

Ref: crm_relationships.user_id       > users.id
Ref: crm_relationships.journalist_id > journalists.id

Ref: campaigns.user_id   > users.id
Ref: campaigns.client_id > clients.id

Ref: pitches.user_id       > users.id
Ref: pitches.client_id     > clients.id
Ref: pitches.campaign_id   > campaigns.id
Ref: pitches.journalist_id > journalists.id

Ref: pitch_versions.pitch_id > pitches.id
```

### Relationship Summary

```
users ──< campaigns ──< pitches >── journalists >── outlets
  │                        │
  └──< crm_relationships >─┘
           │
        journalists

clients ──< campaigns
clients ──< pitches

pitches ──< pitch_versions
```

---

## 6. AI Integration & Cost Control

**Model:** `claude-3-5-sonnet` via Anthropic REST API.

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
        ├── POST → Anthropic API
        │
        └── on success: INSERT pitch_versions (version_number++)
                        UPDATE pitches.status = 'draft'
```

The Go backend is the sole keeper of the Anthropic API key — it is never exposed to the frontend. The HTTP handler returns `202 Accepted` immediately; the frontend polls `GET /pitches/:id/versions` for results.

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
| `ANTHROPIC_API_KEY`   | backend   | Never passed to frontend                      |
| `JWT_SECRET`          | backend   | Token signing key (min 32 chars)              |
| `BCRYPT_COST`         | backend   | bcrypt work factor (default: 12)              |
| `AI_RATE_LIMIT_RPM`   | backend   | Requests per minute per user for AI endpoints |
| `NEXT_PUBLIC_API_URL` | frontend  | Backend base URL                              |

---

## 11. Frontend UX Improvements

- **Optimistic UI for pitch status** — `status` updates locally on "Mark as Sent" before the API confirms, for a snappier experience.
- **Version picker UI** — after AI generation, a side-by-side diff view lets users compare `pitch_versions` before selecting one to send.
- **Cursor-based pagination on `/journalists`** — the global journalist DB can grow large; offset pagination is added from day one to prevent slow full-table scans.
