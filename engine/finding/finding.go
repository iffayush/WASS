// Package finding defines the WASS common finding schema (versioned), the
// redaction boundary, and fingerprinting.
//
// # The redaction boundary
//
// Raw secret values exist in exactly one type: [Candidate], which is
// engine-internal input to [FromCandidate]. FromCandidate is the one-way
// constructor: it fingerprints the normalized raw value, redacts it out of
// the evidence snippet, and returns a [Finding] that carries no raw value at
// all. [RedactedFinding] is the public projection that report renderers,
// exporters (Markdown/PDF/SARIF/JSON), API responses, share links, and the
// LLM client accept — never Finding, never Candidate. Passing a Finding
// where a RedactedFinding is required is a compile error by construction.
//
// Fingerprints are HMAC-SHA256 (hex) of the normalized raw value under a
// caller-supplied, versioned key. There is no default key: an empty key is a
// hard error (DECISIONS.md D-008).
package finding

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// SchemaVersion is the version of the common finding schema emitted by this
// build. It appears in every scan envelope and changes only with a
// documented migration.
const SchemaVersion = 1

// Severity ranks how urgently a finding needs attention.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Classification is the thesis of the product: what kind of danger this
// exposed value actually poses.
type Classification string

const (
	// ClassificationPublicByDesign: safe if provider-side config is correct
	// (Stripe pk_ keys, Firebase web config).
	ClassificationPublicByDesign Classification = "public_by_design"
	// ClassificationDangerousIfMisconfigured: dangerous only if provider-side
	// rules are open (Firebase RTDB reads, Supabase RLS, Maps restrictions).
	ClassificationDangerousIfMisconfigured Classification = "dangerous_if_misconfigured"
	// ClassificationCritical: dangerous regardless of provider config
	// (secret-class keys, private keys).
	ClassificationCritical Classification = "critical"
	// ClassificationUnknown: could not be classified.
	ClassificationUnknown Classification = "unknown"
)

// Verification is the live-check verdict for a finding.
type Verification string

const (
	VerificationNotAttempted     Verification = "not_attempted"
	VerificationVerifiedLive     Verification = "verified_live"
	VerificationVerifiedInactive Verification = "verified_inactive"
	VerificationUnverifiable     Verification = "unverifiable"
)

// Location is one place a finding's value was observed. Dedup attaches
// multiple locations to a single finding (same secret across commits,
// renames, or files).
type Location struct {
	// Path is a file path (repo mode; forward slashes, relative to the scan
	// root where possible). URL is set in domain mode. At least one is set.
	Path string `json:"path,omitempty"`
	URL  string `json:"url,omitempty"`
	// Line is 1-based; 0 means not applicable (e.g. a whole-file probe).
	Line int `json:"line,omitempty"`
	// Commit and Branch are git-history provenance (M3); empty in working-tree scans.
	Commit string `json:"commit,omitempty"`
	Branch string `json:"branch,omitempty"`
	// FirstSeen is the commit date the value first appeared (M3).
	FirstSeen string `json:"first_seen,omitempty"`
	// StillPresent reports whether the value exists at HEAD / is live now.
	StillPresent bool `json:"still_present"`
}

// Finding is the internal, deduplicated record produced by the scanner.
// It contains no raw secret values — not even unexported ones. Raw values
// exist only in Candidate and die inside FromCandidate.
type Finding struct {
	ID          string `json:"id"`
	Fingerprint string `json:"fingerprint"`
	// FingerprintKeyVersion is the version of the HMAC key that produced
	// Fingerprint (D-008). Rotating the key invalidates stored fingerprints.
	FingerprintKeyVersion int            `json:"fingerprint_key_version"`
	Module                string         `json:"module"`
	RuleID                string         `json:"rule_id"`
	Type                  string         `json:"type"`
	Title                 string         `json:"title"`
	Severity              Severity       `json:"severity"`
	Confidence            float64        `json:"confidence"`
	Classification        Classification `json:"classification"`
	Verification          Verification   `json:"verification"`
	Locations             []Location     `json:"locations"`
	// Evidence is a snippet with surrounding context in which every
	// occurrence of the raw value has been replaced by its redacted form.
	Evidence    string   `json:"evidence"`
	Impact      string   `json:"impact"`
	Remediation string   `json:"remediation"`
	References  []string `json:"references,omitempty"`
}

