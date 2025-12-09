#!/usr/bin/env bash
# orchestrator/orchestrator.sh
# Usage: ./orchestrator.sh <git-repo-url> [branch_or_ref]
set -euo pipefail

REPO_URL="${1:-}"
BRANCH="${2:-main}"
WORKDIR="${WORKDIR:-$(mktemp -d)}"
SCAN_NET="scan-net"
SANDBOX_DOCKERFILE="${SANDBOX_DOCKERFILE:-./orchestrator/Dockerfile.sandbox}"
TOOLS_ROOT="$(cd "$(dirname "$0")/.." && pwd)"   # repo root (one level up from orchestrator/)
SCANNER="${TOOLS_ROOT}/tools/repo_check.js"
MEMORY_LIMIT="${MEMORY_LIMIT:-256m}"   # default memory limit for sandbox
CPU_LIMIT="${CPU_LIMIT:-0.5}"          # default CPU limit (0.5 CPUs)
START_TIMEOUT="${START_TIMEOUT:-20}"   # seconds to wait for app to start
CONTAINER_NAME="wass-sandbox-$(date +%s)"
IMAGE_TAG="wass-sandbox:latest"

if [ -z "$REPO_URL" ]; then
  echo "Usage: $0 <git-repo-url> [branch_or_ref]"
  exit 2
fi

echo "Orchestrator starting..."
echo "Working directory: $WORKDIR"
mkdir -p "$WORKDIR"
cd "$WORKDIR"

# 1) Clone
echo "Cloning $REPO_URL (branch: $BRANCH) ..."
git clone --depth 1 --branch "$BRANCH" "$REPO_URL" repo || {
  echo "Shallow clone failed, trying full clone..."
  rm -rf repo
  git clone "$REPO_URL" repo
}
TARGET_DIR="$WORKDIR/repo"
echo "Cloned to $TARGET_DIR"

# 2) Static scan
if [ -f "$SCANNER" ]; then
  echo "Running static scanner on cloned repo..."
  node "$SCANNER" "$TARGET_DIR" || {
    echo "Scanner exited with non-zero code (possible findings). Check scan-report.json in ${WORKDIR}."
  }
  # move scan report into workdir for record
  if [ -f "$TARGET_DIR/scan-report.json" ]; then
    mv "$TARGET_DIR/scan-report.json" "$WORKDIR/scan-report.json"
  elif [ -f "$TOOLS_ROOT/scan-report.json" ]; then
    # if scanner wrote to tools root
    cp "$TOOLS_ROOT/scan-report.json" "$WORKDIR/scan-report.json" || true
  fi
  echo "Scan report saved to $WORKDIR/scan-report.json (if present)."
else
  echo "Scanner not found at $SCANNER — skipping static scan."
fi

# 3) Choose Dockerfile
if [ -f "$TARGET_DIR/Dockerfile" ]; then
  DOCKERFILE_PATH="$TARGET_DIR/Dockerfile"
  echo "Found repository Dockerfile: $DOCKERFILE_PATH"
else
  DOCKERFILE_PATH="$SANDBOX_DOCKERFILE"
  echo "Using default sandbox Dockerfile: $DOCKERFILE_PATH"
fi

# 4) Build image
echo "Building Docker image ($IMAGE_TAG) from $DOCKERFILE_PATH ..."
docker build -t "$IMAGE_TAG" -f "$DOCKERFILE_PATH" "$TARGET_DIR" || {
  echo "Image build failed. Aborting."
  exit 3
}

# 5) Ensure network exists
if ! docker network ls --format '{{.Name}}' | grep -q "^${SCAN_NET}$"; then
  echo "Creating isolated Docker network: $SCAN_NET"
  docker network create "$SCAN_NET" >/dev/null
else
  echo "Docker network $SCAN_NET already exists."
fi

# 6) Run container in sandbox
echo "Running container $CONTAINER_NAME with strict limits..."
# create a read-only mount of the repo into /app (image's entrypoint handles running)
docker run --rm --name "$CONTAINER_NAME" \
  --network "$SCAN_NET" \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --memory "$MEMORY_LIMIT" \
  --cpus "$CPU_LIMIT" \
  --pids-limit 100 \
  --read-only \
  -v "$TARGET_DIR":/app:ro \
  -v "$WORKDIR/tmp":/tmp:rw \
  -w /app \
  -e "START_TIMEOUT=${START_TIMEOUT}" \
  -d "$IMAGE_TAG" >/dev/null

# Wait for startup and collect logs
echo "Waiting up to ${START_TIMEOUT}s for container to start..."
SECS=0
started=0
while [ $SECS -lt "$START_TIMEOUT" ]; do
  sleep 1
  SECS=$((SECS+1))
  # check if container still exists
  if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Container exited early (exit within ${SECS}s). Collecting logs..."
    docker logs "$CONTAINER_NAME" 2>/dev/null || true
    break
  fi
  # look for 'listening' or 'started' keywords in logs as a heuristic
  logs="$(docker logs --since ${SECS}s "$CONTAINER_NAME" 2>/dev/null || true)"
  if echo "$logs" | grep -Ei "listening|started|server running|http server" >/dev/null; then
    started=1
    echo "Detected startup hint in logs."
    break
  fi
done

# collect logs and status
mkdir -p "$WORKDIR/artifacts"
echo "Collecting container logs..."
docker logs "$CONTAINER_NAME" > "$WORKDIR/artifacts/container.log" 2>&1 || true
docker inspect "$CONTAINER_NAME" > "$WORKDIR/artifacts/container.inspect.json" 2>&1 || true

# 7) Report outcome
if [ "$started" -eq 1 ]; then
  echo "Sandbox run: APP STARTED (heuristic). Logs saved to $WORKDIR/artifacts/container.log"
else
  echo "Sandbox run: No startup detected within ${START_TIMEOUT}s. Check logs: $WORKDIR/artifacts/container.log"
fi

# 8) Cleanup: stop container if still running (run was --rm so it will auto-remove on exit)
if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
  echo "Stopping container..."
  docker stop "$CONTAINER_NAME" >/dev/null || true
fi

echo "Orchestrator finished. Workdir: $WORKDIR"
echo "Artifacts: $WORKDIR/artifacts , scan-report: $WORKDIR/scan-report.json (if present)"
