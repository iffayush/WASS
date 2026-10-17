# WASS — Agent Instructions (canonical)

This file is the single source of truth for any AI agent or human working on WASS.
`CLAUDE.md` and `.cursor/rules/wass.mdc` only point here. If rules conflict, this file wins,
then `docs/DECISIONS.md`.

WASS (**W**eb **A**pplication **S**ecret **S**canner) finds exposed secrets, exposed files, and
insecure configuration in a git repository or on a deployed web app, and produces a prioritized
report with specific remediation for every finding.

## The thesis (do not lose it)

Do not just say "possible secret found." Answer *"is this exposed value actually dangerous, and
what is the evidence?"* Many frontend keys are public by design (Firebase config, Stripe
publishable keys, Supabase anon keys, Maps keys). The real findings are secret-class keys, or
public-class keys whose configuration makes them dangerous. **Every finding carries a
classification and a verdict.**

- `public_by_design` — safe if the provider-side config is correct (Stripe `pk_`, Firebase web config)
- `dangerous_if_misconfigured` — dangerous only if provider-side rules are open (Firebase RTDB, Supabase RLS, Maps key restrictions)
- `critical` — dangerous regardless (AWS `AKIA…`, Stripe `sk_`, private keys, GitHub tokens)

## Scope

**Repo mode:** working tree + full git history (all branches). Report first-seen commit, author,
and whether the secret is still present at HEAD. History scanning has caps (commit cap, optional
`--since`); the report states exactly how much history was scanned and how much was not.

**Domain mode:** crawl within an explicit budget (max pages, max chunks, max total JS bytes, max
time). Cover: HTML, inline scripts, lazy-loaded JS chunks discovered from bundler manifests
(webpack runtime chunk map, Vite manifests) — not only from HTML — service workers, web workers,
`importmap`, `<link rel=preload>`, CSS `url()`, config-carrying `data-` attributes, and framework
hydration payloads (`__NEXT_DATA__`, `__NUXT__`, Astro islands). Include source maps (external via
`sourceMappingURL` and inline), reconstructing original source from `sourcesContent`; if a map has
no `sourcesContent`, try fetching the original files if served. The report's coverage statement
must list what the crawl did NOT cover.

**Also in scope (domain mode):** exposed files (`/.git/`, `/.env`, backups, configs), security
headers, cookie flags, CORS misconfiguration, TLS basics.

**Report:** prioritized summary, per-finding explanation and fix, coverage statement, rescan with
diff (fixed / new / still open), export to Markdown, PDF, SARIF 2.1.0, JSON, plus expiring
revocable share links.

## Non-goals (do not build; ask the owner before adding anything here)

- Building or executing scanned code. No sandboxes, no Docker-in-Docker, no running untrusted repos.
- Nikto/ZAP-style general vulnerability scanning.
- Dependency vulnerability scanning and insecure-code-pattern rules (design plugin interfaces for them; do not implement).
- Fine-tuned ML (see docs/SPEC.md §12 — design the feature pipeline, do not train anything).

## Hard safety rules

1. **Tiering.** Passive scans (only what a normal browser visit requests) need no proof. Active
   probes (`/.git`, `/.env`, backups), provider misconfiguration checks (Firebase, Supabase, Maps),
   and live verification of secret-class keys require **ownership proof** (DNS TXT or token file on
   the domain). Public repos may be scanned; private repos only via the user's own GitHub App grant.
2. **SSRF protection in every fetcher.** Resolve DNS, block private/loopback/link-local/cloud
   metadata ranges (`169.254.169.254`), re-check after redirects, pin the resolved IP for the
   connection (DNS rebinding defense), cap redirects/response size/time, reject non-http(s)
   schemes. Regression matrix lives in `engine/fetcher` tests: metadata IPs, `0.0.0.0`, loopback,
   IPv6 loopback, IPv4-mapped IPv6, decimal/octal encodings, redirect-to-private, rebinding.
