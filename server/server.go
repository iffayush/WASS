// Package server is the WASS API server: REST API, SSE scan progress, queue
// workers, and persistence. It consumes the engine through the common finding
// schema and accepts only RedactedFinding past the engine boundary — raw
// secret values never leave the engine.
//
// M0 placeholder: API, queue, SSE, and persistence arrive in M7
// (docs/TASKS.md).
package server

// Version is the server semantic version.
const Version = "0.0.1-dev"
