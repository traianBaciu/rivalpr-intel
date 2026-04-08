## Plan: RivalPR Intel — Full-Stack Implementation

Build the RivalPR Intel PR Agency CRM from your Architecture.md. 4 mandatory layers (Next.js, Go/Gin, PostgreSQL, Docker) plus Anthropic Claude AI for pitch generation. Structured into 4 phases aligned with the hackathon timeline, with custom Copilot agent modes per layer.

---

### Phase 0 — Project Scaffolding & Agent Setup (Start of Week 2)

1. Initialize Git repo with proper `.gitignore` (Go, Node, Docker, `.env`)
2. Create project folder structure:
   - `backend/` — Go module with `cmd/server/`, `internal/{models,handlers,middleware,services,worker,config}/`
   - `frontend/` — Next.js App Router with `src/{app,components,lib}/`
   - `.github/agents/` — 4 Copilot agent modes
   - `.github/copilot-instructions.md` — project-wide conventions
3. Create **4 custom Copilot agent modes**:
   - **`backend.agent.md`** — Go/Gin specialist (tools: read, edit, search, execute). Knows Gin routing, GORM, JWT, Go project layout. Restricted to `backend/` only.
   - **`frontend.agent.md`** — Next.js/shadcn/ui specialist (tools: read, edit, search, execute). Knows App Router, Tailwind, TypeScript. Restricted to `frontend/` only.
   - **`database.agent.md`** — PostgreSQL/GORM specialist (tools: read, edit, search). Knows model definitions, migrations, relationships, soft-delete, UUID PKs. Restricted to models/migration code.
   - **`devops.agent.md`** — Docker/compose/CI specialist (tools: read, edit, search, execute). Knows multi-stage Dockerfiles, compose orchestration, GitHub Actions. Restricted to infra files.
4. Create workspace-wide `copilot-instructions.md` with project conventions (naming, error handling, auth patterns)

---

### Phase 1 — Database & Backend Core (Week 2: Apr 8–14)

5. **Docker Compose baseline** (_parallel with step 6_) — `docker-compose.yml` with `frontend`, `backend`, `db` services. PostgreSQL alpine with persistent volume. Go backend with `air` hot-reload. `.env.example` with all 6 env vars.
6. **GORM models & auto-migration** (_parallel with step 5_) — All 8 tables: `User`, `Outlet`, `Client`, `Journalist`, `CrmRelationship`, `Campaign`, `Pitch`, `PitchVersion`. UUID PKs, `gorm.Model` for soft-delete tables, CHECK constraint on `pitches.status`, composite unique indexes.
7. **Auth system** (_depends on 6_) — `POST /auth/register` (bcrypt cost ≥ 12), `POST /auth/login` (returns JWT), JWT middleware extracting user ID into Gin context.
8. **CRUD API endpoints** (_depends on 7_) — Full REST for Clients, Outlets, Journalists (cursor-paginated), CRM Relationships, Campaigns, Pitches — all user-scoped where applicable. 6 resource groups.
9. **Rate limiting middleware** (_parallel with 8_) — Token bucket per user: 10/min auth, 20/min AI, 120/min other.
10. **Postman/Bruno collection** (_depends on 8_) — Exported collection with all endpoints + auth flow examples.

**Week 2 Deliverable:** Working API with core endpoints + API collection

---

### Phase 2 — Frontend & Integration (Week 3: Apr 15–21)

11. **Next.js project setup** — App Router, TypeScript, Tailwind, shadcn/ui. API client utility with JWT token management. Auth context/provider.
12. **Auth pages** (_depends on 11_) — Login (`/login`), Register (`/register`), auth guard middleware.
13. **Core views (5 pages)** (_depends on 12_):
    - **Dashboard** (`/`) — client count, active campaigns, recent pitches
    - **Clients** (`/clients`) — list, create, edit, soft-delete
    - **Journalists** (`/journalists`) — paginated list, outlet selection, CRM score slider
    - **Campaigns** (`/campaigns`) — list, create with press release text
    - **Pitches** (`/pitches`) — create, generate AI version, version picker with side-by-side comparison
14. **Docker integration** (_parallel with 13_) — Frontend Dockerfile (multi-stage), verify `docker compose up` runs all 3 services.
15. **Responsive design pass** (_depends on 13_) — Mobile viewport support, consistent component usage.

**Week 3 Deliverable:** FE connected to BE + docker-compose running locally