3. **Limits everywhere:** per-scan timeout, page/size caps, crawl budget, per-user rate limits and
   concurrent-scan caps, per-scan outbound request cap, global egress bandwidth cap, negative cache
   for hosts that repeatedly time out.
4. **Cloning untrusted repos** (full checklist in `docs/DECISIONS.md` D-007):
   `GIT_TERMINAL_PROMPT=0`; hooks disabled; `protocol.file.allow=never`; https-only remotes;
   no submodule recursion; partial clones for large repos; size and time limits; scratch dirs
   deleted after scan; nothing from the repo is ever executed.
5. **Secrets handling:** redact before storage (first 4 + last 4 chars + hash, never the raw
   value); encrypt stored evidence and tokens at rest; scrub logs; never send raw secrets to any
   LLM or third party. The `RedactedFinding` type boundary is enforced in code: report renderers,
   exporters, API responses, share links, and the LLM client accept **only** `RedactedFinding`.
6. **Fixtures use obviously fake generated values.** Never commit real credentials. CI runs a
   secret scan (gitleaks) on every PR; the pre-commit hook runs the same check. No default
   fingerprint HMAC key exists — it comes from `--fingerprint-key` or `WASS_FINGERPRINT_KEY`.

## Common finding schema (summary — full version in docs/SPEC.md)

Every finding: `id`, stable `fingerprint` (HMAC-SHA256 of the **normalized raw value** with a
versioned key — never of the redacted form; dedup on `(rule_id, value_hash)` with a location
list), `module`, `type`, `title`, `severity` (critical/high/medium/low/info), `confidence` (0–1),
`classification`, `verification` (`not_attempted`/`verified_live`/`verified_inactive`/
`unverifiable`), `location` (file/URL, line, commit, branch, first_seen, still_present),
`evidence` (redacted snippet), `impact`, `remediation` (steps + config snippets/commands),
`references`.

## Repo layout

```
wass/
  engine/            # Go: scanner core, modules (secrets, exposed-files, headers), fetchers, CLI
  server/            # Go: API, queue workers, SSE, persistence
  frontend/          # web UI (built in M9)
  rules/             # YAML secret rules + fixtures
  benchmark/         # corpus and scoring harness (post-MVP)
  docs/              # SPEC, ARCHITECTURE, TASKS, DECISIONS, HANDOFF
  AGENTS.md  CLAUDE.md  .cursor/rules/wass.mdc
```

Engine never imports server or frontend. Server never imports engine internals — only the
finding schema and redaction types.

## Working rules for every session

- One milestone at a time. Do not start a milestone until the previous one's acceptance criteria
  pass (see `docs/TASKS.md`). Do not expand scope; if something seems necessary but is outside the
  current milestone or the non-goals, **stop and ask**, and record the question in `docs/HANDOFF.md`.
- Small conventional commits (`feat(engine): …`, `fix(server): …`, `docs: …`). One task, one commit.
- Write tests with the code. A rule is not done until its fixtures pass. Fixtures passing is a CI
  merge gate from M2.
- Record every significant choice in `docs/DECISIONS.md` (date, decision, alternatives, reason).
  Justify every new dependency there. Prefer the standard library.
- Keep functions small and typed; handle errors explicitly; no panics on untrusted input.
- Never claim WASS beats other scanners unless the benchmark harness proves it.
- **Before ending any session: update `docs/TASKS.md` and rewrite `docs/HANDOFF.md`** (rewrite,
  not append — template inside). Update HANDOFF after each completed task, not only at the end,
  because sessions can be cut off at any time.

## Session bootstrap (paste at the start of every new session)

1. Read `AGENTS.md`, `docs/HANDOFF.md`, `docs/TASKS.md`, `docs/DECISIONS.md`.
2. Run the verification commands listed in HANDOFF and report whether the repo is green.
3. State the current milestone and the next task, in two lines.
4. Work the next unchecked task only. Update TASKS + HANDOFF and commit after each task.
5. If low on context: stop at a clean point, update HANDOFF first, and tell the owner what to
   paste next session.
