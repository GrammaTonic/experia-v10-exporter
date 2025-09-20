#!/usr/bin/env zsh
# Simple script to POST a NeMo getMIBs request to an Experia V10 device.
# Supports optional authentication (createContext) to obtain a contextID and
# cookie jar which curl will reuse for subsequent requests.

set -euo pipefail
# Avoid inheriting OS or shell-exported user variables that may exist in the
# parent environment (for example $USER or $USERNAME). Unset them so the
# script uses its own configured values or CLI-provided overrides.
unset USER USERNAME 2>/dev/null || true

# If a .env file exists in the repo root, source it so repo-level config can
# provide credentials. This keeps credentials out of the script itself.
ENV_FILE="$(dirname -- "$0")/../.env"
if [[ -f "$ENV_FILE" ]]; then
  # shellcheck source=/dev/null
  source "$ENV_FILE"
fi

HOST="${EXPERIA_V10_ROUTER_IP:-192.168.2.254}"
CANDIDATE="ETH0"
TOKEN=""
# Prefer EXPERIA_V10_ROUTER_USERNAME/PASSWORD from .env, allow CLI override
ROUTER_USERNAME="${EXPERIA_V10_ROUTER_USERNAME:-}"
ROUTER_PASSWORD="${EXPERIA_V10_ROUTER_PASSWORD:-}"
# By default do not auto-authenticate unless requested
AUTH=0
VERBOSE=0

show_help() {
  cat <<-EOF
Usage: $(basename $0) [-h host] [-c candidate] [-t token] [-v]

Options:
  -h host       Host/IP of the router (default: 127.0.0.1)
  -c candidate  Interface candidate (e.g. ETH0) (default: ETH0)
  -t token      Session token (X-Sah / x-context header value)
  -u username   Username to authenticate (triggers auth when -p provided or -a)
  -p password   Password to authenticate
  -a            Authenticate first (perform createContext) and use returned token
  -e            Source credentials from repo .env (if present). CLI flags override.
  -d data       Inline JSON body to send (overrides -c/-f)
  -f file       Read JSON body from file (overrides -c)
  -v            Verbose output (show curl command and response)

Example:
  $(basename $0) -h 192.168.2.254 -c ETH0 -t ABCDE12345
EOF
}

BODY_FILE=""
BODY_INLINE=""
while getopts ":h:c:t:u:p:avd:f:" opt; do
  case ${opt} in
    h) HOST=${OPTARG} ;;
    c) CANDIDATE=${OPTARG} ;;
    d) BODY_INLINE=${OPTARG} ;;
    f) BODY_FILE=${OPTARG} ;;
    t) TOKEN=${OPTARG} ;;
  u) ROUTER_USERNAME=${OPTARG} ;;
  p) ROUTER_PASSWORD=${OPTARG} ;;
    a) AUTH=1 ;;
    e) 
      # re-source .env explicitly if -e passed (already sourced if present)
      if [[ -f "$ENV_FILE" ]]; then
        # shellcheck source=/dev/null
        source "$ENV_FILE"
        ROUTER_USERNAME="${EXPERIA_V10_ROUTER_USERNAME:-$ROUTER_USERNAME}"
        ROUTER_PASSWORD="${EXPERIA_V10_ROUTER_PASSWORD:-$ROUTER_PASSWORD}"
      fi
      ;;
    v) VERBOSE=1 ;;
    *) show_help; exit 1 ;;
  esac
done

# Build request body: precedence CLI inline (-d) > file (-f) > candidate-based default (-c)
if [[ -n "$BODY_INLINE" ]]; then
  BODY="$BODY_INLINE"
elif [[ -n "$BODY_FILE" ]]; then
  if [[ -f "$BODY_FILE" ]]; then
    BODY=$(cat "$BODY_FILE")
  else
    echo "Body file not found: $BODY_FILE" >&2
    exit 2
  fi
else
  BODY=$(printf '{"service":"NeMo.Intf.%s","method":"getMIBs","parameters":{}}' "$CANDIDATE")
fi
URL="http://${HOST}/ws/NeMo/Intf/lan:getMIBs"

# prepare a temporary cookie jar for optional auth flow
COOKIEJAR="$(mktemp -t curl_getmibs_cookies.XXXX)
"

HEADER_ARGS=(
  -H "accept: */*"
  -H "accept-language: en-US,en;q=0.7"
  -H "content-type: application/x-sah-ws-4-call+json"
  -H "sec-gpc: 1"
  -H "Origin: http://${HOST}"
  -H "Referer: http://${HOST}/"
)