// RedactedFinding is the public projection of a Finding. Report renderers,
// exporters, API responses, share links, and the LLM client accept only this
// type. Today it mirrors Finding field-for-field; the distinct type is the
// boundary — the day Finding grows a field that must not leave the engine
// (owner-only evidence, verification material), it simply does not appear
// here and no export can reach it.
type RedactedFinding struct {
	ID                    string         `json:"id"`
	Fingerprint           string         `json:"fingerprint"`
	FingerprintKeyVersion int            `json:"fingerprint_key_version"`
	Module                string         `json:"module"`
	RuleID                string         `json:"rule_id"`
	Type                  string         `json:"type"`
	Title                 string         `json:"title"`
	Severity              Severity       `json:"severity"`
	Confidence            float64        `json:"confidence"`
	Classification        Classification `json:"classification"`
	Verification          Verification   `json:"verification"`
	Locations             []Location     `json:"locations"`
	Evidence              string         `json:"evidence"`
	Impact                string         `json:"impact"`
	Remediation           string         `json:"remediation"`
	References            []string       `json:"references,omitempty"`
}

// Candidate is engine-internal raw material handed to FromCandidate. It is
// the only type in the codebase that may hold a raw secret value, and only
// in memory: MarshalJSON always fails so a Candidate can never leak into
// persisted data, logs, or exports through serialization.
type Candidate struct {
	Module         string
	RuleID         string
	Type           string
	Title          string
	Severity       Severity
	Confidence     float64
	Classification Classification
	Verification   Verification // zero value → VerificationNotAttempted

	// RawValue is the exposed value exactly as found. FromCandidate
	// normalizes it (whitespace-trimmed, base64-unwrapped) before
	// fingerprinting.
	RawValue string
	// Snippet is surrounding context that still contains RawValue verbatim;
	// FromCandidate replaces every occurrence with the redacted form. If the
	// module already redacted the snippet, leave RawValue out of it and this
	// is a no-op.
	Snippet string

	Location    Location
	Impact      string
	Remediation string
	References  []string
}

// ErrCandidateSerialization is returned by Candidate.MarshalJSON. Candidates
// hold raw secret values and must never be serialized.
var ErrCandidateSerialization = errors.New("finding.Candidate holds raw secret values and must never be serialized")

// MarshalJSON implements json.Marshaler with a hard tripwire: any attempt to
// serialize a Candidate fails loudly instead of leaking raw values.
func (c Candidate) MarshalJSON() ([]byte, error) {
	return nil, ErrCandidateSerialization
}

// FromCandidate is the one-way constructor at the heart of the redaction
// boundary. It fingerprints the normalized raw value under the versioned
// HMAC key, redacts the snippet, and returns a Finding with no raw value in
// it. The Candidate (and its raw value) must be discarded by the caller
// afterwards.
//
// An empty key is a hard error (D-008): there is no default fingerprint key.
func FromCandidate(c Candidate, key []byte, keyVersion int) (Finding, error) {
	if len(key) == 0 {
		return Finding{}, errors.New("fingerprint key is required: no built-in default key exists (DECISIONS.md D-008)")
	}
	if keyVersion < 1 {
		return Finding{}, fmt.Errorf("fingerprint key version must be >= 1, got %d", keyVersion)
	}
	if c.RawValue == "" {
		return Finding{}, errors.New("candidate has an empty raw value")
	}
	if err := validateSeverity(c.Severity); err != nil {
		return Finding{}, err
	}
	if c.Confidence < 0 || c.Confidence > 1 {
		return Finding{}, fmt.Errorf("confidence must be in [0,1], got %f", c.Confidence)
	}
	if c.Module == "" || c.RuleID == "" || c.Type == "" || c.Title == "" {
		return Finding{}, errors.New("candidate requires module, rule_id, type, and title")
	}

	normalized := NormalizeRaw(c.RawValue)
	ver := c.Verification
	if ver == "" {
		ver = VerificationNotAttempted
	}
	class := c.Classification
	if class == "" {
		class = ClassificationUnknown
	}

	f := Finding{
		Fingerprint:           ComputeFingerprint(normalized, key),
		FingerprintKeyVersion: keyVersion,
		Module:                c.Module,
		RuleID:                c.RuleID,
		Type:                  c.Type,
		Title:                 c.Title,
		Severity:              c.Severity,
		Confidence:            c.Confidence,
		Classification:        class,
		Verification:          ver,
		Locations:             []Location{c.Location},
		Evidence:              strings.ReplaceAll(c.Snippet, c.RawValue, RedactValue(c.RawValue)),
		Impact:                c.Impact,
		Remediation:           c.Remediation,
		References:            c.References,
	}
	return f, nil
}

// Redacted is the one-way projection to the public type. The engine's
// exporters and the server accept only RedactedFinding.
func (f Finding) Redacted() RedactedFinding {
	return RedactedFinding{
		ID:                    f.ID,
		Fingerprint:           f.Fingerprint,
		FingerprintKeyVersion: f.FingerprintKeyVersion,
		Module:                f.Module,
		RuleID:                f.RuleID,
		Type:                  f.Type,
		Title:                 f.Title,
		Severity:              f.Severity,
		Confidence:            f.Confidence,
		Classification:        f.Classification,
		Verification:          f.Verification,
		Locations:             f.Locations,
		Evidence:              f.Evidence,
		Impact:                f.Impact,
		Remediation:           f.Remediation,
		References:            f.References,
	}
}

