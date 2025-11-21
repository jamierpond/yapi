#!/usr/bin/env bats

setup() {
  export TEST_HOME="$(mktemp -d)"
  export HOME="$TEST_HOME"
  export YAPI_HOME="$TEST_HOME/.config/yapi"
  export SHELL="/bin/bash"

  SCRIPT_DIR="$(cd "$(dirname "$BATS_TEST_DIRNAME")" && pwd)"
  export QUICKSTART="$SCRIPT_DIR/quickstart.sh"
}

teardown() {
  rm -rf "$TEST_HOME"
}

@test "quickstart creates YAPI_HOME directory" {
  run bash "$QUICKSTART"
  [ "$status" -eq 0 ]
  [ -d "$YAPI_HOME" ]
}

@test "quickstart clones yapi executable" {
  run bash "$QUICKSTART"
  [ "$status" -eq 0 ]
  [ -x "$YAPI_HOME/yapi" ]
}

@test "quickstart clones lib files" {
  run bash "$QUICKSTART"
  [ "$status" -eq 0 ]
  [ -f "$YAPI_HOME/lib/yapi_utils.sh" ]
  [ -f "$YAPI_HOME/lib/yapi_config.sh" ]
  [ -f "$YAPI_HOME/lib/yapi_http.sh" ]
}

@test "quickstart adds config to shellrc" {
  run bash "$QUICKSTART"
  [ "$status" -eq 0 ]
  grep -q "YAPI_HOME" "$TEST_HOME/.bashrc"
  grep -q 'alias a="yapi"' "$TEST_HOME/.bashrc"
}

@test "quickstart is idempotent" {
  bash "$QUICKSTART"
  bash "$QUICKSTART"
  count=$(grep -c "YAPI_HOME" "$TEST_HOME/.bashrc")
  [ "$count" -eq 1 ]
}

@test "installed yapi --help works" {
  bash "$QUICKSTART"
  run "$YAPI_HOME/yapi" --help
  [ "$status" -eq 0 ]
  [[ "$output" == *"YAML API Testing Tool"* ]]
}

@test "installed yapi can run unit tests" {
  bash "$QUICKSTART"
  run bats "$YAPI_HOME/test/yapi_utils.bats" "$YAPI_HOME/test/yapi_config.bats"
  [ "$status" -eq 0 ]
}
