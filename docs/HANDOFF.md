# Handoff

Updated: 2026-10-17 by ZCode (GLM-5.3)
Current milestone: **M1 complete** — next is **M2 (secrets rules engine, working tree)**

## Done this session

M1 (schema and engine skeleton), five commits:

- `engine/finding` — versioned common finding schema (SchemaVersion 1) with all SPEC §4 fields;
  redaction boundary implemented as designed in D-012: raw values live only in the internal
  `Candidate` (whose `MarshalJSON` hard-fails), `FromCandidate` is the one-way constructor
  (HMAC-SHA256 fingerprint of the normalized raw value + evidence redaction), `RedactedFinding`
  is the public projection. Redaction = first4+last4 (≤12 chars fully masked). Base64-unwrap
  normalization per D-013 (strict, symmetric — raw and wrapped forms fingerprint identically).
  19 unit tests.
- `engine` — `Module` interface (`Name` + `Scan(ctx, target, emit)`; modules emit candidates
  only) and `Scanner` core: converts candidates behind the boundary, dedups on
  `(rule_id, fingerprint)`, merges unique locations, assigns IDs, validates every finding,
  records modules run + timing. `Report.Redacted()` → `RedactedReport`. Per D-014.
- `engine/stubmod` — placeholder module scanning `WASSFAKE-[A-Z0-9]{32}` over a repo working
  tree: skip lists (.git/node_modules/vendor), binary sniff, per-file size guard, capped
  context snippets; generated fixtures under `testdata/fixture` (positive, negative, binary).
- `cmd/wass` — real CLI: `scan --repo <path> [--pretty] [--timeout]`, key from
  `--fingerprint-key` or `WASS_FINGERPRINT_KEY` (hard error without either, D-008),
  `--domain` rejected with an M4 pointer, JSON `RedactedReport` on stdout / diagnostics on
  stderr. End-to-end tests include "raw fake value absent from CLI output".
- `.gitleaks.toml` — allowlist scoped to `testdata/` only (D-015), so fake fixtures pass our
  own fail-closed hook; verified live with `gitleaks detect` before committing.

Verified live: `wass scan --repo engine/stubmod/testdata/fixture --fingerprint-key demo-key`
emits one finding with two deduped locations, evidence `"api_key = \"WASS…G5H6\""`, no raw
value in output. All acceptance criteria met.

## In progress

Nothing — clean stopping point.

## Next steps (ordered, concrete)

1. **M2 task 1:** YAML rule format + loader in `engine/rules` (id, description, regex,
   keywords, entropy threshold, provider, default_classification, remediation_template) with
   validation tests. Add `golangci-lint` to CI (first M2 task in TASKS.md, D-010).
2. Keyword prefilter (Aho-Corasick — evaluate `cloudflare/ahocorasick` vs hand-rolled; justify
   the dependency in DECISIONS if external).
3. Context scoring: entropy, charset mix, nearby variable names, path downscoring
   (test/docs/example), placeholder detection. **Store candidate features per candidate**
   (SPEC §12 needs them later).
4. Skip lists (configurable; lockfiles, vendor/, node_modules/, binaries, generated assets) —
   each skipped category counted for the coverage statement.
5. Windowed scanning for minified/large inputs + per-file size/time guards + documented max
   token length.
6. Decoders: base64, URL-encoding, string concatenation (`"AKIA"+"IOSFODNN7"+"EXAMPLE"`), one
   fixture each.
7. Static classification table (Stripe pk_ public / sk_ critical, AWS critical, Firebase
   dangerous_if_misconfigured, etc.).
8. Initial rule set: AWS, GitHub, Stripe, Google, Firebase, Supabase, Slack, Twilio, SendGrid,
   private keys, JWTs, generic high-entropy — every rule with positive + negative fixtures
   (CI merge gate).
9. Wire the secrets module into the CLI scanner; no-raw-secrets-in-output-or-logs test.

## Decisions made this session

D-012 (candidate/tripwire design, ≤12-char full mask), D-013 (strict base64 normalization),
D-014 (modules emit candidates; scanner owns the boundary), D-015 (testdata-scoped gitleaks
allowlist) — all in [docs/DECISIONS.md](DECISIONS.md).

## Open questions for the owner

Unchanged from M0, minus the remote question (resolved 2026-10-17: `origin`
re-added and the new history force-pushed over `master`; old content only in
reflogs/GitHub support):

1. **`C:\Users\Acer\.git`** stray home repo (D-011) — still flagged for the owner.
2. **License** — still unchosen.

## Gotchas

- **gitleaks PATH** in pre-2026-10-17 shells: `export PATH="$PATH:$HOME/AppData/Local/Microsoft/WinGet/Links"`.
- **No `make` locally** — use the plain `go` commands (D-010).
- gitleaks needs `.gitleaks.toml` (committed M1) — fixtures under `testdata/` are allowlisted;
  any new fake-secret fixture must live under a `testdata/` path or the hook will block.
- CLI test helper swaps `os.Stdout`/`os.Stderr` via pipes — fine for small outputs; don't
  print >64 KiB from cliMain in tests.
- CRLF conversion warnings from git are expected (`.gitattributes` normalizes to LF).

## How to verify (confirm the current state is green)

```sh
cd "/c/Users/Acer/Desktop/Coding Projects/Go/WASS"
go vet ./... && go test ./...        # all packages ok
go build -o bin/ ./cmd/wass
./bin/wass.exe scan --repo engine/stubmod/testdata/fixture --fingerprint-key demo-key
# → JSON with findings_count 1, two locations, evidence "WASS…G5H6", no raw value
./bin/wass.exe scan --repo .          # → hard error: fingerprint key required (D-008)
export PATH="$PATH:$HOME/AppData/Local/Microsoft/WinGet/Links"
gitleaks detect --no-banner --redact  # no leaks
git log --oneline                     # 12 commits, conventional messages
git status                            # clean
```
