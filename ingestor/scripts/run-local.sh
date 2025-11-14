#!/usr/bin/env bash
set -euo pipefail

# Simple helper to run Mosquitto and the ingestor binary locally.
# - Starts a Docker container named `runsense-mosquitto` if not present.
# - Exports sensible default env vars (you can override them before running).
# - Builds the binary if missing and runs it in foreground.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="${SCRIPT_DIR}/.."

echo "=== runsense: local runner ==="

# Start mosquitto container if not running
if ! docker ps --filter "name=runsense-mosquitto" --filter "status=running" --format "{{.Names}}" | grep -q runsense-mosquitto; then
  if docker ps -a --filter "name=runsense-mosquitto" --format "{{.Names}}" | grep -q runsense-mosquitto; then
    echo "Starting existing container runsense-mosquitto..."
    docker start runsense-mosquitto
  else
    echo "Creating and starting runsense-mosquitto (eclipse-mosquitto)..."
    docker run -d --name runsense-mosquitto -p 1883:1883 eclipse-mosquitto:2
  fi
else
  echo "runsense-mosquitto already running"
fi

echo
echo "Starting Postgres (runsense-postgres) via Docker if needed"
echo

# Default envs (override by exporting beforehand)
export MQTT_HOST=${MQTT_HOST:-localhost}
export MQTT_PORT=${MQTT_PORT:-1883}
export MQTT_TOPICS=${MQTT_TOPICS:-#}
export MQTT_CLIENT_ID=${MQTT_CLIENT_ID:-ingestor-local}
export MQTT_KEEPALIVE=${MQTT_KEEPALIVE:-60}
export MQTT_QOS=${MQTT_QOS:-0}
export MQTT_CLEAN_SESSION=${MQTT_CLEAN_SESSION:-true}

export PG_HOST=${PG_HOST:-localhost}
export PG_PORT=${PG_PORT:-5432}
export PG_DB=${PG_DB:-runsense}
export PG_USER=${PG_USER:-postgres}
export PG_PASSWORD=${PG_PASSWORD:-postgres}

# Start Postgres container if not running
if ! docker ps --filter "name=runsense-postgres" --filter "status=running" --format "{{.Names}}" | grep -q runsense-postgres; then
  if docker ps -a --filter "name=runsense-postgres" --format "{{.Names}}" | grep -q runsense-postgres; then
    echo "Starting existing container runsense-postgres..."
    docker start runsense-postgres
  else
    echo "Creating and starting runsense-postgres (postgres:15)..."
    # Mount repo init.sql into docker-entrypoint-initdb.d if present. This runs only on first initialization.
    INIT_SQL_PATH="$ROOT_DIR/k8s/base/postgres/init.sql"
    if [ -f "$INIT_SQL_PATH" ]; then
      echo "Found init.sql, mounting into container so Postgres will initialize schema"
      docker run -d \
        --name runsense-postgres \
        -p 5432:5432 \
        -e POSTGRES_DB="$PG_DB" \
        -e POSTGRES_USER="$PG_USER" \
        -e POSTGRES_PASSWORD="$PG_PASSWORD" \
        -v "$INIT_SQL_PATH":/docker-entrypoint-initdb.d/init.sql:ro \
        postgres:15
    else
      docker run -d \
        --name runsense-postgres \
        -p 5432:5432 \
        -e POSTGRES_DB="$PG_DB" \
        -e POSTGRES_USER="$PG_USER" \
        -e POSTGRES_PASSWORD="$PG_PASSWORD" \
        postgres:15
    fi
  fi
else
  echo "runsense-postgres already running"
fi

# Wait for Postgres to be ready (up to ~30s)
echo "Waiting for Postgres to become ready..."
READY=0
for i in $(seq 1 30); do
  if docker exec runsense-postgres pg_isready -U "$PG_USER" -d "$PG_DB" >/dev/null 2>&1; then
    READY=1
    break
  fi
  sleep 1
done
if [ "$READY" -ne 1 ]; then
  echo "Warning: Postgres did not become ready in time. You may still try running the ingestor, but /health may fail."
else
  echo "Postgres is ready"
fi

export RAW_TABLE=${RAW_TABLE:-raw_windows}
export FEAT_TABLE=${FEAT_TABLE:-feat_windows}

export HTTP_PORT=${HTTP_PORT:-8080}

echo "Using envs: MQTT_HOST=$MQTT_HOST MQTT_PORT=$MQTT_PORT HTTP_PORT=$HTTP_PORT PG_HOST=$PG_HOST PG_PORT=$PG_PORT"

# Build binary if missing
BIN="$ROOT_DIR/ingestor"
if [ ! -f "$BIN" ]; then
  echo "Building ingestor binary..."
  (cd "$ROOT_DIR" && go build -o ingestor .)
fi

echo "Starting ingestor (foreground). Press Ctrl-C to stop."
exec "$BIN"
