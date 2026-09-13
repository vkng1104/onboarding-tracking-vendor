# Vendor Onboarding Tracker

A small SPA for coordinators to track seeded vendors through onboarding, see how long each vendor has remained in
its current stage, and retain an attributable history of every stage change.

> Status: every behavior the assignment asks for is implemented and verified.

## Business problem and approach

| Spreadsheet problem | Application response |
| --- | --- |
| A manually edited "Last Updated" value is easy to forget, so stalled vendors go unnoticed. | The server records `stage_entered_at`, calculates elapsed time, and highlights non-Active vendors that exceed the configured threshold. |
| The sheet retains only the latest value, so incorrect data cannot be explained. | Every stage change appends who changed it, when it happened, and the previous and new stages. |
| Free-text cells and three concurrent coordinators allow invalid values and silent overwrites. | The API validates the five stages and checks the client-observed stage while holding a database row lock. |

The application is a small modular monolith: a React SPA calls a Go API, and PostgreSQL stores both the current
vendor state and append-only transition history. Vendor creation is intentionally excluded, so representative
vendors and coordinator accounts are seeded for the reviewer.

## Run locally

Prerequisites: Docker with Compose. Local development additionally uses Go 1.25+ and pnpm 11+.

```bash
cp .env.example .env
docker compose up --build
```

Open <http://localhost:5173>. The app redirects to the coordinator login page. API readiness is available at
<http://localhost:8080/health>.

### Demo recording

https://github.com/user-attachments/assets/b0cf4bd2-3296-493f-84ed-e867892af14e

### Reviewer walkthrough

1. Log in with any demo account below.
2. Confirm all six vendors are visible, that the two stuck vendors lead the list, and that the completed vendor
   sits last. Use the **Assigned to me** filter to narrow the presentation.
3. Hover or focus the information icon beside **Need attention** to review the stuck-vendor rule.
4. Select **Review** for a vendor, choose any different stage, and submit the update. Earlier stages are accepted
   because they can represent a correction.
5. Confirm the current stage and elapsed time refresh, then inspect the new history entry for its actor and time.
6. Log out and confirm the protected dashboard returns to the login screen.

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

Pull requests run Go vet, Go tests with the race detector, PostgreSQL integration tests, frontend linting,
type-checking, tests, and the production build in GitHub Actions.

## Architecture

The backend is a Go modular monolith using chi and pgx. PostgreSQL holds the current vendor state and append-only
transition history; each stage change updates both atomically under a row lock and attributes the event to the
authenticated coordinator. The API derives elapsed hours, stuck state, and the informational expected next stage.
The React/Vite frontend uses React Router and TanStack Query, and its components follow pragmatic Atomic Design.
Docker Compose starts PostgreSQL, applies migrations, performs a non-destructive seed, and then starts the API and
web application.

Detailed decisions and diagrams live in
[`docs/RFC-001-vendor-onboarding-tracker.md`](docs/RFC-001-vendor-onboarding-tracker.md).

### Important technical decisions

- **PostgreSQL instead of frontend-only state:** the assignment's audit trail is durable and the current-state
  update can share a real transaction with its history insert.
- **Raw parameterized SQL with pgx:** the queries stay small and explicit. Values are bound separately from SQL,
  and the application allow-lists stage inputs.
- **Atomic transitions:** `SELECT ... FOR UPDATE` serializes changes to one vendor. The request includes
  `expected_current_stage`, so a stale coordinator receives `409 STAGE_CONFLICT` rather than overwriting a newer
  change. The update and history insert either both commit or both roll back.
- **Session-derived attribution:** the request cannot choose the actor; the API reads the coordinator from the
  authenticated session and authors the timestamp.
- **Server-derived stuck state and ordering:** a vendor is stuck only when it is not Active and has spent strictly
  more than `STUCK_AFTER_DAYS` in its current stage. Active vendors are complete, not stuck. The API returns the
  list already ordered by attention needed — stuck first, then the longest wait, with completed vendors last — so
  the client never re-derives the rule.
- **Pragmatic Atomic Design:** reusable controls and patterns are separated into atoms and molecules, composed
  dashboard sections are organisms, layouts are templates, and route-level data orchestration stays in pages.

## Scope and assumptions

The following assumptions make ambiguous workflow behavior explicit:

- Any authenticated coordinator can view and transition any vendor. Assignment is displayed and filterable but is
  not an authorization boundary.
- Any different one of the five stages is valid, including an earlier stage for correction. A same-stage update is
  rejected and creates no history row.
- Server UTC timestamps are authoritative; coordinators never enter a "Last Updated" value.
- "More than 7 days" means strictly more than 168 hours by default. The threshold is configurable through
  `STUCK_AFTER_DAYS`.
- Vendor creation/deletion, account administration, production authentication, external compliance and activation
  systems, notifications, real-time updates, deployment infrastructure, and coordinator reassignment are out of
  scope.

## Shortcuts and trade-offs

- Sessions are held in a mutex-protected in-memory store for eight hours. Restarting the API logs everyone out;
  HTTPS, persistent sessions, CSRF tokens beyond SameSite protection, and rate limiting would be required in
  production.
- The list is intentionally unpaginated and refreshes after mutations. That keeps the seeded assignment simple but
  would need server-side filtering and pagination at larger scale.
- Backward corrections do not require a reason or confirmation. A production workflow should capture one.
- The stage itself acts as a lightweight concurrency token. A version and idempotency key could replay a response
  after an ambiguous retry instead of returning a conflict.

## Testing strategy

Tests focus on business risks rather than a coverage percentage:

- authentication tests prove invalid or expired sessions cannot reach protected routes and logout invalidates the
  old cookie;
- stuck-state tests cover 167, 168, and 169 hours plus an old Active vendor, and a separate test pins the
  attention-first ordering including its name tie-break;
- HTTP tests cover malformed requests, stable domain errors, session-derived attribution, and non-leaking internal
  failures;
- PostgreSQL integration tests prove backward transitions, atomic rollback, attributable history, and exactly one
  winner when two coordinators submit changes from the same observed stage;
- frontend tests cover protected routing, filtering, direct vendor URLs, stage updates, conflict refresh, history,
  and recoverable errors.

## Completed, remaining, and next

Every behavior the assignment asks for is implemented, against its four numbered requirements:

| Requirement | What ships |
| --- | --- |
| 1. Login | Seeded bcrypt accounts, server-side sessions, protected routes, and logout |
| 2. View vendors | Six seeded vendors with stage, region, coordinator, notes, time in stage, and health |
| 3. Update a vendor's stage | Transactional change with append-only history attributed to the session |
| 4. Identify stuck vendors | Server-derived from `stage_entered_at`, configurable, highlighted and ordered first |

Nothing from the assignment remains. Production hardening and deployment are deliberately outside the requested
scope.

The first product improvement would be coordinator assignment/reassignment with append-only ownership history.
After that, corrective transitions could require a reason, and persistent sessions, pagination, observability,
accessibility audits, and end-to-end browser tests could be added as the system grows.