// Validate checks schema invariants. The scanner calls it before emitting a
// report; tests call it on fixtures.
func (f Finding) Validate() error {
	if f.ID == "" {
		return errors.New("finding.id is empty")
	}
	if f.Fingerprint == "" {
		return errors.New("finding.fingerprint is empty")
	}
	if f.FingerprintKeyVersion < 1 {
		return errors.New("finding.fingerprint_key_version must be >= 1")
	}
	if f.Module == "" || f.RuleID == "" || f.Type == "" || f.Title == "" {
		return errors.New("finding requires module, rule_id, type, and title")
	}
	if err := validateSeverity(f.Severity); err != nil {
		return err
	}
	if f.Confidence < 0 || f.Confidence > 1 {
		return fmt.Errorf("confidence must be in [0,1], got %f", f.Confidence)
	}
	if err := validateClassification(f.Classification); err != nil {
		return err
	}
	if err := validateVerification(f.Verification); err != nil {
		return err
	}
	if len(f.Locations) == 0 {
		return errors.New("finding has no locations")
	}
	for i, loc := range f.Locations {
		if loc.Path == "" && loc.URL == "" {
			return fmt.Errorf("location %d has neither path nor url", i)
		}
	}
	return nil
}

// DedupKey is the (rule_id, value_hash) identity used to merge findings that
// are the same secret observed in multiple places (SPEC §4).
func (f Finding) DedupKey() string {
	return f.RuleID + "|" + f.Fingerprint
}

func validateSeverity(s Severity) error {
	switch s {
	case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo:
		return nil
	default:
		return fmt.Errorf("invalid severity %q", s)
	}
}

func validateClassification(c Classification) error {
	switch c {
	case ClassificationPublicByDesign, ClassificationDangerousIfMisconfigured,
		ClassificationCritical, ClassificationUnknown:
		return nil
	default:
		return fmt.Errorf("invalid classification %q", c)
	}
}

func validateVerification(v Verification) error {
	switch v {
	case VerificationNotAttempted, VerificationVerifiedLive,
		VerificationVerifiedInactive, VerificationUnverifiable:
		return nil
	default:
		return fmt.Errorf("invalid verification %q", v)
	}
}

// ComputeFingerprint is HMAC-SHA256 hex of the (already normalized) raw
// value under key. Never fingerprint the redacted form.
func ComputeFingerprint(normalizedRaw string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(normalizedRaw))
	return hex.EncodeToString(mac.Sum(nil))
}

// RedactValue renders the storage/display form of a raw value: first 4 and
// last 4 characters with the middle elided; values of 12 characters or fewer
// are masked entirely so the redacted form does not disclose most of the
// value. The hash lives in the fingerprint, not here.
func RedactValue(raw string) string {
	r := []rune(raw)
	if len(r) <= 12 {
		return strings.Repeat("*", len(r))
	}
	return string(r[:4]) + "…" + string(r[len(r)-4:])
}

// minB64Len is the shortest string worth attempting to unwrap as base64.
const minB64Len = 16

// b64Pattern matches strings that could be standard or URL-safe base64,
// padded or not.
var b64Pattern = regexp.MustCompile(`^[A-Za-z0-9+/=_-]+$`)

// NormalizeRaw canonicalizes a raw value for fingerprinting: whitespace is
// trimmed, and a value that is itself a base64 wrapping of the secret is
// unwrapped, so the same secret observed raw or base64-encoded fingerprints
// identically (SPEC §4). Unwrapping only happens when the decode is strict,
// the result is valid UTF-8 with no control characters, and the input is
// plausibly base64 — an ordinary secret that merely looks like base64 is
// still normalized deterministically, so fingerprints stay stable.
func NormalizeRaw(raw string) string {
	n := strings.TrimSpace(raw)
	if len(n) < minB64Len || !b64Pattern.MatchString(n) {
		return n
	}
	decoded, ok := tryBase64(n)
	if !ok {
		return n
	}
	return decoded
}

func tryBase64(s string) (string, bool) {
	for _, enc := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding,
		base64.URLEncoding, base64.RawURLEncoding,
	} {
		d, err := enc.DecodeString(s)
		if err != nil {
			continue
		}
		if !utf8.Valid(d) || hasControlChars(d) {
			continue
		}
		return string(d), true
	}
	return "", false
}

func hasControlChars(b []byte) bool {
	for _, c := range b {
		if c < 0x20 || c == 0x7f {
			return true
		}
	}
	return false
}
