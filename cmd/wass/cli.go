// The wass CLI: subcommand parsing, the scan command, and JSON output.
//
// Usage (M1):
//
//	wass scan --repo <path> --fingerprint-key <key>
//	wass scan --repo <path>            # key from WASS_FINGERPRINT_KEY
//
// The scan report is emitted as JSON on stdout; diagnostics go to stderr.
// The output is a RedactedReport — it carries no raw secret values by
// construction (engine/finding redaction boundary).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/iffayush/wass/engine"
	"github.com/iffayush/wass/engine/stubmod"
)

// envFingerprintKey is the environment variable the fingerprint key may come
// from when the flag is absent (DECISIONS.md D-008).
const envFingerprintKey = "WASS_FINGERPRINT_KEY"

// FingerprintKeyVersion is the version stamped on fingerprints produced by
// this build's key handling. Key rotation bumps this and requires the
// migration plan in D-008; it is a constant for the MVP, not configurable
// per scan, so all fingerprints of a build agree.
const FingerprintKeyVersion = 1

const defaultTimeout = 10 * time.Minute

// Main runs the CLI and returns the process exit code. Args are the program
// arguments without the command name (os.Args[1:]).
func cliMain(args []string) int {
	stdout, stderr := os.Stdout, os.Stderr
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "scan":
		return runScan(stdout, stderr, args[1:])
	case "--version", "version":
		fmt.Fprintln(stdout, "wass", engine.Version)
		return 0
	case "--help", "-h", "help":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "wass: unknown command %q\n\n", args[0])
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `wass %s — web application secret scanner

Usage:
  wass scan --repo <path>   [--fingerprint-key <key>] [--timeout <dur>] [--pretty]
  wass scan --domain <url>  (arrives in M4)
  wass --version

The fingerprint key comes from --fingerprint-key or %s;
there is no built-in default key (D-008).

The report is printed as JSON on stdout.
`, engine.Version, envFingerprintKey)
}

func runScan(stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		repoPath       = fs.String("repo", "", "repository working-tree path to scan")
		domainURL      = fs.String("domain", "", "deployed app URL to scan (M4)")
		fingerprintKey = fs.String("fingerprint-key", "", "HMAC key for finding fingerprints (or "+envFingerprintKey+")")
		timeout        = fs.Duration("timeout", defaultTimeout, "per-scan timeout")
		pretty         = fs.Bool("pretty", false, "indent the JSON output")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	key := strings.TrimSpace(*fingerprintKey)
	if key == "" {
		key = strings.TrimSpace(os.Getenv(envFingerprintKey))
	}
	if key == "" {
		fmt.Fprintf(stderr, "wass scan: a fingerprint key is required: pass --fingerprint-key or set %s (no built-in default key exists, D-008)\n", envFingerprintKey)
		return 2
	}

	var target engine.Target
	switch {
	case *repoPath != "" && *domainURL == "":
		target = engine.Target{Kind: engine.TargetRepo, RepoPath: *repoPath}
	case *repoPath == "" && *domainURL != "":
		fmt.Fprintln(stderr, "wass scan: domain scanning arrives in M4 (docs/TASKS.md)")
		return 2
	default:
		fmt.Fprintln(stderr, "wass scan: exactly one of --repo or --domain is required")
		return 2
	}

	scanner := engine.NewScanner(&stubmod.Module{})
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	report, err := scanner.Scan(ctx, target, engine.Options{
		FingerprintKey:        []byte(key),
		FingerprintKeyVersion: FingerprintKeyVersion,
	})
	if err != nil {
		fmt.Fprintf(stderr, "wass scan: %v\n", err)
		return 1
	}

	return emitReport(stdout, report.Redacted(), *pretty)
}

func emitReport(stdout io.Writer, rep engine.RedactedReport, pretty bool) int {
	enc := json.NewEncoder(stdout)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(rep); err != nil {
		fmt.Fprintf(os.Stderr, "wass scan: encoding report: %v\n", err)
		return 1
	}
	return 0
}

// ErrNoKey is kept exported for callers embedding the CLI (tests).
var ErrNoKey = errors.New("fingerprint key required")
