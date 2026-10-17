# WASS — Product Specification

Status: v1 (M0). Source of truth for what WASS does and does not do. Architecture lives in
`docs/ARCHITECTURE.md`; the running decision log in `docs/DECISIONS.md`.

## 1. Product

A user pastes a **domain** or a **git repository URL** into a web app. WASS scans it, finds exposed
secrets, exposed files, and insecure configuration, then produces a prioritized **report with
specific remediation** for each finding.

**Thesis.** Do not just say "possible secret found." Answer *"is this exposed value actually
dangerous, and what is the evidence?"* Many frontend keys are public by design. The real findings
are secret-class keys, or public-class keys whose provider-side configuration makes them
dangerous. Every finding gets a classification, a verification verdict, and evidence.

### Classification model

| Class | Meaning | Examples |
|---|---|---|
| `public_by_design` | Safe if provider config is correct | Stripe `pk_…`, Firebase web config, Supabase `anon` key (with RLS), Maps key with referrer restrictions |
| `dangerous_if_misconfigured` | Dangerous only if provider-side rules are open | Firebase RTDB with unauthenticated reads, Supabase table readable by `anon`, Maps key without restrictions |
| `critical` | Dangerous regardless of provider config | AWS `AKIA…` secret keys, Stripe `sk_…`, private keys, GitHub PATs, Slack tokens, JWT signing secrets |
| `unknown` | Could not classify | Generic high-entropy strings with no provider match |

### Verification verdicts

`not_attempted` (default), `verified_live` (key confirmed working — opt-in only),
`verified_inactive` (key confirmed dead/revoked), `unverifiable` (provider provides no safe check).

## 2. Scope

### In scope (MVP)

- **Repo mode:** secrets in the working tree and full git history (all branches). Per finding:
  first-seen commit, author, still-present-at-HEAD. History caps: per-repo commit cap, optional
  `--since`; the report states exactly how much history was and was not scanned.
- **Domain mode:** crawl within an explicit **crawl budget** (max pages, max chunks, max total JS
  bytes, max time). Coverage:
  - HTML, inline scripts, lazy-loaded JS chunks discovered from bundler manifests (webpack runtime
    chunk map, Vite manifests) — not only from HTML;
  - service workers, web workers, `importmap`, `<link rel=preload>`, CSS `url()` references,
    `data-` attributes carrying config;
  - framework hydration payloads: `__NEXT_DATA__`, `__NUXT__`, Astro islands;
  - source maps: external `.map` via `sourceMappingURL` and inline maps; reconstruct original
    source from `sourcesContent`; if absent, try fetching the original files if served.
  - The coverage statement must list what the crawl did **not** cover (machine-readable coverage
    data in the scan output).
- **Exposed files** (domain mode): `/.git/`, `/.env`, backup and config files.
- **Headers & cookies:** security headers, cookie flags, CORS misconfiguration, TLS basics.
- **Classification:** per-provider table (static in M2; safe live checks in M6).
- **Report:** prioritized summary, per-finding explanation and fix, coverage statement, rescan with
  diff, export to Markdown / PDF / SARIF 2.1.0 / JSON, expiring revocable share links.
- **Web UI** with GitHub App auth.

### Explicit non-goals (do not build; ask the owner before adding)

- Building or executing scanned code; sandboxes; Docker-in-Docker; running untrusted repos.
- Nikto/ZAP-style general vulnerability scanning.
- Dependency vulnerability scanning and insecure-code-pattern rules (design plugin interfaces now,
  implement later).
- Fine-tuned ML (§12).

## 3. Users & access model

- Public git repos: no GitHub access needed at all.
- Private repos: only through the user's own **GitHub App** installation (permissions
  `metadata: read`, `contents: read`), chosen per repository, read-only.
- Domains: passive scanning needs no proof; active probes and provider checks need **ownership
  proof** (DNS TXT record with a per-scan token, or a token file served from the domain).

