package stubmod

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iffayush/wass/engine"
	"github.com/iffayush/wass/engine/finding"
)

const fixtureValue = "WASSFAKE-Q1W2E3R4T5Y6U7I8O9P0A1S2D3F4G5H6"

func scanFixture(t *testing.T, mutate func(*Module)) ([]finding.Candidate, *Module) {
	t.Helper()
	m := &Module{}
	if mutate != nil {
		mutate(m)
	}
	var got []finding.Candidate
	root := filepath.Join("testdata", "fixture")
	err := m.Scan(context.Background(), engine.Target{Kind: engine.TargetRepo, RepoPath: root},
		func(c finding.Candidate) error {
			got = append(got, c)
			return nil
		})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	return got, m
}

func TestStubFindsFakeKeys(t *testing.T) {
	got, m := scanFixture(t, nil)

	// settings.txt (1) + notes.md (1); clean.txt has none; binary.bin is skipped.
	if len(got) != 2 {
		t.Fatalf("expected 2 candidates, got %d: %+v", len(got), got)
	}
	if m.FilesScanned != 3 || m.FilesSkippedBinary != 1 {
		t.Fatalf("counters wrong: scanned=%d skippedBinary=%d", m.FilesScanned, m.FilesSkippedBinary)
	}

	byPath := map[string]finding.Candidate{}
	for _, c := range got {
		byPath[c.Location.Path] = c
	}
	settings, ok := byPath["config/settings.txt"]
	if !ok {
		t.Fatalf("no candidate for config/settings.txt: %+v", byPath)
	}
	if settings.Location.Line != 3 {
		t.Fatalf("settings.txt match should be line 3, got %d", settings.Location.Line)
	}
	notes := byPath["docs/notes.md"]
	if notes.Location.Line != 3 {
		t.Fatalf("notes.md match should be line 3, got %d", notes.Location.Line)
	}

	for _, c := range got {
		if c.RawValue != fixtureValue {
			t.Fatalf("unexpected raw value %q", c.RawValue)
		}
		if !strings.Contains(c.Snippet, c.RawValue) {
			t.Fatalf("candidate snippet must carry the raw value pre-redaction: %q", c.Snippet)
		}
		if c.RuleID != RuleID || c.Module != "stub" || c.Type != TypeName {
			t.Fatalf("candidate identity fields wrong: %+v", c)
		}
		if !c.Location.StillPresent {
			t.Fatal("working-tree findings are still present")
		}
	}
}

func TestStubSkipsLargeFiles(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "big.txt")
	payload := strings.Repeat("WASSFAKE-Q1W2E3R4T5Y6U7I8O9P0A1S2D3F4G5H6\n", 20)
	if err := os.WriteFile(big, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	m := &Module{MaxFileBytes: 64}
	var got []finding.Candidate
	err := m.Scan(context.Background(), engine.Target{Kind: engine.TargetRepo, RepoPath: dir},
		func(c finding.Candidate) error { got = append(got, c); return nil })
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 0 || m.FilesSkippedSize != 1 {
		t.Fatalf("large file not skipped: candidates=%d skippedSize=%d", len(got), m.FilesSkippedSize)
	}
}

func TestStubSkipsSkipListedDirs(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	hidden := filepath.Join(dir, ".git", "fake.txt")
	if err := os.WriteFile(hidden, []byte("api_key = \""+fixtureValue+"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m := &Module{}
	var got []finding.Candidate
	err := m.Scan(context.Background(), engine.Target{Kind: engine.TargetRepo, RepoPath: dir},
		func(c finding.Candidate) error { got = append(got, c); return nil })
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 0 || m.DirsSkippedSkipList != 1 {
		t.Fatalf(".git not skipped: candidates=%d skippedDirs=%d", len(got), m.DirsSkippedSkipList)
	}
}

func TestStubRejectsNonRepoTargets(t *testing.T) {
	m := &Module{}
	err := m.Scan(context.Background(), engine.Target{Kind: engine.TargetDomain, URL: "https://example.com"},
		func(finding.Candidate) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "domain") {
		t.Fatalf("expected domain-target rejection, got %v", err)
	}
}

func TestStubErrorsOnMissingRoot(t *testing.T) {
	m := &Module{}
	err := m.Scan(context.Background(), engine.Target{Kind: engine.TargetRepo, RepoPath: "testdata/does-not-exist"},
		func(finding.Candidate) error { return nil })
	if err == nil {
		t.Fatal("expected error for missing root")
	}
}

func TestStubHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := &Module{}
	err := m.Scan(ctx, engine.Target{Kind: engine.TargetRepo, RepoPath: filepath.Join("testdata", "fixture")},
		func(finding.Candidate) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
