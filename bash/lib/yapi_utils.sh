#!/bin/bash
# yapi_utils.sh - General utilities for yapi

# Display help message
show_help() {
  cat << EOF
yapi - YAML API Testing Tool (HTTP/REST and gRPC)

Usage: yapi [OPTIONS]

Options:
  -c, --config FILE    Path to YAML config file (required)
  -u, --url URL        Override base URL from config file
  -a, --all            Search all YAML files (default: git-tracked only)
  -h, --help           Display this help message

Supported Protocols:
  HTTP/REST:  http://, https://
  gRPC:       grpc:// (plaintext), grpcs:// (TLS)
  TCP:        tcp://host:port

Examples:
  yapi -c test.yaml
  yapi --config test.yaml --url http://localhost:8080
  yapi -c grpc-service.yaml
  yapi --all

EOF
  exit 0
}

# Error handler
error_exit() {
  echo "Error: $1" >&2
  echo "Use -h or --help for usage information" >&2
  exit 1
}

# Check if a command exists
check_dependency() {
  local cmd="$1"
  local install_hint="$2"
  if ! command -v "$cmd" &>/dev/null; then
    if [[ -n "$install_hint" ]]; then
      error_exit "$cmd is required but not found. $install_hint"
    else
      error_exit "$cmd is required but not found"
    fi
  fi
}

# Log command to history file
log_history() {
  local config_path="$1"
  local cli_url="$2"
  local history_file="${HOME}/.yapi_history"
  local realpath_config
  realpath_config=$(realpath "$config_path")

  local command="yapi -c \"$realpath_config\""
  if [[ -n "$cli_url" ]]; then
    command+=" -u \"$cli_url\""
  fi

  echo "$(date +%s) | $command" >> "$history_file"
}

# Format and print response (JSON if possible)
print_response() {
  local response="$1"
  if echo "$response" | jq . &>/dev/null; then
    echo "$response" | jq
  else
    echo "$response"
  fi
}

# Print timing info
print_timing() {
  local elapsed_ms="$1"
  echo "Request completed in ${elapsed_ms}ms" >&2
}
