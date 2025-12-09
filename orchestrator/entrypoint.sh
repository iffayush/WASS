#!/usr/bin/env sh
# orchestrator/entrypoint.sh
set -e

START_TIMEOUT="${START_TIMEOUT:-20}"
APP_DIR="/app"

echo "[sandbox] entrypoint: checking for package.json in $APP_DIR"

if [ -f "$APP_DIR/package.json" ]; then
  # Read package.json to know if a start script exists
  has_start=$(node -e "console.log(require('$APP_DIR/package.json').scripts && require('$APP_DIR/package.json').scripts.start ? 'yes' : 'no')" 2>/dev/null || echo "no")
  echo "[sandbox] package.json detected. has_start=$has_start"

  if [ "$has_start" = "yes" ]; then
    echo "[sandbox] Installing dependencies (production)..."
    # Use npm ci if lockfile exists
    if [ -f "$APP_DIR/package-lock.json" ] || [ -f "$APP_DIR/yarn.lock" ]; then
      (cd "$APP_DIR" && npm ci --silent --omit=dev) || echo "[sandbox] npm ci failed (continuing)"
    else
      (cd "$APP_DIR" && npm install --silent --omit=dev) || echo "[sandbox] npm install failed (continuing)"
    fi

    echo "[sandbox] Starting app with npm start..."
    # start app; any logs will be captured by docker logs
    (cd "$APP_DIR" && npm start) &
    APP_PID=$!

    # wait up to START_TIMEOUT seconds and then exit if process detached?
    sleep "$START_TIMEOUT"
    if ps -p $APP_PID > /dev/null 2>&1; then
      echo "[sandbox] App process $APP_PID is running after $START_TIMEOUT seconds. Leaving container running for log inspection."
      # keep tailing logs (read-only filesystem might forbid logs in some apps)
      wait $APP_PID
      exit $?
    else
      echo "[sandbox] App process exited before timeout."
      wait $APP_PID || true
      exit 0
    fi

  else
    echo "[sandbox] No start script found in package.json. Nothing to run. Exiting."
    exit 0
  fi
else
  echo "[sandbox] No package.json found. Nothing to run. Exiting."
  exit 0
fi