## 4. Common finding schema

Versioned, tested, and defined before any module. Every finding has at least:

| Field | Notes |
|---|---|
| `id` | Unique per scan. |
| `fingerprint` | HMAC-SHA256 with a configurable **versioned** key of the **normalized raw value** (decoded if base64-wrapped, whitespace-trimmed) — never of the redacted form. No built-in default key: CLI takes `--fingerprint-key` or `WASS_FINGERPRINT_KEY`; the server passes its own so CLI and server fingerprints agree. Key version stored alongside every fingerprint. Rotating the key invalidates stored fingerprints and rescan diffs → migration plan required (D-008). |
| dedup | On `(rule_id, value_hash)`; multiple **locations** attach to one finding (same secret across 40 commits, renames, or files = one finding). A rotated key is a new value = new finding. |
| `module` | `secrets`, `exposed_files`, `headers_cookies`, … |
| `type`, `title` | Machine type + human title. |
| `severity` | `critical`, `high`, `medium`, `low`, `info`. |
| `confidence` | 0–1. |
| `classification` | See §1 table. |
| `verification` | See §1 verdicts. |
| `location` | File or URL, line, commit, branch, `first_seen`, `still_present`. |
| `evidence` | Redacted snippet with context. Raw values are never stored. |
| `impact` | Plain-language explanation. |
| `remediation` | Steps + config snippets or commands (provider rotation steps; `git filter-repo`/BFG commands; nginx/Express/Vercel header configs). |
| `references` | Provider docs, CWE where applicable. |

**Redaction boundary (compile-time enforced):** the report renderer, every exporter, API
responses, share links, and the LLM client accept **only** `RedactedFinding` — never the internal
`Finding`. Leaking a raw value becomes a compile error, not a code-review hope. Redaction = first
4 + last 4 characters plus a hash.

## 5. Secrets engine (the core)

- **Rules** in a WASS YAML format: `id`, `description`, `regex`, `keywords`, optional `entropy`
  threshold, `provider`, `default_classification`, `remediation_template`. Every rule ships with
  positive and negative **fixtures** and must pass its tests (CI merge gate from M2).
- **Fast path:** Aho-Corasick keyword prefilter before regex.
- **Skip lists (biggest FP lever):** lockfiles (`package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`,
  `go.sum`, `Cargo.lock`, `poetry.lock`), `vendor/`, `node_modules/`, binaries, known
  generated/minified assets — repo mode only. Domain mode scans minified bundles on purpose
  (rule precision + context scoring carry the load there). Skip lists are configurable; every
  skipped path category is counted in the coverage statement.
- **Context scoring:** entropy, character-set mix, nearby variable names (`api_key`, `secret`,
  `token`), file path (test/docs/example files score lower), placeholder detection
  (`YOUR_KEY_HERE`, `xxxx`, repeated characters). Candidate features are **stored per candidate**
  (§12 needs them later).
- **Decoding:** base64, URL-encoding, and simple string concatenation in minified JS
  (`"AKIA"+"IOSFODNN7"+"EXAMPLE"`), each with a fixture.
- **Large/minified inputs:** scan in windows with overlap (not line-by-line); documented maximum
  line/token length; per-file input-size and time guards on regex execution.
- **Baseline/suppress mechanism** for accepted findings.
- **Initial rule set (quality over quantity):** AWS, GitHub, Stripe, Google, Firebase, Supabase,
  Slack, Twilio, SendGrid, private keys, JWTs, generic high-entropy assignments.

## 6. Classification & verification

- **M2:** static provider table ships (shape → classification), so every early finding already
  carries a classification and the thesis is visible in the first demo.
