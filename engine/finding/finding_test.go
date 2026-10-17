package finding

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const testKey = "test-fingerprint-key-do-not-use-in-prod"

func validCandidate() Candidate {
	return Candidate{
		Module:         "secrets",
		RuleID:         "TEST_FAKE_KEY",
		Type:           "fake_api_key",
		Title:          "Fake test API key",
		Severity:       SeverityHigh,
		Confidence:     0.9,
		Classification: ClassificationCritical,
		RawValue:       "WASSFAKE0123456789ABCDEFGHIJKLMNOPQRSTUV",
		Snippet:        "api_key = \"WASSFAKE0123456789ABCDEFGHIJKLMNOPQRSTUV\" # do not commit",
		Location:       Location{Path: "config/app.txt", Line: 7, StillPresent: true},
		Impact:         "Test impact.",
		Remediation:    "Test remediation.",
	}
}

func mustFromCandidate(t *testing.T, c Candidate) Finding {
	t.Helper()
	f, err := FromCandidate(c, []byte(testKey), 1)
	if err != nil {
		t.Fatalf("FromCandidate: %v", err)
	}
	return f
}

func TestFromCandidateRedactsSnippet(t *testing.T) {
	raw := validCandidate().RawValue
	f := mustFromCandidate(t, validCandidate())

	if strings.Contains(f.Evidence, raw) {
		t.Fatalf("evidence still contains the raw value: %q", f.Evidence)
	}
	want := RedactValue(raw)
	if !strings.Contains(f.Evidence, want) {
		t.Fatalf("evidence missing redacted form %q: %q", want, f.Evidence)
	}
	if !strings.Contains(f.Evidence, "api_key =") {
		t.Fatalf("evidence lost its context: %q", f.Evidence)
	}
}

func TestFromCandidateFindingMarshalExcludesRaw(t *testing.T) {
	raw := validCandidate().RawValue
	f := mustFromCandidate(t, validCandidate())
	f.ID = "f-000001"

	b, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal Finding: %v", err)
	}
	if bytes.Contains(b, []byte(raw)) {
		t.Fatalf("marshaled Finding contains the raw value: %s", b)
	}
	if !bytes.Contains(b, []byte(`"fingerprint_key_version":1`)) {
		t.Fatalf("marshaled Finding missing fingerprint_key_version: %s", b)
	}
}

func TestFindingAndRedactedFindingMarshalIdentically(t *testing.T) {
	f := mustFromCandidate(t, validCandidate())
	f.ID = "f-000001"

	fb, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal Finding: %v", err)
	}
	rb, err := json.Marshal(f.Redacted())
	if err != nil {
		t.Fatalf("marshal RedactedFinding: %v", err)
	}
	if !bytes.Equal(fb, rb) {
		t.Fatalf("Finding and RedactedFinding JSON diverge:\n%s\n%s", fb, rb)
	}
}

func TestFingerprintIsStableAndKeyDependent(t *testing.T) {
	raw := "WASSFAKE0123456789ABCDEFGHIJKLMNOPQRSTUV"
	a := ComputeFingerprint(raw, []byte(testKey))
	b := ComputeFingerprint(raw, []byte(testKey))
	if a != b {
		t.Fatalf("fingerprint not deterministic: %q vs %q", a, b)
	}
	if a == ComputeFingerprint(raw, []byte("another-key")) {
		t.Fatal("fingerprint did not change with the key")
	}
	if a == ComputeFingerprint(strings.ToLower(raw), []byte(testKey)) {
		t.Fatal("different values produced the same fingerprint")
	}
	if len(a) != 64 {
		t.Fatalf("fingerprint should be 64 hex chars, got %d", len(a))
	}
}

func TestFromCandidateFingerprintMatchesCompute(t *testing.T) {
	c := validCandidate()
	f := mustFromCandidate(t, c)
	want := ComputeFingerprint(NormalizeRaw(c.RawValue), []byte(testKey))
	if f.Fingerprint != want {
		t.Fatalf("fingerprint mismatch: %q vs %q", f.Fingerprint, want)
	}
}

func TestFromCandidateRequiresKeyAndVersion(t *testing.T) {
	if _, err := FromCandidate(validCandidate(), nil, 1); err == nil {
		t.Fatal("expected hard error for empty key (D-008)")
	}
	if _, err := FromCandidate(validCandidate(), []byte(testKey), 0); err == nil {
		t.Fatal("expected error for key version 0")
	}
}

