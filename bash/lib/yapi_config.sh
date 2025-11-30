#!/bin/bash
# yapi_config.sh - Config file handling and validation

YAPI_EXTENSION="yapi"

# Validate config file exists and is valid YAML
validate_config_file() {
  local config="$1"

  if [[ ! -f "$config" ]]; then
    error_exit "Config file '$config' does not exist"
  fi

  if ! yq e 'true' "$config" &>/dev/null; then
    error_exit "Config file '$config' is not a valid YAML file"
  fi
}

# Get a value from config with optional default
get_config_value() {
  local config="$1"
  local key="$2"
  local default="$3"

  if [[ -n "$default" ]]; then
    yq e "$key // \"$default\"" "$config"
  else
    yq e "$key // \"\"" "$config"
  fi
}

# Check if a key exists in config
config_has_key() {
  local config="$1"
  local key="$2"
  yq e "has(\"$key\")" "$config"
}

# Detect protocol from URL and method
detect_protocol() {
  local url="$1"
  local method="$2"

  if [[ "$url" =~ ^grpcs?:// ]] || [[ "$method" == "grpc" ]]; then
    echo "grpc"
  elif [[ "$url" =~ ^tcp:// ]] || [[ "$method" == "tcp" ]]; then
    echo "tcp"
  else
    echo "http"
  fi
}

# Parse all common config values into global variables
parse_common_config() {
  local config="$1"

  CONFIG_URL=$(get_config_value "$config" ".url")
  CONFIG_PATH=$(get_config_value "$config" ".path")
  CONFIG_METHOD=$(get_config_value "$config" ".method" "GET")
  CONFIG_CONTENT_TYPE=$(get_config_value "$config" ".content_type")
  CONFIG_BODY_EXISTS=$(config_has_key "$config" "body")
  CONFIG_JSON_EXISTS=$(config_has_key "$config" "json")
  CONFIG_QUERY_EXISTS=$(config_has_key "$config" "query")
  CONFIG_JQ_FILTER=$(get_config_value "$config" ".jq_filter")
}

# Parse gRPC-specific config
parse_grpc_config() {
  local config="$1"

  GRPC_PROTO=$(get_config_value "$config" ".proto")
  GRPC_PROTO_PATH=$(get_config_value "$config" ".proto_path")
  GRPC_SERVICE=$(get_config_value "$config" ".service")
  GRPC_RPC=$(get_config_value "$config" ".rpc")
  GRPC_PLAINTEXT=$(get_config_value "$config" ".plaintext")
  GRPC_INSECURE=$(get_config_value "$config" ".insecure")
  GRPC_METADATA_EXISTS=$(config_has_key "$config" "metadata")
}

# Parse TCP-specific config
parse_tcp_config() {
  local config="$1"

  TCP_DATA=$(get_config_value "$config" ".data")
  TCP_ENCODING=$(get_config_value "$config" ".encoding" "text")
  TCP_READ_TIMEOUT=$(get_config_value "$config" ".read_timeout" "5")
  TCP_CLOSE_AFTER_SEND=$(get_config_value "$config" ".close_after_send" "true")
}