---

### Phase 3 — AI Integration & Polish (Week 4: Apr 22–30)

16. **Anthropic Claude integration** — Async goroutine worker + job channel. Context injection: press release, journalist niche, outlet name, CRM score. POST to Claude API. Store `pitch_versions` with `prompt_snapshot`.
17. **Optimistic UI** (_depends on 16_) — Local status updates on "Mark as Sent", loading states with polling for AI generation.
18. **README.md** (_parallel with 16_) — Project description, prerequisites, setup, architecture overview, env var docs, `docker compose up` instructions.
19. **Unit tests** (_bonus +5 pts, parallel with 16_) — Backend tests for auth service and CRUD handlers.
20. **CI/CD pipeline** (_bonus +5 pts, parallel with 19_) — GitHub Actions: lint, test, build Docker images.
21. **Demo prep** (_depends on all_) — Seed DB with realistic PR data. Practice demo flow: register → create client → add journalist → create campaign → generate AI pitch → compare versions → send. Prepare architecture defense.

**Week 4 Deliverable:** Live demo + code walkthrough + architecture defense

---

### Relevant Files

| Area          | Path                                       | Purpose                            |
| ------------- | ------------------------------------------ | ---------------------------------- |
| Agents        | `.github/agents/backend.agent.md`          | Go/Gin coding agent                |
| Agents        | `.github/agents/frontend.agent.md`         | Next.js/shadcn agent               |
| Agents        | `.github/agents/database.agent.md`         | GORM/PostgreSQL agent              |
| Agents        | `.github/agents/devops.agent.md`           | Docker/CI agent                    |
| Instructions  | `.github/copilot-instructions.md`          | Project-wide conventions           |
| Backend entry | `backend/cmd/server/main.go`               | App entrypoint, router, DB connect |
| Models        | `backend/internal/models/*.go`             | 8 GORM models                      |
| Handlers      | `backend/internal/handlers/*.go`           | HTTP handlers per resource         |
| Auth MW       | `backend/internal/middleware/auth.go`      | JWT validation                     |
| Rate limit    | `backend/internal/middleware/ratelimit.go` | Per-user rate limiting             |
| AI worker     | `backend/internal/worker/worker.go`        | Async Claude API goroutine pool    |
| AI client     | `backend/internal/services/anthropic.go`   | Anthropic API wrapper              |
| Config        | `backend/internal/config/config.go`        | Env var loader                     |
| Frontend API  | `frontend/src/lib/api.ts`                  | REST client with JWT               |
| Pages         | `frontend/src/app/(dashboard)/*/page.tsx`  | 5 core views                       |
| Docker        | `docker-compose.yml`                       | Orchestrates all 3 services        |
| CI/CD         | `.github/workflows/ci.yml`                 | Bonus: GitHub Actions pipeline     |

---

### Verification

1. `docker compose up` — all 3 containers start; frontend at :3000, backend at :8080, DB at :5432
2. Register user → login → receive JWT → use for authenticated requests
3. CRUD all resources (clients, outlets, journalists, campaigns, pitches) via API
4. Trigger AI pitch generation → verify `pitch_versions` created with body + prompt snapshot
5. Navigate all 5 frontend views — real backend data, no hardcoded mocks
6. Full demo flow: register → client → journalist with outlet → campaign → generate pitch → compare versions → send
7. Rate limiting: confirm 429 responses when exceeding limits
8. Soft delete: delete a client, verify it disappears but historical pitches remain

---

### Decisions

- **Go backend** — aligns with company's V3 rewrite evaluation; demonstrates concurrency via goroutine worker pool
- **PostgreSQL** — relational integrity for PR domain (pitches → clients, campaigns, journalists)
- **shadcn/ui** — maximizes frontend velocity for solo developer
- **Async AI (202 + polling)** — prevents request timeouts on Claude API calls
- **Scope boundary** — Core CRM + AI pitch generation. **Excluded**: real email sending, analytics dashboards, multi-tenancy

---

### Further Considerations

1. **Anthropic API key**: Obtain before Week 4. During Weeks 2-3, stub the AI worker with mock responses so the full flow is testable without the key.
2. **Live deployment (+5 bonus)**: Docker images make this straightforward via Railway or Fly.io. Decision deferred — plan supports it.
3. **Test coverage (+5 bonus)**: Focus on backend unit tests for auth and pitch generation — highest ROI for a solo build.
