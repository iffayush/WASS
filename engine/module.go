package engine

import (
	"context"

	"github.com/iffayush/wass/engine/finding"
)

// TargetKind distinguishes what a scan examines.
type TargetKind string

const (
	// TargetRepo is a git repository (working tree; history arrives in M3).
	TargetRepo TargetKind = "repo"
	// TargetDomain is a deployed web app (crawler arrives in M4).
	TargetDomain TargetKind = "domain"
)

// Target is what the scanner runs against. Repo mode sets RepoPath to a
// local directory; domain mode sets URL (M4).
type Target struct {
	Kind     TargetKind `json:"kind"`
	RepoPath string     `json:"repo_path,omitempty"`
	URL      string     `json:"url,omitempty"`
}

// Describe returns the human/machine location string for the target.
func (t Target) Describe() string {
	if t.Kind == TargetDomain {
		return t.URL
	}
	return t.RepoPath
}

// Module is one check that examines a target and emits candidate findings.
// The engine ships three real modules over the milestones (secrets: M2,
// exposed files + headers: M4/M5); future modules plug in the same way. The
// API, UI, report, and exports only know the common finding schema — never
// module internals.
//
// Modules emit [finding.Candidate] values; they never construct
// [finding.Finding] themselves. The scanner converts candidates behind the
// redaction boundary (fingerprinting + evidence redaction happen in exactly
// one place, engine/finding.FromCandidate). Raw values in candidates exist
// in memory only and are never serialized.
//
// Scan must honor ctx cancellation, must not execute anything from the
// target, and must not make outbound network requests (that is fetcher
// territory, M4, behind the SSRF guard).
type Module interface {
	// Name returns the module identifier used in findings and coverage data.
	Name() string
	// Scan examines the target and emits candidates to emit. Returning an
	// error aborts the whole scan.
	Scan(ctx context.Context, target Target, emit func(finding.Candidate) error) error
}