- **M6:** safe, read-only misconfiguration checks for `dangerous_if_misconfigured` classes:
  - Firebase: does the RTDB/Firestore allow unauthenticated reads?
  - Supabase: is a table readable with the `anon` key?
  - Maps: does the key lack referrer restrictions (only where safely testable)?
  Each check is defined precisely in DECISIONS: exact request shape (single `GET` limited to one
  record, or `HEAD`), what response proves exposure, and what is discarded immediately. Never call
  a write endpoint; never retrieve data beyond proving access exists (status code / record count,
  never contents).
- **Live verification of secret-class keys** (AWS `GetCallerIdentity`, GitHub `/user`): opt-in,
  read-only, off by default, clearly labeled, only for targets with proven ownership.
- **Gating:** all provider checks (misconfiguration or live) run only after the ownership
  verifier (DNS TXT or token file) reports verified. Without proof: finding is reported with
  `classification: dangerous_if_misconfigured`, `verification: not_attempted`, plus self-service
  instructions for the owner.
- **Tests:** network-egress allowlist test — the verifier may contact only allowlisted hosts/paths;
  the test fails on any other outbound request. Mocked providers alone are not sufficient.

## 7. Report (the product)

- Executive summary: prioritized "fix these first, and why".
- Per finding: title, severity, classification + verdict, evidence, plain-language impact,
  specific fix (provider rotation steps; `git filter-repo`/BFG command for history; ready-to-paste
  nginx/Express/Vercel config for header issues).
- **Coverage statement:** exactly what was checked and what was not ("active probes skipped:
  ownership not verified"), history depth scanned, crawl budget limits hit, skipped content
  categories. Never imply completeness.
- **Rescan with diff:** fixed / new / still open (findings keyed by `(scan_id, fingerprint)`).
- **Export:** Markdown, PDF (generated on demand from stored `RedactedFinding` data, not stored as
  files), SARIF 2.1.0 (validated against the official schema; no raw snippets in
  `region.snippet`), JSON, expiring revocable share links (owner-only by default — file paths and
  surrounding code are a disclosure surface even redacted).
- Optional LLM may polish prose and propose patches, clearly marked as suggestions; it can never
  set `severity`, `classification`, or `verification`, and cannot create findings.

## 8. Optional LLM layer

Behind an interface with a provider switch (Gemini, Claude, local via Ollama). Input: only the
structured fields of `RedactedFinding` — no free text from the target (page content, commit
messages, file contents), so scanned content cannot prompt-inject it. Output: may populate only
`report_prose` and `suggested_patch`. Everything works without it.

## 9. Milestones

See `docs/TASKS.md` for the full breakdown with acceptance criteria (M0–M11).

## 10. Safety model (summary — full rules in AGENTS.md and DECISIONS.md)

Tiering (passive vs active), SSRF protection with regression matrix, limits and caps everywhere,
untrusted-clone hardening, redaction-before-storage with encryption at rest, fake fixtures only,
gitleaks in CI and pre-commit.

## 11. Storage model (summary)

Postgres via pgx v5 (`pgxpool`). Postgres-backed job queue (`SELECT … FOR UPDATE SKIP LOCKED`
table; River is the fallback if the hand-rolled queue proves insufficient). No Redis.
Findings keyed by `(scan_id, fingerprint)` for rescan diffs. Stored evidence and OAuth tokens
encrypted at rest (key from environment for MVP; rotation plan in M11).

## 12. Later: ML layer (design for, do not build)

A false-positive classifier can slot in: candidate features (entropy, length, variable name, path,
rule id, surrounding tokens) are computed and stored per candidate from M2. Baseline
gradient-boosted or logistic model first, exported for plain-Go inference; fine-tuned code models
are a later experiment kept only if the benchmark shows real gain. Split train/test by repository.

## 13. Post-MVP (design for, do not build yet)

Benchmark harness (precision/recall vs gitleaks and TruffleHog, ablation by layer), dependency
scanning via OSV.dev, code-pattern rules, repo-to-production correlation (did a committed secret
reach the live bundle, and when), ML false-positive classifier.
