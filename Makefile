# WASS — canonical build entry point.
#
# Windows (no make installed): run the underlying commands directly.
#   make test   →  go test ./...
#   make vet    →  go vet ./...
#   make build  →  go build -o bin/ ./cmd/wass   (add .exe on Windows)
#   make secret-scan  →  gitleaks detect --no-banner --redact
#   make hooks  →  git config core.hooksPath .githooks
#   make compose-up   →  docker compose up -d
# See docs/DECISIONS.md D-010.

.PHONY: test vet lint build secret-scan hooks compose-up compose-down ci

test:
	go test ./...

vet:
	go vet ./...

# golangci-lint is added to CI with M2 (no config yet at M0).
lint:
	golangci-lint run

build:
	go build -o bin/ ./cmd/wass

# Fail-closed secret scan over the whole history and working tree.
secret-scan:
	gitleaks detect --no-banner --redact

# Enable the fail-closed pre-commit hook (gitleaks). One-time setup.
hooks:
	git config core.hooksPath .githooks

compose-up:
	docker compose up -d

compose-down:
	docker compose down

# What CI runs.
ci: vet test
