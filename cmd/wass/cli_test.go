package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fakeValue = "WASSFAKE-Q1W2E3R4T5Y6U7I8O9P0A1S2D3F4G5H6"

func fixtureDir() string {
	// The stub module's fixture tree doubles as the CLI end-to-end target.
	return filepath.Join("..", "..", "engine", "stubmod", "testdata", "fixture")
}

func TestCLIScanRepoEmitsJSON(t *testing.T) {
	var out, errb bytes.Buffer
	code := run(t, &out, &errb, "scan", "--repo", fixtureDir(), "--fingerprint-key", "cli-test-key", "--pretty")
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, errb.String())
	}

	var rep map[string]any
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out.String())
	}
	if rep["schema_version"] != float64(1) {
		t.Fatalf("schema_version = %v", rep["schema_version"])
	}
	findings, ok := rep["findings"].([]any)
	if !ok || len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %v", rep["findings"])
	}
	f := findings[0].(map[string]any)
	if f["rule_id"] != "STUB_FAKE_KEY" {
		t.Fatalf("rule_id = %v", f["rule_id"])
	}
	locs := f["locations"].([]any)
	if len(locs) != 2 {
		t.Fatalf("expected 2 deduped locations, got %d", len(locs))
	}
	if ev, _ := f["evidence"].(string); !strings.Contains(ev, "WASS…G5H6") {
		t.Fatalf("evidence not redacted: %q", ev)
	}

	// The redaction boundary, end to end: no raw value anywhere in the output.
	if bytes.Contains(out.Bytes(), []byte(fakeValue)) {
		t.Fatal("CLI output contains the raw fake value")
	}
}

func TestCLIScanKeyFromEnv(t *testing.T) {
	t.Setenv("WASS_FINGERPRINT_KEY", "env-key")
	var out, errb bytes.Buffer
	code := run(t, &out, &errb, "scan", "--repo", fixtureDir())
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, errb.String())
	}
	if !bytes.Contains(out.Bytes(), []byte(`"findings_count":1`)) {
		t.Fatalf("env-key scan found nothing: %s", out.String())
	}
}

func TestCLIScanRequiresKey(t *testing.T) {
	os.Unsetenv("WASS_FINGERPRINT_KEY")
	var out, errb bytes.Buffer
	code := run(t, &out, &errb, "scan", "--repo", fixtureDir())
	if code == 0 {
		t.Fatal("scan without a key must fail (D-008)")
	}
	if !strings.Contains(errb.String(), "fingerprint key") {
		t.Fatalf("missing-key error not helpful: %q", errb.String())
	}
	if out.Len() != 0 {
		t.Fatalf("failed scan must not emit a report: %s", out.String())
	}
}

func TestCLIScanDomainRejectedUntilM4(t *testing.T) {
	var out, errb bytes.Buffer
	code := run(t, &out, &errb, "scan", "--domain", "https://example.com", "--fingerprint-key", "k")
	if code == 0 || !strings.Contains(errb.String(), "M4") {
		t.Fatalf("domain scan should be rejected with an M4 pointer: %d %q", code, errb.String())
	}
}

func TestCLIScanBothTargetsRejected(t *testing.T) {
	var out, errb bytes.Buffer
	code := run(t, &out, &errb, "scan", "--repo", ".", "--domain", "https://x", "--fingerprint-key", "k")
	if code == 0 || !strings.Contains(errb.String(), "exactly one") {
		t.Fatalf("both-targets scan should be rejected: %d %q", code, errb.String())
	}
}

func TestCLIScanMissingRepo(t *testing.T) {
	var out, errb bytes.Buffer
	code := run(t, &out, &errb, "scan", "--repo", "testdata/no-such-dir", "--fingerprint-key", "k")
	if code == 0 {
		t.Fatal("scan of missing directory must fail")
	}
}

func TestCLIUsageAndVersion(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run(t, &out, &errb, "--version"); code != 0 || !strings.HasPrefix(out.String(), "wass ") {
		t.Fatalf("--version: code=%d out=%q", code, out.String())
	}
	out.Reset()
	errb.Reset()
	if code := run(t, &out, &errb); code != 2 || !strings.Contains(errb.String(), "Usage") {
		t.Fatalf("no-args usage: code=%d err=%q", code, errb.String())
	}
	out.Reset()
	errb.Reset()
	if code := run(t, &out, &errb, "frobnicate"); code != 2 || !strings.Contains(errb.String(), "unknown command") {
		t.Fatalf("unknown command: code=%d err=%q", code, errb.String())
	}
}

// run executes cliMain with stdout/stderr captured, because cliMain writes
// to the real os.Stdout otherwise.
func run(t *testing.T, out, errb *bytes.Buffer, args ...string) int {
	t.Helper()
	stdout, stderr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = wOut, wErr

	code := cliMain(args)

	os.Stdout, os.Stderr = stdout, stderr
	wOut.Close()
	wErr.Close()
	buf := make([]byte, 64<<10)
	n, _ := rOut.Read(buf)
	out.Write(buf[:n])
	n, _ = rErr.Read(buf)
	errb.Write(buf[:n])
	return code
}
