// Command wass is the WASS CLI.
//
// From M1 it provides:
//
//	wass scan --repo <url-or-path> [--fingerprint-key <key>]
//	wass scan --domain <url>    [--fingerprint-key <key>]
//
// emitting findings as JSON on stdout. See docs/TASKS.md and docs/SPEC.md.
package main

import (
	"fmt"
	"os"

	"github.com/iffayush/wass/engine"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println("wass", engine.Version)
		return
	}
	fmt.Fprintln(os.Stderr, "wass", engine.Version, "— the scan command arrives in M1 (docs/TASKS.md)")
	os.Exit(2)
}
