# Vendor Onboarding Tracker

A small SPA for coordinators to track seeded vendors through onboarding, see how long each vendor has remained in
its current stage, and retain an attributable history of every stage change.

> Implementation status: Phase 1 of 5 is complete. The repository shell, database-backed health check, React
> application shell, Compose orchestration, and documentation structure are available. Login and vendor workflows
> are intentionally introduced in later phases.

## Run locally

Prerequisites: Docker with Compose. Local development additionally uses Go 1.25+ and pnpm 11+.

```bash
cp .env.example .env
docker compose up --build
```

Open <http://localhost:5173>. API readiness is available at <http://localhost:8080/health>.

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

The backend is a Go modular monolith using chi and pgx. PostgreSQL will hold the current vendor state and its
append-only transition history in one transaction. The React/Vite frontend uses React Router and TanStack Query;
its components follow pragmatic Atomic Design. Docker Compose starts PostgreSQL, applies migrations, performs a
non-destructive seed, and then starts the API and web application.

Detailed decisions and diagrams live in
[`docs/RFC-001-vendor-onboarding-tracker.md`](docs/RFC-001-vendor-onboarding-tracker.md).

## Scope and assumptions

Phase 1 establishes infrastructure only. The locked product scope is a coordinator-only view over seeded vendor
data. Vendor creation/deletion, production authentication, external compliance/activation integrations, and
coordinator assignment/reassignment are out of scope. Assignment/reassignment with append-only history is the
first planned improvement after the assignment.

## Delivery phases

1. Repository and runnable shell.
2. Seeded coordinator login.
3. Vendor dashboard, stuck-state calculation, and history view.
4. Transactional stage transition and audit trail.
5. Documentation and release QA.

## Testing strategy

Tests focus on business risks rather than framework coverage: legal workflow movement, the stuck-time boundary,
and the guarantee that the current state and audit history cannot disagree. Phase 1 includes a focused readiness
handler test; later phases add the domain, HTTP, frontend, transaction, and concurrency cases.