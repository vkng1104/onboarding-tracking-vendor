# Vendor Onboarding Tracker

A small SPA for coordinators to track seeded vendors through onboarding, see how long each vendor has remained in
its current stage, and retain an attributable history of every stage change.

> Implementation status: Phase 4 of 5 is complete. Coordinators can review all seeded vendors, identify stuck
> vendors, inspect durable history, and move a vendor to any different workflow stage, including an earlier stage.

## Run locally

Prerequisites: Docker with Compose. Local development additionally uses Go 1.25+ and pnpm 11+.

```bash
cp .env.example .env
docker compose up --build
```

Open <http://localhost:5173>. The app redirects to the coordinator login page. API readiness is available at
<http://localhost:8080/health>.

### Demo accounts

| Coordinator | Email | Password |
| --- | --- | --- |
| Linh Nguyen | `linh@demo.local` | `demo1234` |
| Huy Tran | `huy@demo.local` | `demo1234` |
| Mai Pham | `mai@demo.local` | `demo1234` |

Stop the stack without deleting data:

```bash
make down
```

Remove containers and the local database volume:

```bash
make down-volumes
```

## Developer commands

| Command | Purpose |
| --- | --- |
| `make up` | Build and run the full local stack. |
| `make down` | Stop containers and preserve database data. |
| `make down-volumes` | Stop containers and remove local volumes. |
| `make logs` | Follow Compose logs. |
| `make migrate` | Apply pending application migrations. |
| `make seed` | Idempotently insert missing development fixtures. |
| `make seed-reset` | **Destructively** restore demo fixtures to their initial state. |
| `make test-unit` | Run backend and frontend unit tests. |
| `make test-integration` | Run tagged backend tests in a separate Compose project against an ephemeral test database. |

For checks outside Docker:

```bash
go test ./...
pnpm --dir web lint
pnpm --dir web typecheck
pnpm --dir web test --run
pnpm --dir web build
```

## Architecture

The backend is a Go modular monolith using chi and pgx. PostgreSQL holds the current vendor state and append-only
transition history; each stage change updates both atomically under a row lock and attributes the event to the
authenticated coordinator. The API derives elapsed hours, stuck state, and the informational expected next stage.
The React/Vite frontend uses React Router and TanStack Query, and its components follow pragmatic Atomic Design.
Docker Compose starts PostgreSQL, applies migrations, performs a non-destructive seed, and then starts the API and
web application.

Detailed decisions and diagrams live in
[`docs/RFC-001-vendor-onboarding-tracker.md`](docs/RFC-001-vendor-onboarding-tracker.md).

## Scope and assumptions

The locked product scope is a coordinator-only view over seeded vendor data. Authentication is intentionally
local-only: sessions live in API memory for eight hours, so restarting the API requires signing in again. Vendor
creation/deletion, production authentication, external compliance/activation integrations, and coordinator
assignment/reassignment are out of scope. Assignment/reassignment with append-only history is the first planned
improvement after the assignment.

## Delivery phases

1. Repository and runnable shell.
2. Seeded coordinator login.
3. Vendor dashboard, stuck-state calculation, and history view.
4. Transactional stage transition and audit trail.
5. Documentation and release QA.

## Testing strategy

Tests focus on business risks rather than framework coverage: identity comes from the server session, invalid or
expired sessions cannot reach protected routes, the stuck threshold is correct at its time boundary, Active is
never stuck, direct vendor URLs recover safely, and internal errors do not leak. Stage-transition tests cover
backward corrections, invalid and unchanged stages, stale concurrent writes, and rollback if audit insertion fails.
