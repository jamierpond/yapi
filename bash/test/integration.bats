#!/usr/bin/env bats

setup() {
  SCRIPT_DIR="$(cd "$(dirname "$BATS_TEST_DIRNAME")" && pwd)"
  YAPI="$SCRIPT_DIR/yapi"
}

@test "yapi --help shows usage" {
  run "$YAPI" --help
  [ "$status" -eq 0 ]
  [[ "$output" == *"YAML API Testing Tool"* ]]
  [[ "$output" == *"Usage:"* ]]
}

@test "yapi -h shows usage" {
  run "$YAPI" -h
  [ "$status" -eq 0 ]
  [[ "$output" == *"Usage:"* ]]
}

@test "yapi fails with unknown option" {
  run "$YAPI" --unknown-option
  [ "$status" -eq 1 ]
  [[ "$output" == *"Unknown option"* ]]
}

@test "yapi fails with missing config file" {
  run "$YAPI" -c nonexistent.yaml
  [ "$status" -eq 1 ]
  [[ "$output" == *"does not exist"* ]]
}

@test "yapi executes HTTP GET request" {
  run "$YAPI" -c "$SCRIPT_DIR/examples/google.yapi.yml"
  [ "$status" -eq 0 ]
  [[ "$output" == *"Google"* ]]
}

@test "yapi executes HTTP POST request" {
  run "$YAPI" -c "$SCRIPT_DIR/examples/create-post.yapi.yml"
  [ "$status" -eq 0 ]
  [[ "$output" == *"httpbin.org"* ]]
}

@test "yapi URL override works" {
  run "$YAPI" -c "$SCRIPT_DIR/examples/google.yapi.yml" -u "https://httpbin.org/get"
  [ "$status" -eq 0 ]
  [[ "$output" == *"httpbin.org"* ]]
}

@test "yapi executes jq post-processing" {
  run "$YAPI" -c "$SCRIPT_DIR/test/jq_filter.yapi.yml"
  [ "$status" -eq 0 ]
  [[ "$output" == *"samples"* ]]
}

