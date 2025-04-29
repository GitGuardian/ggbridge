#!/bin/bash

OK="✅"
KO="❌"

GGBRIDGE_CLIENT_HOST=${GGBRIDGE_CLIENT_HOST:-ggbridge-client}
GGBRIDGE_SERVER_HOST=${GGBRIDGE_SERVER_HOST:-ggbridge-server}
GGBRIDGE_TUNNEL_HEALTH_PORT=${GGBRIDGE_TUNNEL_HEALTH_PORT:-9081}
GGBRIDGE_TUNNEL_SOCKS_PORT=${GGBRIDGE_TUNNEL_SOCKS_PORT:-9180}

TABLE_TEST_COL_WIDTH=35
TABLE_RESULT_COL_WIDTH=6

# -- Functions
# Print a formatted result row
print_result() {
  local name="$1"
  local status="$2"
  printf "| %-*s | %-*s |\n" "$TABLE_TEST_COL_WIDTH" "$name" "$TABLE_RESULT_COL_WIDTH" "$status"
}

test_http() {
  local name="$1"
  local url="$2"

  if curl -s --max-time 5 "$url" >/dev/null; then
    print_result "$name" "$OK"
  else
    print_result "$name" "$KO"
  fi
}

test_socks_proxy() {
  local name="$1"
  local proxy="$2"
  local url="$3"

  if curl -s --max-time 5 --proxy "$proxy" "$url" >/dev/null; then
    print_result "$name" "$OK"
  else
    print_result "$name" "$KO"
  fi
}

# Header
print_result "Test" "Result"
printf "|-%s-|-%s-|\n" "$(printf '%.0s-' $(seq 1 $TABLE_TEST_COL_WIDTH))" "$(printf '%.0s-' $(seq 1 $TABLE_RESULT_COL_WIDTH))"

# Tests
test_http "Client → Server health tunnel" "http://${GGBRIDGE_CLIENT_HOST}:${GGBRIDGE_TUNNEL_HEALTH_PORT}/healthz"
test_http "Server → Client health tunnel" "http://${GGBRIDGE_SERVER_HOST}:${GGBRIDGE_TUNNEL_HEALTH_PORT}/healthz"
test_socks_proxy "Server → Client proxy tunnel" "socks5h://${GGBRIDGE_SERVER_HOST}:${GGBRIDGE_TUNNEL_SOCKS_PORT}" "https://httpstat.us/200"
