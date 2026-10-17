# WASS — Tasks

One milestone at a time. Do not start a milestone until the previous one's acceptance criteria
pass. Each unchecked task should be small enough for one session. Check off as completed and
commit after each.

**Acceptance gate for every milestone:** `make test` (or `go test ./...` on Windows) and CI green;
docs updated; HANDOFF rewritten.

---

## M0 — Audit and reset ✅ (complete)

- [x] Analyze the master build prompt; verify assumptions against the (then-existing) repo state
- [x] Owner decisions: full reset (D-001), frontend starts empty (D-002)
- [x] Fresh git repository initialized; scoped to the project directory; `.gitignore` for Go/Node/Windows/env
- [x] Toolchain verified (Go 1.26.5, Docker 28.2.2, Git 2.45.2; make/gitleaks absent — D-010)
- [x] `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/wass.mdc` (canonical instructions + pointers)
- [x] `docs/SPEC.md`, `docs/ARCHITECTURE.md`, `docs/DECISIONS.md` (D-001…D-011), `docs/TASKS.md`, `docs/HANDOFF.md`
- [x] Scaffold: `go.mod`, `engine/` + `server/` placeholder packages with passing tests, `cmd/wass/` CLI stub, `rules/` placeholder
- [x] `Makefile` (test, vet, lint, secret-scan, compose targets) + Windows fallbacks documented
- [x] `docker-compose.yml` (Postgres only)
- [x] CI: `.github/workflows/ci.yml` (go test, go vet, gitleaks secret scan)
- [x] Pre-commit hook (gitleaks, fail-closed) enabled via `core.hooksPath`
- [x] New `README.md` (old orchestrator README removed — D-001)
- [x] Conventional commits; HANDOFF written
- **Acceptance criteria:** `make test` and CI pass on the empty scaffold; docs exist; git-clone
  safety checklist (D-007), Go version (D-003), pgx choice (D-005), and GitHub App decision
  (D-006) recorded; CI includes a secret-scan step and a pre-commit hook exists.

## M1 — Schema and engine skeleton ✅ (complete)

- [x] Define the versioned common finding schema (all fields per SPEC §4) with unit tests (`engine/finding`)
- [x] Implement `RedactedFinding` with the one-way `FromCandidate` boundary (first4+last4+hash, ≤12 chars fully masked); raw values live only in `Candidate`, whose `MarshalJSON` is a hard tripwire; tested that raw values cannot leak
- [x] Define the module interface (`engine.Module`: `Name()` + `Scan(ctx, target, emit)`); modules emit candidates only
- [x] Scanner core: run registered modules, convert candidates behind the boundary, dedup on `(rule_id, fingerprint)`, assign IDs, validate every finding
- [x] CLI `cmd/wass`: `scan --repo <path>` emitting `RedactedReport` JSON to stdout; `--domain` rejects with an M4 pointer; fingerprinting: HMAC-SHA256 of normalized raw value (trim + strict base64-unwrap), versioned key, hard error without key (D-008)
- [x] Stub module (`engine/stubmod`) with generated fixtures: dedup, skip lists (.git/node_modules/vendor), binary + size guards, snippet windowing — all tested
- [x] `.gitleaks.toml` allowlist scoped to `testdata/` so fake fixtures pass our own fail-closed hook
- **Acceptance:** unit tests for schema and redaction pass; `wass scan` runs on a fixture directory with a stub module and emits JSON (verified live: 2 locations dedup to 1 finding, evidence redacted, raw value absent from output).

## M2 — Secrets rules engine (working tree)

- [ ] Add golangci-lint to CI (deferred from M0 per D-010; no lint config exists yet)
- [ ] YAML rule format + loader (id, description, regex, keywords, entropy, provider, default_classification, remediation_template)
- [ ] Aho-Corasick keyword prefilter before regex
- [ ] Entropy + character-set scoring, placeholder detection, nearby-variable-name context, path-based downscoring; store candidate features (for §12)
- [ ] Skip lists (lockfiles, `vendor/`, `node_modules/`, binaries, generated assets) — configurable; skipped categories counted in coverage
- [ ] Windowed scanning for large/minified inputs; per-file size + time guards; documented max token length
- [ ] Decoders: base64, URL-encoding, string concatenation — each with a fixture
- [ ] Redaction tested end-to-end: no raw secrets in output or logs
- [ ] Minimal static classification table (AWS, GitHub, Stripe, Google, Firebase, Supabase, Slack, Twilio, SendGrid, private keys, JWTs, generic)
- [ ] Rules: initial set with positive + negative fixtures each
- **Acceptance:** every rule's fixtures pass (CI merge gate); scans a sample repo; lockfiles/vendored dirs skipped and counted; findings carry a classification; no raw secrets in output or logs (tested).

## M3 — Git history scanning

- [ ] Hardened clone per D-007 checklist (each item enforced + tested where testable)
- [ ] All-branch history walk with per-commit diffs; commit cap + optional `--since`; coverage of history depth in output
- [ ] HMAC fingerprint dedup across commits/renames/re-encodings (locations list)
- [ ] first-seen commit + author; still-present-at-HEAD
- **Acceptance:** planted secret in an old commit found and marked removed-but-in-history; same secret moved between files and re-encoded as base64 dedups to one finding; capped scans report how much history was skipped.

