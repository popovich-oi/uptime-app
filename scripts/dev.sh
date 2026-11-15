#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_PID=""
FRONTEND_PID=""

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker is required to start PostgreSQL. Start PostgreSQL manually before continuing." >&2
else
  cd "$ROOT_DIR"
  if ! docker info >/dev/null 2>&1; then
    if [[ "$(uname -s)" == "Darwin" ]] && command -v open >/dev/null 2>&1; then
      echo "Starting Docker Desktop..."
      open -a Docker >/dev/null 2>&1 || true
    fi
    for _ in {1..60}; do
      docker info >/dev/null 2>&1 && break
      sleep 1
    done
    if ! docker info >/dev/null 2>&1; then
      echo "Docker daemon is not running. Start Docker Desktop and run this script again." >&2
      exit 1
    fi
  fi
  POSTGRES_PORT="${POSTGRES_PORT:-5432}"
  USE_EXISTING_POSTGRES=false
  if command -v pg_isready >/dev/null 2>&1 && pg_isready -h localhost -p "$POSTGRES_PORT" -U "${POSTGRES_USER:-uptime}" >/dev/null 2>&1; then
    USE_EXISTING_POSTGRES=true
    echo "Using PostgreSQL already running on localhost:${POSTGRES_PORT}."
  fi
  if [[ "$USE_EXISTING_POSTGRES" == false ]]; then
    while command -v nc >/dev/null 2>&1 && nc -z 127.0.0.1 "$POSTGRES_PORT" >/dev/null 2>&1; do
      POSTGRES_PORT=$((POSTGRES_PORT + 1))
    done
    export POSTGRES_PORT
    echo "Starting PostgreSQL on localhost:${POSTGRES_PORT}..."
    docker compose up -d postgres
    until docker compose exec -T postgres pg_isready -U "${POSTGRES_USER:-uptime}" -d "${POSTGRES_DB:-uptime}" >/dev/null 2>&1; do
      sleep 1
    done
  fi
  if [[ -z "${DATABASE_URL:-}" ]]; then
    export DATABASE_URL="host=localhost user=${POSTGRES_USER:-uptime} password=${POSTGRES_PASSWORD:-uptime} dbname=${POSTGRES_DB:-uptime} port=${POSTGRES_PORT} sslmode=disable"
  fi
fi

cleanup() {
  trap - INT TERM EXIT
  [[ -z "$BACKEND_PID" ]] || kill "$BACKEND_PID" 2>/dev/null || true
  [[ -z "$FRONTEND_PID" ]] || kill "$FRONTEND_PID" 2>/dev/null || true
  [[ -z "$BACKEND_PID" ]] || wait "$BACKEND_PID" 2>/dev/null || true
  [[ -z "$FRONTEND_PID" ]] || wait "$FRONTEND_PID" 2>/dev/null || true
}

trap cleanup INT TERM EXIT

(
  cd "$ROOT_DIR/backend"
  exec go run ./cmd/server
) &
BACKEND_PID=$!

(
  cd "$ROOT_DIR/frontend"
  exec npm run dev
) &
FRONTEND_PID=$!

while kill -0 "$BACKEND_PID" 2>/dev/null && kill -0 "$FRONTEND_PID" 2>/dev/null; do
  sleep 1
done

exit 1
