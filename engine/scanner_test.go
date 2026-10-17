package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/iffayush/wass/engine/finding"
)

const scanKey = "scanner-test-key"

// fakeModule emits a fixed list of candidates (and optionally errors).
type fakeModule struct {
	name      string
	cands     []finding.Candidate
	err       error
	seenT     Target
	seenEmitN int
}

func (f *fakeModule) Name() string { return f.name }

func (f *fakeModule) Scan(ctx context.Context, t Target, emit func(finding.Candidate) error) error {
	f.seenT = t
	for _, c := range f.cands {
		f.seenEmitN++
		if err := emit(c); err != nil {
			return err
		}
	}
	return f.err
}

func cand(ruleID, raw, path string, line int) finding.Candidate {
	return finding.Candidate{
		Module:         "fake",
		RuleID:         ruleID,
		Type:           "fake_secret",
		Title:          "Fake secret",
		Severity:       finding.SeverityHigh,
		Confidence:     0.9,
		Classification: finding.ClassificationCritical,
		RawValue:       raw,
		Snippet:        "token=" + raw,
		Location:       finding.Location{Path: path, Line: line, StillPresent: true},
		Impact:         "impact",
		Remediation:    "remediation",
	}
}

var (
	valA = "WASSFAKE0123456789ABCDEFGHIJKLMNOPQRSTUV"
	valB = "WASSFAKE9876543210ZYXWVUTSRQPONMLKJIHGFED"
)

func TestScannerDedupsSameValueAcrossLocations(t *testing.T) {
	m := &fakeModule{name: "fake", cands: []finding.Candidate{
		cand("RULE_A", valA, "a.txt", 1),
		cand("RULE_A", valA, "b.txt", 20),
		cand("RULE_A", valA, "a.txt", 1), // exact duplicate location
	}}
	rep, err := NewScanner(m).Scan(context.Background(),
		Target{Kind: TargetRepo, RepoPath: "/fixtures"}, Options{FingerprintKey: []byte(scanKey), FingerprintKeyVersion: 1})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if rep.FindingsCount != 1 || len(rep.Findings) != 1 {
		t.Fatalf("expected 1 deduped finding, got %d", rep.FindingsCount)
	}
	f := rep.Findings[0]
	if len(f.Locations) != 2 {
		t.Fatalf("expected 2 unique locations, got %d: %+v", len(f.Locations), f.Locations)
	}
	if f.ID != "f-000001" {
		t.Fatalf("finding id not assigned: %q", f.ID)
	}
	if f.Locations[0].Path != "a.txt" || f.Locations[1].Path != "b.txt" {
		t.Fatalf("unexpected location order: %+v", f.Locations)
	}
}

func TestScannerKeepsDistinctValuesAndRulesApart(t *testing.T) {
	m := &fakeModule{name: "fake", cands: []finding.Candidate{
		cand("RULE_A", valA, "a.txt", 1),
		cand("RULE_A", valB, "a.txt", 5), // different value → different finding
		cand("RULE_B", valA, "a.txt", 9), // same value, different rule → different finding
	}}
	rep, err := NewScanner(m).Scan(context.Background(),
		Target{Kind: TargetRepo, RepoPath: "/fixtures"}, Options{FingerprintKey: []byte(scanKey), FingerprintKeyVersion: 1})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if rep.FindingsCount != 3 {
		t.Fatalf("expected 3 findings, got %d", rep.FindingsCount)
	}
	ids := map[string]bool{}
	for _, f := range rep.Findings {
		if ids[f.ID] {
			t.Fatalf("duplicate finding id %q", f.ID)
		}
		ids[f.ID] = true
		if err := f.Validate(); err != nil {
			t.Fatalf("assembled finding invalid: %v", err)
		}
	}
}

func TestScannerModuleErrorPropagates(t *testing.T) {
	boom := errors.New("boom")
	m := &fakeModule{name: "exploding", err: boom}
	_, err := NewScanner(m).Scan(context.Background(),
		Target{Kind: TargetRepo, RepoPath: "/fixtures"}, Options{FingerprintKey: []byte(scanKey), FingerprintKeyVersion: 1})
	if err == nil || !strings.Contains(err.Error(), "exploding") || !errors.Is(err, boom) {
		t.Fatalf("module error not wrapped with module name: %v", err)
	}
}

