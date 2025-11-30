#!/usr/bin/env bats

setup() {
  SCRIPT_DIR="$(cd "$(dirname "$BATS_TEST_DIRNAME")" && pwd)"
  source "$SCRIPT_DIR/lib/yapi_utils.sh"
  source "$SCRIPT_DIR/lib/yapi_config.sh"
  source "$SCRIPT_DIR/lib/yapi_http.sh"

  TEST_DIR="$(mktemp -d)"
  cat > "$TEST_DIR/with_query.yaml" << 'EOF'
url: https://example.com
method: GET
path: /api/test
query:
  foo: bar
  baz: qux
EOF
}

teardown() {
  rm -rf "$TEST_DIR"
}

@test "http_build_url builds basic URL" {
  result=$(http_build_url "https://example.com" "/api/test" "$TEST_DIR/with_query.yaml" "false")
  [ "$result" = "https://example.com/api/test" ]
}

@test "http_build_url handles URL without path" {
  result=$(http_build_url "https://example.com" "" "$TEST_DIR/with_query.yaml" "false")
  [ "$result" = "https://example.com" ]
}

@test "http_build_url adds query string" {
  result=$(http_build_url "https://example.com" "/api" "$TEST_DIR/with_query.yaml" "true")
  [[ "$result" == *"foo=bar"* ]]
  [[ "$result" == *"baz=qux"* ]]
}

@test "http_build_url encodes special characters in path" {
  result=$(http_build_url "https://example.com" "/api/test with spaces" "$TEST_DIR/with_query.yaml" "false")
  [[ "$result" == *"%20"* ]]
}
