WASS Orchestrator Engine (Day 2)
Overview

This module handles the core execution pipeline for WASS.
It takes a repository URL, clones it, performs static analysis, builds a container image, and runs the application inside an isolated sandbox.

Pipeline Steps

Clone the repository

Run static scanning

Detect Dockerfile or use Dockerfile.sandbox

Build container image

Create isolated Docker network (scan-net)

Run container with security restrictions

Validate application startup

Files

orchestrator.sh – main pipeline script

static_scan.sh – basic static scan placeholder

Dockerfile.sandbox – fallback Dockerfile if none is found in the repo

Requirements

Docker installed

Git installed

Linux or WSL recommended

Usage
./orchestrator.sh <repo-url>

Notes

This is the Day-2 version of the orchestrator.
Functionality will expand in later stages to include dynamic scanning, result parsing, reporting, and dashboard integration.