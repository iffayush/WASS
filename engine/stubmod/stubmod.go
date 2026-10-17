// Package stubmod implements the M1 placeholder engine module. It scans a
// repository working tree for the obviously-fake WASSFAKE-* token so the
// scanner core, CLI, and report pipeline can be exercised end to end before
// the real secrets rules engine lands in M2 (docs/TASKS.md).
//
// The pattern deliberately resembles no real provider's key formats, and
// every value in testdata fixtures is generated and fake.
package stubmod

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/iffayush/wass/engine"
	"github.com/iffayush/wass/engine/finding"
)

const (
	// RuleID is the stable rule identifier used in dedup keys and reports.
	RuleID = "STUB_FAKE_KEY"
	// TypeName is the finding type emitted by this module.
	TypeName = "stub_fake_key"

	fakePattern     = `WASSFAKE-[A-Z0-9]{32}`
	defaultMaxFile  = int64(5 << 20) // 5 MiB per-file guard (M2 adds the real, configurable guards)
	snippetContext  = 80             // runes of context on each side of a match
	snippetMaxRunes = 240            // maximum snippet length
	binarySniffLen  = 8192
)

var fakeRe = regexp.MustCompile(fakePattern)

// Module is the placeholder scanner module. Its exported counters are
// simple per-run statistics (single-threaded walk; not concurrency-safe).
type Module struct {
	// MaxFileBytes skips files larger than this. 0 uses the default (5 MiB).
	MaxFileBytes int64

	FilesScanned        int
	FilesSkippedSize    int
	FilesSkippedBinary  int
	DirsSkippedSkipList int
}

// Name implements engine.Module.
func (m *Module) Name() string { return "stub" }

// Scan implements engine.Module over repo targets.
func (m *Module) Scan(ctx context.Context, target engine.Target, emit func(finding.Candidate) error) error {
	if target.Kind != engine.TargetRepo {
		return fmt.Errorf("stub module supports only %q targets, got %q", engine.TargetRepo, target.Kind)
	}
	root := target.RepoPath
	if root == "" {
		return errors.New("stub module: repo target has an empty path")
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("stub module: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("stub module: %q is not a directory", root)
	}
	maxFile := m.MaxFileBytes
	if maxFile <= 0 {
		maxFile = defaultMaxFile
	}

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor":
				m.DirsSkippedSkipList++
				return fs.SkipDir
			}
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if int64(len(data)) > maxFile {
			m.FilesSkippedSize++
			return nil
		}
		if looksBinary(data) {
			m.FilesSkippedBinary++
			return nil
		}
		m.FilesScanned++

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		for lineNo, line := range splitLines(data) {
			for _, loc := range fakeRe.FindAllStringIndex(line, -1) {
				raw := line[loc[0]:loc[1]]
				if err := emit(finding.Candidate{
					Module:         "stub",
					RuleID:         RuleID,
					Type:           TypeName,
					Title:          "Stub fake key (M1 placeholder module)",
					Severity:       finding.SeverityHigh,
					Confidence:     0.9,
					Classification: finding.ClassificationUnknown,
					RawValue:       raw,
					Snippet:        windowSnippet(line, loc[0], loc[1]),
					Location: finding.Location{
						Path:         rel,
						Line:         lineNo + 1,
						StillPresent: true,
					},
					Impact:      "Placeholder module: real impact text arrives with the secrets rules engine in M2.",
					Remediation: "Placeholder module: real remediation text arrives with the secrets rules engine in M2.",
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// looksBinary reports whether the data starts like a binary file: any NUL
// byte in the first 8 KiB, or invalid UTF-8 overall.
func looksBinary(data []byte) bool {
	prefix := data
	if len(prefix) > binarySniffLen {
		prefix = prefix[:binarySniffLen]
	}
	if bytes.IndexByte(prefix, 0) >= 0 {
		return true
	}
	return !utf8.Valid(data)
}

// splitLines splits on \n and drops trailing \r (CRLF files), returning the
// logical lines without terminators.
func splitLines(data []byte) []string {
	s := string(data)
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1] // trailing newline is not an extra line
	}
	return lines
}

// windowSnippet renders a line around the match, capped at snippetMaxRunes,
// with ellipses marking elided context. Byte offsets are converted to rune
// offsets so multi-byte characters are not split.
func windowSnippet(line string, matchStart, matchEnd int) string {
	r := []rune(line)
	s := utf8.RuneCountInString(line[:matchStart])
	e := utf8.RuneCountInString(line[:matchEnd])

	start := s - snippetContext
	if start < 0 {
		start = 0
	}
	end := e + snippetContext
	if end > len(r) {
		end = len(r)
	}
	if end-start > snippetMaxRunes {
		end = start + snippetMaxRunes
	}
	out := string(r[start:end])
	if start > 0 {
		out = "…" + out
	}
	if end < len(r) {
		out += "…"
	}
	return out
}
