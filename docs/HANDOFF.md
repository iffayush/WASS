# Handoff

Updated: 2026-10-17 by ZCode (GLM-5.3-Flash)
Current milestone: **M0 complete** — next is **M1 (schema and engine skeleton)**

## Done this session

- Analyzed the master build prompt against the actual (pre-existing) repo state; owner then
  ordered a **full reset**: old README (Docker-orchestrator design, contradicted the non-goals),
  old frontend (n8n-webhook client), and old git history all removed. WASS derives solely from
  the master prompt. Old content remains on the GitHub remote (`github.com/iffayush/WASS`,
  commit `69b5c18`) if ever needed (D-001, D-002).
- Fresh git repository initialized **inside the project folder** (previously git resolved to a
  stray home-directory repo at `C:\Users\Acer\.git` — see D-011).
- Toolchain verified: Go 1.26.5, Docker 28.2.2, Git 2.45.2; `make`/`golangci-lint` not installed
  (documented fallbacks, D-010). **gitleaks 8.30.1 installed via winget** so the fail-closed
  pre-commit hook works.
- All M0 artifacts committed (7 commits, conventional messages): `AGENTS.md` + `CLAUDE.md` +
  `.cursor/rules/wass.mdc`; `docs/SPEC.md`, `ARCHITECTURE.md`, `DECISIONS.md` (D-001…D-011),
  `TASKS.md`, `HANDOFF.md`; `go.mod` (module `github.com/iffayush/wass`, floor go 1.22) with
  `engine/`, `server/`, `cmd/wass/` stubs + passing tests; `rules/README.md`; `Makefile`,
  `docker-compose.yml` (Postgres 17), `.github/workflows/ci.yml` (test matrix 1.22.x+stable,
  vet, gitleaks), `.githooks/pre-commit` (fail-closed), `.gitattributes`, `.gitignore`, README.

## In progress

Nothing — clean stopping point.

## Next steps (ordered, concrete)

1. **M1 task 1:** define the versioned common finding schema in `engine/` (all fields per
   `docs/SPEC.md` §4) with unit tests.
2. Implement `RedactedFinding` + one-way `redact()` (first4+last4+hash); test raw values cannot
   leak through the boundary.
3. Define the module interface; scanner core that runs modules and collects findings.
4. Fingerprinting: HMAC-SHA256 of normalized raw value, versioned key, **hard error without key**
   (D-008).
5. CLI: `wass scan --repo <path> --fingerprint-key …` emitting JSON; run against a fixture
   directory with a stub module.
6. M1 acceptance: unit tests for schema and redaction pass; CLI works on fixtures. Update
   TASKS/HANDOFF per task and commit.

## Decisions made

All in [docs/DECISIONS.md](DECISIONS.md): D-001 full reset (owner-authorized) · D-002 frontend
starts empty, new UI in M9 · D-003 Go 1.26.x / floor 1.22 · D-004 module path lowercase ·
D-005 pgx v5 + SKIP LOCKED queue, no Redis · D-006 GitHub App (fine-grained PAT rejected) ·
D-007 clone-hardening checklist with Windows adaptation · D-008 fingerprint key handling (no
default) · D-009 stdlib mux first · D-010 Windows dev reality (make/lint deferred) · D-011
home-dir git repo hazard flagged.

## Open questions for the owner

1. **Remote:** no `origin` is configured (fresh start). Re-adding
   `https://github.com/iffayush/WASS.git` and **force-pushing** the new history over the old
   `master` rewrites public history — say the word and it will be done (or point at a new repo).
2. **`C:\Users\Acer\.git`** (stray home-directory repo from "Lens Studio", D-011): recommend
   removing it, but it is not ours to delete.
3. **License:** none chosen yet — needed before the first public release (noted in README).
4. **Install `make` locally?** Optional; until then use the plain `go` commands (D-010).

## Gotchas

- **gitleaks PATH:** winget added it to the user PATH; shells opened *before* install don't see
  it. Fix: open a new terminal, or
  `export PATH="$PATH:$HOME/AppData/Local/Microsoft/WinGet/Links"`.
- **No `make` on this machine** — use the plain commands from the Makefile header (D-010).
- **golangci-lint intentionally deferred to M2** (nothing to lint at M0; `go vet` runs in CI).
- CRLF→LF conversion warnings on `git add` are expected (`.gitattributes` normalizes; the repo
  stores LF).
- Windows + Git Bash: quote the project path (it contains a space).
- CI's gitleaks step uses `fetch-depth: 0`; keep it or full-history scanning breaks.

## How to verify (confirm the current state is green)

```sh
cd "/c/Users/Acer/Desktop/Coding Projects/Go/WASS"
go vet ./...                      # clean
go test ./...                     # engine + server packages: ok
go build -o bin/ ./cmd/wass && ./bin/wass.exe --version   # wass 0.0.1-dev
docker compose config --quiet     # valid compose file (Docker Desktop must be installed)
git config core.hooksPath         # .githooks
git log --oneline                 # 7 conventional commits
git status                        # clean
export PATH="$PATH:$HOME/AppData/Local/Microsoft/WinGet/Links"
gitleaks detect --no-banner --redact   # no leaks found
```