if [[ -n "$TOKEN" ]]; then
  HEADER_ARGS+=( -H "Authorization: X-Sah ${TOKEN}" -H "x-context: ${TOKEN}" )
fi

# If requested, perform authentication first to obtain contextID and cookies
# Ensure we prefer the EXPRIA_* vars from .env for auth if present to avoid
# accidentally using any remaining inherited shell variables.
ROUTER_USERNAME="${EXPERIA_V10_ROUTER_USERNAME:-$ROUTER_USERNAME}"
ROUTER_PASSWORD="${EXPERIA_V10_ROUTER_PASSWORD:-$ROUTER_PASSWORD}"
if [[ $AUTH -eq 1 || ( -n "$ROUTER_USERNAME" && -n "$ROUTER_PASSWORD" ) ]]; then
  # createContext payload
  AUTH_BODY=$(printf '{"service":"sah.Device.Information","method":"createContext","parameters":{"applicationName":"webui","username":"%s","password":"%s"}}' "$ROUTER_USERNAME" "$ROUTER_PASSWORD")
  AUTH_URL="http://${HOST}/ws/NeMo/Intf/lan:getMIBs"
  # auth uses X-Sah-Login header on initial POST
  if (( VERBOSE )); then
    echo "AUTH POST $AUTH_URL"
    echo "AUTH BODY: $AUTH_BODY"
  fi
  # perform auth and save cookies; capture response
  AUTH_RESP=$(curl -sS -c "$COOKIEJAR" -H "content-type: application/x-sah-ws-4-call+json" -H "Authorization: X-Sah-Login" -H "accept: */*" -H "accept-language: en-US,en;q=0.7" -H "sec-gpc: 1" -H "Origin: http://${HOST}" -H "Referer: http://${HOST}/" -X POST "$AUTH_URL" -d "$AUTH_BODY")
  if (( VERBOSE )); then
    echo "AUTH RESP:"
    pretty_print_json() {
      local data="$1"
      if command -v jq >/dev/null 2>&1; then
        echo "$data" | jq .
      elif command -v python3 >/dev/null 2>&1; then
        echo "$data" | python3 -m json.tool
      elif command -v python >/dev/null 2>&1; then
        echo "$data" | python -m json.tool
      else
        # naive fallback: add some newlines for readability
        echo "$data" | sed -e 's/{/\n{\n/g' -e 's/}/\n}\n/g' -e 's/,/,\n/g'
      fi
    }
    pretty_print_json "$AUTH_RESP"
  fi
  # extract contextID from response JSON using jq if available, else fallback
  if command -v jq >/dev/null 2>&1; then
    CTX=$(echo "$AUTH_RESP" | jq -r '.data.contextID // empty')
  else
    # naive extraction
    CTX=$(echo "$AUTH_RESP" | sed -n 's/.*"contextID"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
  fi
  if [[ -n "$CTX" ]]; then
    TOKEN="$CTX"
    HEADER_ARGS+=( -H "Authorization: X-Sah ${TOKEN}" -H "x-context: ${TOKEN}" )
    # reuse cookies via curl --cookie
    COOKIE_ARG=(--cookie "$COOKIEJAR")
  else
    echo "WARNING: authentication did not return a contextID" >&2
  fi
fi

if (( VERBOSE )); then
  echo "POST $URL"
  echo "BODY: $BODY"
  echo "HEADERS: ${HEADER_ARGS[@]}"
  if [[ -n "${COOKIE_ARG[*]:-}" ]]; then
    echo "Using cookie jar: $COOKIEJAR"
  fi
fi

if [[ -n "${COOKIE_ARG[*]:-}" ]]; then
  RESP=$(curl -sS ${COOKIE_ARG[@]} ${HEADER_ARGS[@]} -X POST "$URL" -d "$BODY")
else
  RESP=$(curl -sS ${HEADER_ARGS[@]} -X POST "$URL" -d "$BODY")
fi

# Print nicely formatted JSON when verbose; otherwise print raw response
if (( VERBOSE )); then
  echo "RESPONSE:"
  pretty_print_json "$RESP"
else
  echo "$RESP"
fi

# cleanup tmp cookie jar file
if [[ -f "$COOKIEJAR" ]]; then
  rm -f "$COOKIEJAR"
fi
