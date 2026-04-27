# RivalPR Intel — Copilot Workspace Instructions

## Project Overview

RivalPR Intel is a PR Agency CRM and Outreach Tracker. A 3-container Docker application:

- **Frontend**: Next.js (App Router) + Tailwind CSS + shadcn/ui → `frontend/`
- **Backend**: Go (Golang) + Gin + GORM → `backend/`
- **Database**: PostgreSQL (alpine) → managed via docker-compose

Architecture doc: `Architecture.md`
Implementation plan: `.github/prompts/plan-rivalPrIntel.prompt.md`

---

## Conventions

### General

- Never hardcode secrets. Use environment variables loaded from `.env` (see `.env.example`).
- All IDs are UUIDs (`github.com/google/uuid`), never auto-increment integers.
- Use soft-delete (`deleted_at`) instead of hard deletes on `clients`, `journalists`, `campaigns`, `outlets`.
- Follow the existing folder structure strictly — do not create files outside established directories without a clear reason.

### Go / Backend

- Package names: lowercase, no underscores (e.g., `handlers`, `models`, `middleware`).
- Error handling: always return structured JSON error responses `{ "error": "message" }` with appropriate HTTP status codes. Never panic in handlers.
- GORM: use `gorm.Model` for entities with soft-delete. Use `uuid.UUID` as primary key type.
- JWT: extract user ID from the token in middleware and inject into `gin.Context` as `"userID"`.
- bcrypt cost minimum: 12. Enforced in the `users` service layer.
- Validation: use `github.com/go-playground/validator/v10` on request structs.
- All handlers live in `backend/internal/handlers/`, one file per resource (e.g., `clients.go`, `pitches.go`).

### Next.js / Frontend

- Use the App Router (`src/app/`) exclusively — no `pages/` directory.
- All API calls go through `src/lib/api.ts` — never call `fetch` directly in components.
- JWT token stored in `localStorage` under the key `"rivalpr_token"`.
- Use shadcn/ui components from `src/components/ui/` for all UI primitives.
- TypeScript strict mode. No `any` types.
- Server Components by default; add `"use client"` only when interactivity or hooks are needed.

### Database

- All table names are plural snake_case (enforced by GORM conventions).
- The `pitches.status` column must have a CHECK constraint: `('draft', 'sent', 'opened', 'replied')`.
- Composite unique indexes: `(user_id, journalist_id)` on `crm_relationships`, `(pitch_id, version_number)` on `pitch_versions`.

### Docker

- Each service must have a dedicated `Dockerfile` inside its folder (`backend/Dockerfile`, `frontend/Dockerfile`).
- No secrets in `docker-compose.yml` — use `env_file: .env`.
- Backend uses `air` for hot-reload in development.

---

## Key Environment Variables

| Variable              | Container | Notes                        |
| --------------------- | --------- | ---------------------------- |
| `DATABASE_URL`        | backend   | PostgreSQL connection string |
| `GEMINI_API_KEY`      | backend   | Never passed to frontend     |
| `JWT_SECRET`          | backend   | Min 32 characters            |
| `BCRYPT_COST`         | backend   | Default: 12                  |
| `AI_RATE_LIMIT_RPM`   | backend   | Per-user AI rate limit       |
| `NEXT_PUBLIC_API_URL` | frontend  | Backend base URL             |
