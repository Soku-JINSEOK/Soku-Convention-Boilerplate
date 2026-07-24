#!/usr/bin/env bash
# Loads templates/mysql/schema.sql and templates/postgresql/schema.sql
# against real MySQL/PostgreSQL services started from
# docker-compose.verify.yml. Mirrors the mysql/postgresql jobs in
# .github/workflows/templates-ci.yml, which today only run hosted (via
# GitHub Actions service containers) — this is the new local coverage
# issue #112 calls out as missing.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="${WORKSPACE:?WORKSPACE must be set}"
# shellcheck source=verification/commands/_lib.sh
source "$SCRIPT_DIR/_lib.sh"

MYSQL_SCHEMA="$WORKSPACE/templates/mysql/schema.sql"
POSTGRES_SCHEMA="$WORKSPACE/templates/postgresql/schema.sql"
COMPOSE_FILE="$WORKSPACE/docker-compose.verify.yml"

if [[ ! -f "$MYSQL_SCHEMA" && ! -f "$POSTGRES_SCHEMA" ]]; then
  echo "No DB schema files found — skipping"
  exit 0
fi

require_command docker "db-schema" 100
if ! docker compose version >/dev/null 2>&1; then
  action_fail 100 "db-schema" "docker compose plugin is required"
fi

cleanup() {
  docker compose -f "$COMPOSE_FILE" down --volumes >/dev/null 2>&1 || true
}
trap cleanup EXIT

print_step "Starting local MySQL/PostgreSQL services"
run_or_fail "db-schema::up" 101 docker compose -f "$COMPOSE_FILE" up -d --wait
print_step_end

if [[ -f "$MYSQL_SCHEMA" ]]; then
  print_step "Loading MySQL schema"
  # shellcheck disable=SC2016 # $1/$2 are expanded by the sub-shell, not here,
  # so paths are passed as arguments rather than interpolated into the string.
  run_or_fail "db-schema::mysql" 102 bash -c \
    'docker compose -f "$1" exec -T mysql mysql -uroot template_check < "$2"' \
    _ "$COMPOSE_FILE" "$MYSQL_SCHEMA"
  print_step_end
fi

if [[ -f "$POSTGRES_SCHEMA" ]]; then
  print_step "Loading PostgreSQL schema"
  # shellcheck disable=SC2016 # $1/$2 are expanded by the sub-shell, not here,
  # so paths are passed as arguments rather than interpolated into the string.
  run_or_fail "db-schema::postgresql" 103 bash -c \
    'PGPASSWORD=postgres docker compose -f "$1" exec -T postgres psql -U postgres -d template_check < "$2"' \
    _ "$COMPOSE_FILE" "$POSTGRES_SCHEMA"
  print_step_end
fi
