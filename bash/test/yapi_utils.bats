#!/usr/bin/env bats

setup() {
  SCRIPT_DIR="$(cd "$(dirname "$BATS_TEST_DIRNAME")" && pwd)"
  source "$SCRIPT_DIR/lib/yapi_utils.sh"
}

@test "error_exit prints error message to stderr" {
  run error_exit "test error"
  [ "$status" -eq 1 ]
  [[ "$output" == *"Error: test error"* ]]
}

@test "check_dependency passes for existing command" {
  run check_dependency "bash"
  [ "$status" -eq 0 ]
}

@test "check_dependency fails for missing command" {
  run check_dependency "nonexistent_command_xyz"
  [ "$status" -eq 1 ]
  [[ "$output" == *"nonexistent_command_xyz is required"* ]]
}

@test "print_response formats valid JSON" {
  run print_response '{"key": "value"}'
  [ "$status" -eq 0 ]
  [[ "$output" == *'"key": "value"'* ]]
}

@test "print_response outputs non-JSON as-is" {
  run print_response "plain text"
  [ "$status" -eq 0 ]
  [ "$output" = "plain text" ]
}
