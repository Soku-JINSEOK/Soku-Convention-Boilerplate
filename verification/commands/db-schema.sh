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

RUN_MYSQL=false
RUN_POSTGRES=false
scope_selected "mysql" && [[ -f "$MYSQL_SCHEMA" ]] && RUN_MYSQL=true
scope_selected "postgresql" && [[ -f "$POSTGRES_SCHEMA" ]] && RUN_POSTGRES=true

if [[ "$RUN_MYSQL" == false && "$RUN_POSTGRES" == false ]]; then
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
services=()
[[ "$RUN_MYSQL" == true ]] && services+=(mysql)
[[ "$RUN_POSTGRES" == true ]] && services+=(postgres)
run_or_fail "db-schema::up" 101 \
  docker compose -f "$COMPOSE_FILE" up -d --wait "${services[@]}"
print_step_end

if [[ "$RUN_MYSQL" == true ]]; then
  print_step "Loading MySQL schema"
  # shellcheck disable=SC2016 # $1/$2 are expanded by the sub-shell, not here,
  # so paths are passed as arguments rather than interpolated into the string.
  run_or_fail "db-schema::mysql" 102 bash -c \
    'docker compose -f "$1" exec -T mysql mysql -uroot template_check < "$2"' \
    _ "$COMPOSE_FILE" "$MYSQL_SCHEMA"
  print_step_end
fi

if [[ "$RUN_POSTGRES" == true ]]; then
  print_step "Loading PostgreSQL schema"
  # shellcheck disable=SC2016 # $1/$2 are expanded by the sub-shell, not here,
  # so paths are passed as arguments rather than interpolated into the string.
  run_or_fail "db-schema::postgresql" 103 bash -c \
    'PGPASSWORD=postgres docker compose -f "$1" exec -T postgres psql -U postgres -d template_check < "$2"' \
    _ "$COMPOSE_FILE" "$POSTGRES_SCHEMA"
  print_step_end
fi
