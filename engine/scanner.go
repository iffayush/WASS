package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iffayush/wass/engine/finding"
)

// Options configures a scan run.
type Options struct {
	// FingerprintKey is the HMAC key for finding fingerprints. It is
	// required — the engine has no built-in default key (D-008). The CLI
	// takes it from --fingerprint-key or WASS_FINGERPRINT_KEY; the server
	// passes its own so CLI and server fingerprints agree.
	FingerprintKey []byte
	// FingerprintKeyVersion is stored alongside every fingerprint (D-008).
	FingerprintKeyVersion int
}

func (o Options) validate() error {
	if len(o.FingerprintKey) == 0 {
		return errors.New("scan options: fingerprint key is required (no built-in default key, D-008)")
	}
	if o.FingerprintKeyVersion < 1 {
		return errors.New("scan options: fingerprint key version must be >= 1")
	}
	return nil
}

// Report is the result of one scan run: the versioned envelope plus the
// deduplicated findings. Findings carry no raw secret values by construction.
type Report struct {
	SchemaVersion int               `json:"schema_version"`
	Target        Target            `json:"target"`
	StartedAt     time.Time         `json:"started_at"`
	FinishedAt    time.Time         `json:"finished_at"`
	ModulesRun    []string          `json:"modules"`
	FindingsCount int               `json:"findings_count"`
	Findings      []finding.Finding `json:"findings"`
}

// Redacted is the one-way projection of a Report into its publishable form.
// Everything that leaves the engine — CLI output, server persistence,
// exports — is built from RedactedReport.
func (r Report) Redacted() RedactedReport {
	out := RedactedReport{
		SchemaVersion: r.SchemaVersion,
		Target:        r.Target,
		StartedAt:     r.StartedAt,
		FinishedAt:    r.FinishedAt,
		ModulesRun:    r.ModulesRun,
		FindingsCount: r.FindingsCount,
		Findings:      make([]finding.RedactedFinding, 0, len(r.Findings)),
	}
	for _, f := range r.Findings {
		out.Findings = append(out.Findings, f.Redacted())
	}
	return out
}

// RedactedReport is the public projection of Report. The CLI, server, and
// every exporter accept only this type (compile-time redaction boundary).
type RedactedReport struct {
	SchemaVersion int                       `json:"schema_version"`
	Target        Target                    `json:"target"`
	StartedAt     time.Time                 `json:"started_at"`
	FinishedAt    time.Time                 `json:"finished_at"`
	ModulesRun    []string                  `json:"modules"`
	FindingsCount int                       `json:"findings_count"`
	Findings      []finding.RedactedFinding `json:"findings"`
}

// Scanner runs registered modules over a target and assembles the report.
type Scanner struct {
	modules []Module
}

// NewScanner returns a scanner for the given modules, in registration order.
func NewScanner(modules ...Module) *Scanner {
	return &Scanner{modules: modules}
}

// Scan runs every module against the target and returns the deduplicated
// report. Conversion from Candidate to Finding (fingerprint + redaction)
// happens here and only here; modules never see Findings.
func (s *Scanner) Scan(ctx context.Context, target Target, opts Options) (Report, error) {
	if err := opts.validate(); err != nil {
		return Report{}, err
	}
	if target.Kind != TargetRepo && target.Kind != TargetDomain {
		return Report{}, fmt.Errorf("scan: unknown target kind %q", target.Kind)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	report := Report{
		SchemaVersion: finding.SchemaVersion,
		Target:        target,
		StartedAt:     time.Now().UTC(),
		Findings:      nil,
	}

	type dedupEntry struct {
		finding   finding.Finding
		locations []finding.Location
	}
	dedup := map[string]*dedupEntry{}
	var order []string // dedup keys in first-seen order

	for _, m := range s.modules {
		if err := ctx.Err(); err != nil {
			return Report{}, err
		}
		report.ModulesRun = append(report.ModulesRun, m.Name())

		emit := func(c finding.Candidate) error {
			f, err := finding.FromCandidate(c, opts.FingerprintKey, opts.FingerprintKeyVersion)
			if err != nil {
				return fmt.Errorf("module %s: %w", m.Name(), err)
			}
			if e, ok := dedup[f.DedupKey()]; ok {
				e.locations = appendLocationIfNew(e.locations, f.Locations...)
			} else {
				dedup[f.DedupKey()] = &dedupEntry{finding: f, locations: f.Locations}
				order = append(order, f.DedupKey())
			}
			return nil
		}
		if err := m.Scan(ctx, target, emit); err != nil {
			return Report{}, fmt.Errorf("module %s: %w", m.Name(), err)
		}
	}

	for i, key := range order {
		e := dedup[key]
		f := e.finding
		f.ID = fmt.Sprintf("f-%06d", i+1)
		f.Locations = e.locations
		if err := f.Validate(); err != nil {
			return Report{}, fmt.Errorf("internal: assembled finding failed validation: %w", err)
		}
		report.Findings = append(report.Findings, f)
	}
	report.FindingsCount = len(report.Findings)
	report.FinishedAt = time.Now().UTC()
	return report, nil
}

// appendLocationIfNew appends locations not already present (same path/url,
// line, commit, branch identity).
func appendLocationIfNew(locs []finding.Location, add ...finding.Location) []finding.Location {
	for _, a := range add {
		dup := false
		for _, l := range locs {
			if l.Path == a.Path && l.URL == a.URL && l.Line == a.Line &&
				l.Commit == a.Commit && l.Branch == a.Branch {
				dup = true
				break
			}
		}
		if !dup {
			locs = append(locs, a)
		}
	}
	return locs
}
