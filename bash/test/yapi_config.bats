#!/usr/bin/env bats

setup() {
  SCRIPT_DIR="$(cd "$(dirname "$BATS_TEST_DIRNAME")" && pwd)"
  source "$SCRIPT_DIR/lib/yapi_utils.sh"
  source "$SCRIPT_DIR/lib/yapi_config.sh"

  # Create temp test files
  TEST_DIR="$(mktemp -d)"
  cat > "$TEST_DIR/valid.yaml" << 'EOF'
url: https://example.com
method: GET
path: /api/test
EOF

  cat > "$TEST_DIR/grpc.yaml" << 'EOF'
url: grpc://localhost:50051
method: grpc
service: test.Service
rpc: GetData
EOF

  cat > "$TEST_DIR/tcp.yaml" << 'EOF'
url: tcp://localhost:9000
data: hello
EOF

  echo "not: valid: yaml: {{" > "$TEST_DIR/invalid.yaml"
}

teardown() {
  rm -rf "$TEST_DIR"
}

@test "validate_config_file passes for valid YAML" {
  run validate_config_file "$TEST_DIR/valid.yaml"
  [ "$status" -eq 0 ]
}

@test "validate_config_file fails for missing file" {
  run validate_config_file "$TEST_DIR/nonexistent.yaml"
  [ "$status" -eq 1 ]
  [[ "$output" == *"does not exist"* ]]
}

@test "validate_config_file fails for invalid YAML" {
  run validate_config_file "$TEST_DIR/invalid.yaml"
  [ "$status" -eq 1 ]
  [[ "$output" == *"not a valid YAML"* ]]
}

@test "get_config_value returns value from config" {
  result=$(get_config_value "$TEST_DIR/valid.yaml" ".url")
  [ "$result" = "https://example.com" ]
}

@test "get_config_value returns default for missing key" {
  result=$(get_config_value "$TEST_DIR/valid.yaml" ".missing" "default_value")
  [ "$result" = "default_value" ]
}

@test "detect_protocol returns http for https URL" {
  result=$(detect_protocol "https://example.com" "GET")
  [ "$result" = "http" ]
}

@test "detect_protocol returns grpc for grpc URL" {
  result=$(detect_protocol "grpc://localhost:50051" "GET")
  [ "$result" = "grpc" ]
}

@test "detect_protocol returns grpc for grpcs URL" {
  result=$(detect_protocol "grpcs://localhost:50051" "GET")
  [ "$result" = "grpc" ]
}

@test "detect_protocol returns tcp for tcp URL" {
  result=$(detect_protocol "tcp://localhost:9000" "GET")
  [ "$result" = "tcp" ]
}

@test "config_has_key returns true for existing key" {
  result=$(config_has_key "$TEST_DIR/valid.yaml" "url")
  [ "$result" = "true" ]
}

@test "config_has_key returns false for missing key" {
  result=$(config_has_key "$TEST_DIR/valid.yaml" "nonexistent")
  [ "$result" = "false" ]
}