## M4 — Domain crawler, JS, hydration payloads, source maps

- [ ] SSRF-guarded fetcher (`engine/fetcher`): DNS resolve + blocklists, pinned-IP connections, redirect re-checks, caps, https-only
- [ ] SSRF regression matrix (metadata IPs, `0.0.0.0`, loopback, IPv6 loopback, IPv4-mapped IPv6, decimal/octal, redirect-to-private, DNS rebinding via controlled resolver)
- [ ] HTML + inline script extraction; `importmap`, `preload`, CSS `url()`, `data-*` config attributes
- [ ] Bundler-manifest chunk discovery (webpack runtime chunk map, Vite manifests)
- [ ] Hydration payloads (`__NEXT_DATA__`, `__NUXT__`, Astro islands); service workers, web workers
- [ ] Source maps: external via `sourceMappingURL`, inline; `sourcesContent` reconstruction; fetch original files when map lacks content
- [ ] Crawl budget enforcement (pages, chunks, JS bytes, time) + machine-readable coverage output
- [ ] Windowed scanning of minified bundles (4 MB single-line must not stall); string-concatenation decoding
- **Acceptance:** full SSRF matrix passes; planted secrets in a lazy chunk, in `__NEXT_DATA__`, and in `sourcesContent` all found on a local test site; budget limits enforced and emitted as coverage data; concatenation-split secret found.

## M5 — Headers, cookies, CORS, TLS (passive module)

- [ ] Module: security headers, cookie flags, CORS misconfiguration, TLS basics
- [ ] Remediation snippets (nginx, Express, Vercel) per finding
- **Acceptance:** correct findings with remediation against a local test site with deliberate misconfigurations.

## M6 — Full classification layer, safe checks, ownership verifier

- [ ] Ownership verifier: DNS TXT + token-file methods; refuse-without-proof logic
- [ ] Provider safe checks (Firebase RTDB/Firestore, Supabase PostgREST, Maps where safely testable) — exact request shapes documented in DECISIONS first
- [ ] Complete provider classification table
- [ ] Network-egress allowlist test for the verifier
- **Acceptance:** verdicts correct on fixtures; checks proven read-only (mocked providers); egress allowlist test passes; checks refuse to run unless ownership verified (tested with both real methods against a local test site and with a stub).

## M7 — API, queue, persistence, SSE, encryption at rest

- [ ] REST API (stdlib mux, D-009): submit scan, get scan, list scans
- [ ] Postgres job queue (`FOR UPDATE SKIP LOCKED`); worker pool invoking the engine
- [ ] SSE live progress (auth required on SSE endpoints; CSRF protection on non-GET; SameSite=Lax+ cookies)
- [ ] Persistence: scans, findings `(scan_id, fingerprint)`, redacted evidence; encryption at rest for evidence + tokens
- [ ] Recorded decisions: rate limiting (in-process limiter + per-user concurrent-scan cap at the queue), negative-cache TTL cleanup, encryption key source, scan retention/deletion, rescan-diff data model
- **Acceptance:** submit a scan, watch live progress, results persisted; failure and timeout paths tested; stored evidence/tokens verified encrypted; auth/CSRF requirements tested.

## M8 — Active probes + opt-in live verification

- [ ] Probes (`/.git`, `/.env`, backups) behind ownership verifier
- [ ] Opt-in live verification of secret-class keys (AWS `GetCallerIdentity`, GitHub `/user`) — off by default, labeled, ownership-gated
- [ ] Probe allowlist enforcement test
- **Acceptance:** probes and live verification blocked without proof, run with it; nothing outside the allowlist requested.

## M9 — Web UI

- [ ] New frontend scaffold (React + TypeScript + Tailwind, D-002)
- [ ] GitHub App sign-in + repo installation flow (per D-006)
- [ ] New scan (repo/domain), live progress, findings table with filters, finding detail, scan history
- [ ] Empty / loading / error states
- **Acceptance:** full flow works end to end locally.

## M10 — Report generation and export

- [ ] Executive summary + prioritized fix list
- [ ] Per-finding report rendering from `RedactedFinding` only
- [ ] Coverage statement rendering
- [ ] Rescan diff (fixed / new / still open)
- [ ] Exports: Markdown, PDF (on demand), SARIF 2.1.0 (validator chosen + recorded), JSON
- [ ] Expiring revocable share links (owner-only by default)
- [ ] Export canary test: planted fake secret absent from every MD/PDF/SARIF/JSON/API/share-link output
- **Acceptance:** reports generated for fixtures; SARIF validates against official 2.1.0 schema; diff correct across two scans; canary test passes.

## M11 — Hardening and demo

- [ ] Rate limits + egress caps wired end-to-end
- [ ] Key management + rotation plan implemented (fingerprint key, encryption keys)
- [ ] Security checklist in `docs/` all ticked
- [ ] README: architecture diagram, demo, deployment guide
- [ ] Demo scan recorded
- **Acceptance:** checklist complete; demo recorded.

---

## Post-MVP (design for, do not build)

Benchmark harness (precision/recall vs gitleaks/TruffleHog, ablation by layer) · dependency
scanning via OSV.dev · code-pattern rules · repo-to-production correlation · ML FP classifier.