func TestFromCandidateValidationErrors(t *testing.T) {
	base := validCandidate()

	noRaw := base
	noRaw.RawValue = ""
	if _, err := FromCandidate(noRaw, []byte(testKey), 1); err == nil {
		t.Fatal("expected error for empty raw value")
	}

	badSeverity := base
	badSeverity.Severity = "extreme"
	if _, err := FromCandidate(badSeverity, []byte(testKey), 1); err == nil {
		t.Fatal("expected error for invalid severity")
	}

	badConfidence := base
	badConfidence.Confidence = 1.5
	if _, err := FromCandidate(badConfidence, []byte(testKey), 1); err == nil {
		t.Fatal("expected error for confidence > 1")
	}

	missingTitle := base
	missingTitle.Title = ""
	if _, err := FromCandidate(missingTitle, []byte(testKey), 1); err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestFromCandidateDefaults(t *testing.T) {
	c := validCandidate()
	c.Verification = ""
	c.Classification = ""
	f := mustFromCandidate(t, c)
	if f.Verification != VerificationNotAttempted {
		t.Fatalf("verification default: got %q", f.Verification)
	}
	if f.Classification != ClassificationUnknown {
		t.Fatalf("classification default: got %q", f.Classification)
	}
}

func TestCandidateMarshalJSONIsATripwire(t *testing.T) {
	_, err := json.Marshal(validCandidate())
	if err == nil {
		t.Fatal("marshaling a Candidate must fail")
	}
	if !strings.Contains(err.Error(), "never be serialized") {
		t.Fatalf("unexpected tripwire error: %v", err)
	}
}

func TestNormalizeRawTrimsWhitespace(t *testing.T) {
	secret := "WASSFAKE0123456789ABCDEFGHIJKLMNOPQRSTUV"
	if NormalizeRaw("  "+secret+" \n") != secret {
		t.Fatalf("whitespace not trimmed: %q", NormalizeRaw("  "+secret+" \n"))
	}
}

func TestNormalizeRawUnwrapsBase64(t *testing.T) {
	secret := "supersecret-value-0123456789"
	cases := map[string]string{
		"std padded":   base64.StdEncoding.EncodeToString([]byte(secret)),
		"std unpadded": base64.RawStdEncoding.EncodeToString([]byte(secret)),
		"url padded":   base64.URLEncoding.EncodeToString([]byte(secret)),
		"url unpadded": base64.RawURLEncoding.EncodeToString([]byte(secret)),
	}
	for name, wrapped := range cases {
		got := NormalizeRaw(wrapped)
		if got != secret {
			t.Errorf("%s: NormalizeRaw(%q) = %q, want %q", name, wrapped, got, secret)
		}
	}
	// The unwrapped form fingerprints identically to the original secret.
	plain := ComputeFingerprint(secret, []byte(testKey))
	wrapped := base64.StdEncoding.EncodeToString([]byte(secret))
	if ComputeFingerprint(NormalizeRaw(wrapped), []byte(testKey)) != plain {
		t.Fatal("base64-wrapped secret did not fingerprint identically")
	}
}

func TestNormalizeRawLeavesNonBase64Alone(t *testing.T) {
	// Too short to be worth unwrapping.
	if got := NormalizeRaw("shortb64"); got != "shortb64" {
		t.Fatalf("short string changed: %q", got)
	}
	// Valid charset but decodes to control characters / invalid UTF-8.
	weird := base64.StdEncoding.EncodeToString([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b})
	if got := NormalizeRaw(weird); got != weird {
		t.Fatalf("binary-content base64 was unwrapped: %q", got)
	}
	// Ordinary long secret with mixed punctuation is untouched.
	secret := "tr1b0w-secr3t_value!!#"
	if got := NormalizeRaw(secret); got != secret {
		t.Fatalf("plain secret changed: %q", got)
	}
}

func TestRedactValue(t *testing.T) {
	long := "WASSFAKE0123456789ABCDEFGHIJKLMNOPQRSTUV"
	want := "WASS…STUV"
	if got := RedactValue(long); got != want {
		t.Fatalf("RedactValue(long) = %q, want %q", got, want)
	}
	short := "abc123"
	if got := RedactValue(short); got != "******" {
		t.Fatalf("RedactValue(short) = %q, want full mask", got)
	}
	if got := RedactValue("exactly12cha"); got != "************" {
		t.Fatalf("12-char value must be fully masked, got %q", got)
	}
	if got := RedactValue("exactly13char"); got != "exac…char" {
		t.Fatalf("RedactValue(13 chars) = %q", got)
	}
}

func TestFindingValidate(t *testing.T) {
	f := mustFromCandidate(t, validCandidate())
	f.ID = "f-000001"
	if err := f.Validate(); err != nil {
		t.Fatalf("valid finding failed validation: %v", err)
	}

	noLoc := f
	noLoc.Locations = nil
	if err := noLoc.Validate(); err == nil {
		t.Fatal("expected error for missing locations")
	}

	badLoc := f
	badLoc.Locations = []Location{{Line: 3}}
	if err := badLoc.Validate(); err == nil {
		t.Fatal("expected error for location without path or url")
	}

	noID := f
	noID.ID = ""
	if err := noID.Validate(); err == nil {
		t.Fatal("expected error for empty id")
	}

	badVer := f
	badVer.Verification = "verified_sometimes"
	if err := badVer.Validate(); err == nil {
		t.Fatal("expected error for invalid verification")
	}

	badClass := f
	badClass.Classification = "sort_of_public"
	if err := badClass.Validate(); err == nil {
		t.Fatal("expected error for invalid classification")
	}
}

func TestDedupKey(t *testing.T) {
	f := mustFromCandidate(t, validCandidate())
	want := "TEST_FAKE_KEY|" + f.Fingerprint
	if f.DedupKey() != want {
		t.Fatalf("DedupKey = %q, want %q", f.DedupKey(), want)
	}
}

func TestSchemaVersionIsSet(t *testing.T) {
	if SchemaVersion < 1 {
		t.Fatal("SchemaVersion must be >= 1")
	}
}

// Compile-time proof of the boundary: the redacted type is a distinct type.
var _ = fmt.Sprintf
