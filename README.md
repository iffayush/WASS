# WASS — Web Application Secret Scanner

WASS scans a **git repository** or a **deployed web app** for exposed secrets, exposed files, and
insecure configuration — and tells you not just *what* leaked, but *whether it is actually
dangerous*, with evidence and a specific fix for every finding.

## The thesis

Most secret scanners stop at "possible secret found." That is not good enough, because many
frontend keys are **public by design** (Firebase web config, Stripe publishable keys, Supabase
anon keys). WASS classifies every finding:

| Class | Meaning | Example |
|---|---|---|
| `public_by_design` | Safe if provider-side config is correct | Stripe `pk_…` |
| `dangerous_if_misconfigured` | Dangerous only if provider rules are open | Firebase DB with unauthenticated reads |
| `critical` | Dangerous regardless | AWS `AKIA…` secret keys, private keys |

…then proves or disproves exposure with **safe, read-only, ownership-gated** checks, and reports
a verdict: is this live, inactive, or unverifiable — with evidence.

## What it scans

- **Repo mode** — working tree + full git history (all branches): first-seen commit, author,
  still-present-at-HEAD, dedup across renames and re-encodings.
- **Domain mode** — HTML, inline scripts, lazy-loaded JS chunks (via bundler manifests), service
  workers, `importmap`, hydration payloads (`__NEXT_DATA__`, `__NUXT__`), and source maps
  (including reconstructing original source from `sourcesContent`) — within an explicit crawl
  budget. Exposed files (`/.git/`, `/.env`, backups), security headers, cookies, CORS, TLS.

Every report carries a **coverage statement**: exactly what was checked and what was not. WASS
never implies completeness.

## Status

**M0 complete** — scaffold, docs, CI. The engine lands in M1–M4; see
[docs/TASKS.md](docs/TASKS.md) for the milestone plan and [docs/SPEC.md](docs/SPEC.md) for the
product specification. Architecture: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md); decision log:
[docs/DECISIONS.md](docs/DECISIONS.md).

## Quick start (development)

Requirements: Go ≥ 1.22 (developed on 1.26), Docker (for Postgres), gitleaks (for the pre-commit
hook — fail-closed).

```sh
git config core.hooksPath .githooks   # enable the pre-commit secret scan (or: make hooks)
go test ./...                          # or: make test
go build -o bin/ ./cmd/wass            # or: make build
./bin/wass --version                   # (bin/wass.exe on Windows)
docker compose up -d                   # Postgres for the server (M7)
```

Windows users: if `make` is not installed, run the underlying commands directly (they are listed
in the [Makefile](Makefile)).

## Safety model (non-negotiable)

- **Ownership gating:** passive scanning only needs a URL. Active probes, provider
  misconfiguration checks, and live key verification require **proof of ownership** (DNS TXT or
  token file) — no proof, no probe.
- **SSRF protection** in every fetcher: private/loopback/link-local/cloud-metadata ranges blocked,
  redirect re-checks, DNS-rebinding defense, strict caps. Regression-tested.
- **Redaction boundary:** raw secret values never leave the engine. Reports, exports, API
  responses, share links, and the optional LLM layer handle only `RedactedFinding` (first 4 +
  last 4 characters + hash) — enforced by the type system.
- **Nothing scanned is ever executed.** No sandboxes, no building, no running untrusted code.
- Cloned repos are handled under a hardening checklist (hooks disabled, file protocol denied,
  https-only, size/time caps, scratch dirs deleted) — see [docs/DECISIONS.md](docs/DECISIONS.md) D-007.
- This repo practices what it scans: gitleaks runs in CI and in the pre-commit hook; fixtures use
  obviously fake values only.

## Repository layout

```
engine/    Go scanner core, modules, fetchers, CLI   (standalone — never imports server/frontend)
server/    Go API, queue workers, SSE, persistence
frontend/  Web UI (M9)
rules/     YAML secret rules + fixtures (M2)
docs/      SPEC · ARCHITECTURE · TASKS · DECISIONS · HANDOFF
```

## License

To be chosen before the first public release (tracked in docs/DECISIONS.md).
