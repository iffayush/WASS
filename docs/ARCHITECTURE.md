# WASS — Architecture

Status: v1 (M0). Component map, data flow, and directory layout. Decisions live in
`docs/DECISIONS.md`; the product behavior in `docs/SPEC.md`.

## Components

```
                    ┌──────────────────────────────────────────────────┐
                    │                     frontend/                    │
                    │   React + TypeScript + Tailwind (built in M9)    │
                    │   new scan · live progress · findings · reports  │
                    └───────────────▲──────────────────────────────────┘
                                    │ REST + SSE (auth: GitHub App session)
                    ┌───────────────┴──────────────────────────────────┐
                    │                      server/                     │
                    │  API · auth · queue workers · SSE · persistence  │
                    │  ownership verifier · rate limits · redaction    │
                    └───────▲───────────────────────────▲──────────────┘
                            │ job rows                  │ findings (schema only)
                    ┌───────┴──────────┐      ┌──────────┴───────────────┐
                    │    Postgres      │      │         engine/          │
                    │  scans, jobs,    │      │  scanner core · modules  │
                    │  findings, users │      │  (secrets, exposed_files,│
                    │  (pgx v5 pool)   │      │  headers) · fetchers ·   │
                    └──────────────────┘      │  CLI (JSON out)          │
                                              └──────────────────────────┘
```

- **engine/** — the scanner. Standalone CLI (`wass scan --repo <url|path>` /
  `wass scan --domain <url>`) emitting JSON. Contains: scanner core, the module interface, the
  finding schema + redaction, fetchers (with SSRF guards), git history reader, domain crawler,
  and the secrets rules engine. **Never imports server or frontend.**
- **server/** — wraps the engine. REST API + SSE progress, GitHub App auth, Postgres-backed job
  queue (`FOR UPDATE SKIP LOCKED`), persistence, ownership verifier, rate limiting, encryption of
  evidence/tokens at rest. Knows the finding schema and `RedactedFinding` only — not engine
  internals.
- **frontend/** — web UI (M9). Talks to the server only.
- **rules/** — YAML secret rules + fixtures, loaded by the engine.

## Data flow (scan lifecycle)

1. User submits a scan (repo URL or domain) via API. Server validates, applies rate limits,
   creates a `scan` row + queued `job` row.
2. A worker claims the job (`SELECT … FOR UPDATE SKIP LOCKED`), sets progress, and invokes the
   engine for the scan mode.
3. Engine does the work:
   - *Repo mode:* hardened clone → working-tree scan → history scan (per-commit diffs, capped) →
     findings with first-seen/still-present.
   - *Domain mode:* SSRF-guarded crawl within budget → JS/hydration/source-map collection →
     windowed scanning → (optionally, after ownership proof) provider checks / active probes.
4. Engine emits findings in the common schema → server redacts (`RedactedFinding`), fingerprints,
   dedups, persists, and streams progress over SSE.
5. Report rendering + exports read **only** `RedactedFinding` data. Rescan diff compares on
   `(scan_id, fingerprint)`.

## Redaction boundary (the load-bearing wall)

```
engine: Finding (may hold raw value, in memory only) ──redact()──▶ RedactedFinding
                                                                       │
        renderers, exporters (MD/PDF/SARIF/JSON), API responses,       ▼
        share links, LLM client ─────────────accept ONLY RedactedFinding
```

`RedactedFinding` does not expose the raw value through any method or field. Converting is
one-way and happens exactly once, in the engine→server handoff.

## Safety architecture

- **fetcher (engine):** every outbound HTTP request goes through one guarded client: DNS resolve
  → IP blocklist (private/loopback/link-local/metadata `169.254.169.254`) → connect with pinned
  IP → re-verify on every redirect (capped) → size + time caps → https-only schemes.
  Regression matrix in `engine/fetcher` tests.
- **ownership verifier (server):** DNS TXT `_wass-<token>.domain` or token file at
  `https://domain/.well-known/wass-<token>.txt`. Gates active probes and provider checks.
- **limits (server + engine):** per-scan timeout, page/size caps, crawl budget, per-user rate
  limit + concurrent-scan cap, per-scan outbound request cap, global egress bandwidth cap,
  negative cache for repeat-timeout hosts.
- **clone hardening (engine):** see D-007 checklist.
- **secrets handling:** redact before storage; encrypt evidence + tokens at rest; scrub logs;
  fixtures are obviously fake; gitleaks in CI + pre-commit.

## Directory layout

```
wass/
  engine/            # scanner core, modules, fetchers, CLI (Go package per concern)
    fetcher/         #   SSRF-guarded HTTP client + regression matrix
  server/            # API, queue workers, SSE, persistence, auth
  frontend/          # web UI (M9)
  rules/             # YAML rules + fixtures
  benchmark/         # corpus + scoring harness (post-MVP)
  docs/              # SPEC, ARCHITECTURE, TASKS, DECISIONS, HANDOFF
  cmd/wass/          # CLI entry point (`go run ./cmd/wass scan --repo …`)
  Makefile           # test / vet / lint / secret-scan / compose targets
  docker-compose.yml # Postgres for local dev
  .github/workflows/ # CI: test, vet, gitleaks
  AGENTS.md          # canonical agent instructions
  CLAUDE.md          # → AGENTS.md
  .cursor/rules/wass.mdc  # → AGENTS.md
```

Go module: `github.com/iffayush/wass` (D-004). Go toolchain: 1.26.x, minimum 1.22 enforced in
`go.mod` (per-iteration loop variables, D-003).

## Technology choices (summary — rationale in DECISIONS.md)

| Concern | Choice | Alternative rejected/deferred |
|---|---|---|
| Language | Go 1.26 (min 1.22) | — |
| HTTP router | stdlib `net/http` mux (1.22 patterns) | chi — add if middleware needs outgrow stdlib |
| Postgres driver | pgx v5 + `pgxpool` | `lib/pq` (maintenance mode) |
| Queue | Postgres `SKIP LOCKED` table | River (fallback), Redis (no) |
| Repo auth | GitHub App (metadata:read, contents:read) | OAuth App (repo scope = read/write everywhere); fine-grained PAT |
| Rules | WASS YAML + fixtures | — |
| Frontend | React + TypeScript + Tailwind (M9) | — |
| Secret scanning of our own repo | gitleaks (bootstrap) | WASS itself (post-MVP dogfooding) |
