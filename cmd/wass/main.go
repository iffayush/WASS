// Command wass is the WASS CLI.
//
//	wass scan --repo <url-or-path> [--fingerprint-key <key>]
//	wass scan --domain <url>    (M4)
//
// It emits the scan report as JSON on stdout. See docs/SPEC.md and docs/TASKS.md.
package main

import (
	"os"
)

func main() {
	os.Exit(cliMain(os.Args[1:]))
}
