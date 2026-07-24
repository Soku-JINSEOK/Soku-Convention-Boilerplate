#!/usr/bin/env bash
# Thin wrapper kept alive for existing muscle memory and any external
# references — the actual checks now live behind scripts/verify.sh's `full`
# profile (see verification/CLASSIFICATION.md for what that covers and
# verification/commands/*.sh for the individual check groups).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCRIPT_DIR/verify.sh" --profile full "$@"
