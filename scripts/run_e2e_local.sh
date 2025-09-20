#!/usr/bin/env zsh
# Helper to run the Experia E2E test locally by loading environment variables
# from a local file (not checked into VCS). The default file is `.experia.env`.
#
# Usage:
#   ./scripts/run_e2e_local.sh            # uses .experia.env in repo root
#   ./scripts/run_e2e_local.sh -e /path/to/file
#
# Example .experia.env contents (keep this file out of git):
# EXPERIA_E2E=1
# EXPERIA_IP=192.168.2.254
# EXPERIA_USER=admin
# EXPERIA_PASS=yourpassword
# EXPERIA_EXPECT_CONNECTION_STATE=Connected

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$ROOT_DIR/.experia.env"

while getopts "e:" opt; do
  case $opt in
    e) ENV_FILE="$OPTARG" ;;
    *) echo "Usage: $0 [-e envfile]"; exit 1 ;;
  esac
done

if [ ! -f "$ENV_FILE" ]; then
  echo "Env file not found: $ENV_FILE"
  echo "Create a file with these variables and keep it out of git:"
  echo "  EXPERIA_E2E=1"
  echo "  EXPERIA_IP=192.168.2.254"
  echo "  EXPERIA_USER=admin"
  echo "  EXPERIA_PASS=secret"
  exit 2
fi

# shellcheck disable=SC1090
source "$ENV_FILE"

if [ "${EXPERIA_E2E:-}" != "1" ]; then
  echo "EXPERIA_E2E not set to 1 in $ENV_FILE; set it to enable the test"
  exit 3
fi

echo "Running Experia E2E test using env file: $ENV_FILE"
echo "IP: ${EXPERIA_IP:-<unset>}"
echo "USER: ${EXPERIA_USER:-<unset>}"
echo "EXPECTED CONNECTION STATE: ${EXPERIA_EXPECT_CONNECTION_STATE:-Connected}"

# Mask password in printed output
echo "PASSWORD: (will not be printed)"

# Export variables from the env file safely (ignores comments). This avoids
# invoking `env` with raw file contents which can break on comments.
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

exec go test ./internal/collector -run TestE2E -v