func TestScannerCandidateConversionErrorAborts(t *testing.T) {
	bad := cand("RULE_A", valA, "a.txt", 1)
	bad.Severity = "not-a-severity"
	m := &fakeModule{name: "fake", cands: []finding.Candidate{bad}}
	_, err := NewScanner(m).Scan(context.Background(),
		Target{Kind: TargetRepo, RepoPath: "/fixtures"}, Options{FingerprintKey: []byte(scanKey), FingerprintKeyVersion: 1})
	if err == nil || !strings.Contains(err.Error(), "module fake") {
		t.Fatalf("candidate conversion error not attributed to module: %v", err)
	}
}

func TestScannerRequiresKeyAndVersion(t *testing.T) {
	m := &fakeModule{name: "fake"}
	if _, err := NewScanner(m).Scan(context.Background(), Target{Kind: TargetRepo, RepoPath: "/f"}, Options{}); err == nil {
		t.Fatal("expected error for missing fingerprint key")
	}
	if _, err := NewScanner(m).Scan(context.Background(), Target{Kind: TargetRepo, RepoPath: "/f"},
		Options{FingerprintKey: []byte(scanKey)}); err == nil {
		t.Fatal("expected error for missing key version")
	}
	if _, err := NewScanner(m).Scan(context.Background(), Target{Kind: TargetRepo, RepoPath: "/f"},
		Options{FingerprintKey: []byte(scanKey), FingerprintKeyVersion: 1}); err != nil {
		t.Fatalf("valid options rejected: %v", err)
	}
}

func TestScannerRecordsModulesAndTarget(t *testing.T) {
	m1 := &fakeModule{name: "first"}
	m2 := &fakeModule{name: "second"}
	tgt := Target{Kind: TargetRepo, RepoPath: "/fixtures"}
	rep, err := NewScanner(m1, m2).Scan(context.Background(), tgt,
		Options{FingerprintKey: []byte(scanKey), FingerprintKeyVersion: 1})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(rep.ModulesRun) != 2 || rep.ModulesRun[0] != "first" || rep.ModulesRun[1] != "second" {
		t.Fatalf("modules run not recorded in order: %v", rep.ModulesRun)
	}
	if m1.seenT.RepoPath != "/fixtures" || m2.seenT.RepoPath != "/fixtures" {
		t.Fatal("modules did not receive the target")
	}
	if rep.SchemaVersion != finding.SchemaVersion || rep.FindingsCount != 0 || rep.Findings != nil {
		t.Fatalf("empty report malformed: %+v", rep)
	}
	if rep.FinishedAt.Before(rep.StartedAt) {
		t.Fatal("finished before started")
	}
}

func TestRedactedReportJSONExcludesRaw(t *testing.T) {
	m := &fakeModule{name: "fake", cands: []finding.Candidate{cand("RULE_A", valA, "a.txt", 1)}}
	rep, err := NewScanner(m).Scan(context.Background(),
		Target{Kind: TargetRepo, RepoPath: "/fixtures"}, Options{FingerprintKey: []byte(scanKey), FingerprintKeyVersion: 1})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	b, err := json.Marshal(rep.Redacted())
	if err != nil {
		t.Fatalf("marshal RedactedReport: %v", err)
	}
	if bytes.Contains(b, []byte(valA)) {
		t.Fatalf("redacted report contains the raw value: %s", b)
	}
	if !bytes.Contains(b, []byte(`"findings_count":1`)) || !bytes.Contains(b, []byte(`"schema_version":1`)) {
		t.Fatalf("redacted report envelope incomplete: %s", b)
	}
	// And a Candidate can never be smuggled through JSON at all.
	if _, err := json.Marshal(m.cands[0]); err == nil {
		t.Fatal("Candidate serialization tripwire did not fire")
	}
}

func TestTargetDescribe(t *testing.T) {
	if got := (Target{Kind: TargetRepo, RepoPath: "/x"}).Describe(); got != "/x" {
		t.Fatalf("repo describe: %q", got)
	}
	if got := (Target{Kind: TargetDomain, URL: "https://x"}).Describe(); got != "https://x" {
		t.Fatalf("domain describe: %q", got)
	}
}
