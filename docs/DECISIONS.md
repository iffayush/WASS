# WASS — Decision Log

Format: date, decision, alternatives considered, reason. Newest at the bottom. Every significant
choice in the project gets an entry here.

---

## D-001 — Full repository reset (owner-authorized)

- **Date:** 2026-10-17
- **Decision:** The repository was reset from scratch at the owner's explicit instruction
  ("remove what's already present in the repo, I want fresh start"). Removed: the previous README
  (which described an unimplemented Docker-sandbox "Orchestrator Engine" — a design that
  contradicts WASS's non-goals: no sandboxes, no executing scanned code), the previous
  `frontend/` (a Next.js scaffold whose app layer POSTed to an n8n webhook at
  `localhost:5678/webhook/security-scan` and parsed LLM-style reports), and the local git history.
  WASS is now built solely from the master build prompt.
- **Alternatives:** keeping the frontend scaffold and dropping only the n8n integration (owner
  initially leaned this way, then chose full removal); preserving the old commit as audit trail
  (superseded by the owner's instruction).
- **Reason:** owner's explicit choice; nothing of the old content is unrecoverable — the GitHub
  remote (`github.com/iffayush/WASS`, commit `69b5c18`) still holds it if ever needed.
- **Consequence:** no git remote is attached. Re-adding `origin` and force-pushing the new
  history to `master` is an owner decision and requires explicit confirmation (it rewrites the
  public repository history).

## D-002 — Frontend starts empty; new UI in M9

- **Date:** 2026-10-17
- **Decision:** No frontend code exists in the repo after the reset (D-001). The M9 UI will be
  built from scratch as React + TypeScript + Tailwind, per the master prompt's stated preference.
- **Alternatives:** keep/reuse the old scaffold (rejected via D-001).
- **Reason:** the old scaffold's only non-boilerplate logic was the n8n integration, which is
  architecturally incompatible with the Go engine + Go API design.

## D-003 — Go version: 1.26.x, minimum 1.22

- **Date:** 2026-10-17
- **Decision:** Develop on Go **1.26.5** (the toolchain installed on this machine);
  `go.mod` declares `go 1.22` as the floor.
- **Alternatives:** pinning to 1.22 exactly.
- **Reason:** the master prompt requires ≥ 1.22 for per-iteration loop variables (critical for
  the parallel scanning); 1.26 is the current stable installed locally. The 1.22 floor keeps the
  build portable.

## D-004 — Go module path: `github.com/iffayush/wass`

- **Date:** 2026-10-17
- **Decision:** Module path is lowercase `github.com/iffayush/wass` regardless of the GitHub
  repository's casing (`WASS`).
- **Alternatives:** matching the repo casing exactly.
- **Reason:** lowercase is the Go convention and avoids import-path confusion; GitHub serves
  module paths case-insensitively.

## D-005 — Storage: Postgres via pgx v5, queue via `SKIP LOCKED`, no Redis

- **Date:** 2026-10-17
- **Decision:** Postgres for all persistence, accessed with **pgx v5** and `pgxpool` (not
  `lib/pq`, which is in maintenance mode). Scan jobs are queued in a Postgres table claimed with
  `SELECT … FOR UPDATE SKIP LOCKED`. Findings keyed by `(scan_id, fingerprint)` for rescan diffs.
- **Alternatives:** River (a Postgres-native job queue — the fallback if the hand-rolled queue
  proves insufficient; revisit in M7); Redis (rejected — one more moving part for zero MVP need).
- **Reason:** master prompt requirement; recorded here per its instruction to log the choice in
  M0.

## D-006 — Repo access & sign-in: GitHub App, not OAuth App

- **Date:** 2026-10-17
- **Decision:** Use a **GitHub App** (permissions: `metadata: read`, `contents: read`) for both
  sign-in (user authorization flow) and private-repo access (per-repo installations, read-only).
- **Alternatives:**
  - *OAuth App* — rejected: the classic `repo` scope grants read **and write** to everything the
    user can reach; no per-repo scoping.
  - *Fine-grained personal access token* — considered and rejected: it does offer per-repo
    read-only, but it puts token management on users (creation, expiry, revocation) and does not
    provide a sign-in flow; we would still need OAuth for identity. The GitHub App unifies
    sign-in + per-repo installation + revocation in one mechanism, and its installation UI is
    maintained by GitHub.
- **Reason:** least-privilege access with per-repo user control; the master prompt's requirement,
  with the fine-grained PAT alternative now explicitly on record.
- **Consequence:** M7/M9 build the app installation flow and token handling; stored tokens are
  encrypted at rest from the first milestone that persists them (M7).

## D-007 — Untrusted-repo clone hardening checklist

- **Date:** 2026-10-17
- **Decision:** Every clone performed by the engine enforces, in code:
  1. `GIT_TERMINAL_PROMPT=0` (no credential prompts);
  2. hooks disabled — `core.hooksPath` pointed at an empty directory;
  3. `protocol.file.allow=never` and https-only remote allowlist;
  4. no submodule recursion by default;
  5. repo-local config includes ignored (`include`, `includeIf` never honored from the cloned
     config — use `--no-config`-equivalent flags / a scrubbed environment);
  6. partial clones for large repos (`--filter=blob:none`) or `--depth` where history scanning
     allows;
  7. size and time limits on clone and on history walking (per-repo commit cap, optional
     `--since`);
  8. scratch directories under a per-scan temp root, deleted after the scan;
  9. nothing from the repo is ever executed.
- **Windows adaptation (this dev machine):** item "run git as a dedicated unprivileged user" is
  recorded as a **deployment requirement** (Linux server / container in M11); on Windows dev the
  config-level hardening above applies and no scanned content is executed. The master prompt's
  checklist assumed Linux; this entry adapts it without weakening it.
- **Alternatives:** none material.
- **Reason:** master prompt M0 requirement.

## D-008 — Fingerprint key handling: flag or env, versioned, no default

- **Date:** 2026-10-17
- **Decision:** The engine CLI takes the HMAC key from `--fingerprint-key` or
  `WASS_FINGERPRINT_KEY`; there is **no built-in default key** (a missing key is a hard error for
  fingerprint-producing scans). The server passes its own key so CLI and server fingerprints
  agree. The key **version** is stored alongside every fingerprint. Rotating the key invalidates
  stored fingerprints and rescan diffs → the rotation plan (M11) includes a re-fingerprint
  migration.
- **Alternatives:** a built-in default key (rejected: silent cross-installation fingerprint
  agreement is a footgun and a weak-secret risk).
- **Reason:** master prompt requirement (decide and record in M1 — recorded now while M0 writes
  the schema spec it affects).

## D-009 — API router: standard library first

- **Date:** 2026-10-17
- **Decision:** The server uses Go 1.22+ `net/http` ServeMux (method + path patterns).
- **Alternatives:** chi (revisit if middleware needs — auth, CSRF, rate limiting, SSE — outgrow
  hand-rolled middleware; SSE works fine on stdlib).
- **Reason:** master prompt allows either; stdlib minimizes dependencies, and every dependency
  must be justified. Adding chi later is cheap; removing it is not.

## D-010 — Development environment reality (Windows)

- **Date:** 2026-10-17
- **Decision:** Development happens on Windows (Git Bash). Toolchain verified at M0: Go 1.26.5,
  Docker 28.2.2, Git 2.45.2. Not installed locally: `make`, `gitleaks`, `golangci-lint` /
  `staticcheck`.
  - The `Makefile` is kept as the canonical entry point (CI and Linux/macOS); on Windows the
    documented fallbacks are the plain commands (`go test ./...`, `go vet ./...`) — see README.
  - The pre-commit hook is **fail-closed**: if `gitleaks` is not installed it aborts the commit
    with install instructions (`winget install Gitleaks.Gitleaks` or `scoop install gitleaks`).
  - CI runs gitleaks and golangci-lint via their official GitHub Actions, so local absence does
    not weaken the merge gate.
- **Alternatives:** installing make/gitleaks locally now (owner can do so any time; not required
  for M0 acceptance).
- **Reason:** honesty in the acceptance criteria: `make test` must work where `make` exists;
  Windows developers get the equivalent commands.

## D-011 — Environment hazard flagged (out-of-repo)

- **Date:** 2026-10-17
- **Decision:** Flagged to the owner, not fixed by the agent: a stray git repository exists at
  `C:\Users\Acer\.git` (created by "Lens Studio", no commits, remote pointing at the old WASS
  repo). Before WASS got its own `.git`, any `git add` run anywhere under the home directory
  would have staged the entire home directory. WASS now has its own repository, so it is
  unaffected — but the home-level repo remains a hazard for every other project under
  `C:\Users\Acer`.
- **Alternatives:** deleting `C:\Users\Acer\.git` unilaterally (rejected: not ours to remove;
  a tool may depend on it).
- **Reason:** owner should decide whether to remove it.
