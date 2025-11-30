#!/bin/bash
# yapi_grpc.sh - gRPC request handling

# Execute gRPC request
execute_grpc_request() {
  local config="$1"
  local url="$2"

  # Check dependency
  check_dependency "grpcurl" "Install it: brew install grpcurl (macOS) or go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"

  # Validate required fields
  if [[ -z "$GRPC_SERVICE" ]]; then
    error_exit "service field is required for gRPC requests"
  fi
  if [[ -z "$GRPC_RPC" ]]; then
    error_exit "rpc field is required for gRPC requests"
  fi

  # Handle proto files (optional - uses server reflection if not provided)
  local use_proto_files=false
  local resolved_proto_path=""

  if [[ -n "$GRPC_PROTO" ]]; then
    use_proto_files=true

    if [[ -z "$GRPC_PROTO_PATH" ]]; then
      error_exit "proto_path is required when proto is specified"
    fi

    # Resolve proto_path relative to config file directory
    local config_dir
    config_dir=$(dirname "$config")

    if [[ "$GRPC_PROTO_PATH" != /* ]]; then
      resolved_proto_path=$(cd "$config_dir" && cd "$GRPC_PROTO_PATH" && pwd)
    else
      resolved_proto_path="$GRPC_PROTO_PATH"
    fi

    if [[ ! -d "$resolved_proto_path" ]]; then
      error_exit "proto_path directory does not exist: $resolved_proto_path"
    fi
  fi

  # Extract server address (remove scheme)
  local server_addr
  server_addr="${url#grpc://}"
  server_addr="${server_addr#grpcs://}"

  # Build grpcurl arguments
  local grpcurl_args=()

  if [[ "$use_proto_files" == "true" ]]; then
    grpcurl_args+=(
      -import-path "$resolved_proto_path"
      -proto "$GRPC_PROTO"
    )
  fi

  # Determine plaintext mode
  local use_plaintext=false
  if [[ "$url" =~ ^grpc:// ]]; then
    use_plaintext=true
  fi
  if [[ "$GRPC_PLAINTEXT" == "true" ]]; then
    use_plaintext=true
  fi

  if [[ "$use_plaintext" == "true" ]]; then
    grpcurl_args+=(-plaintext)
  fi

  if [[ "$GRPC_INSECURE" == "true" ]]; then
    grpcurl_args+=(-insecure)
  fi

  # Add metadata headers (using safe extraction to prevent yq injection)
  if [[ "$GRPC_METADATA_EXISTS" == "true" ]]; then
    # Safe extraction: get all key-value pairs at once without interpolating user input into yq expressions
    while IFS=$'\t' read -r key value; do
      if [[ -n "$key" ]]; then
        grpcurl_args+=(-H "$key: $value")
      fi
    done < <(yq e '.metadata | to_entries | .[] | [.key, .value] | @tsv' "$config")
  fi

  # Add request body
  if [[ "$CONFIG_BODY_EXISTS" == "true" ]]; then
    local request_json
    request_json=$(yq e '.body' -o=json "$config")
    grpcurl_args+=(-d "$request_json")
  fi

  grpcurl_args+=("$server_addr")
  grpcurl_args+=("$GRPC_SERVICE/$GRPC_RPC")

  echo "Executing gRPC request to $server_addr ($GRPC_SERVICE/$GRPC_RPC)" >&2

  local start_time end_time elapsed_ms response
  start_time=$(date +%s%N)
  response=$(grpcurl "${grpcurl_args[@]}")
  end_time=$(date +%s%N)
  elapsed_ms=$(( (end_time - start_time) / 1000000 ))

  print_response "$response"
  print_timing "$elapsed_ms"
}
