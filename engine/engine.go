// Package engine is the WASS scanning engine: the scanner core, check
// modules (secrets, exposed files, headers/cookies), fetchers, the finding
// schema, and redaction. The engine is standalone — it never depends on the
// server or the web frontend; the server wraps the engine.
//
// M0 placeholder: the finding schema, module interface, and CLI scan command
// arrive in M1 (docs/TASKS.md).
package engine

// Version is the engine semantic version, reported by the CLI.
const Version = "0.0.1-dev"
