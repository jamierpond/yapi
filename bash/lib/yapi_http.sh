#!/bin/bash
# yapi_http.sh - HTTP request handling

# Build URL with encoded path and query string
http_build_url() {
  local base_url="$1"
  local path="$2"
  local config="$3"
  local query_exists="$4"
  local full_url

  if [[ -n "$path" ]]; then
    # Encode path segments but preserve slashes
    local protected
    protected=$(echo "$path" | sed 's/%\([0-9A-Fa-f][0-9A-Fa-f]\)/___PERCENT___\1/g')
    protected=$(echo "$protected" | sed 's/\//___SLASH___/g')
    local encoded
    encoded=$(printf "%s" "$protected" | jq -Rr @uri)
    local encoded_path
    encoded_path=$(echo "$encoded" | sed 's/___SLASH___/\//g' | sed 's/___PERCENT___/%/g')
    full_url="${base_url%/}${encoded_path}"
  else
    full_url="$base_url"
  fi

  # Build query string if present (using safe extraction to prevent yq injection)
  if [[ "$query_exists" == "true" ]]; then
    local query_string=""
    local first=true

    # Safe extraction: get all key-value pairs at once without interpolating user input into yq expressions
    while IFS=$'\t' read -r key value; do
      if [[ -n "$key" ]]; then
        local encoded_key
        encoded_key=$(printf "%s" "$key" | jq -Rr @uri)
        local encoded_value
        encoded_value=$(printf "%s" "$value" | jq -Rr @uri)

        if [[ "$first" == "true" ]]; then
          query_string="?${encoded_key}=${encoded_value}"
          first=false
        else
          query_string="${query_string}&${encoded_key}=${encoded_value}"
        fi
      fi
    done < <(yq e '.query | to_entries | .[] | [.key, .value] | @tsv' "$config")

    full_url="${full_url}${query_string}"
  fi

  echo "$full_url"
}

# Build curl arguments array
http_build_curl_args() {
  local method="$1"
  local full_url="$2"
  local config="$3"
  local content_type="$4"
  local body_exists="$5"
  local json_exists="$6"

  # Start with basic args
  CURL_ARGS=(
    -X "$method"
    "$full_url"
    -s
  )

  # Validate body and json are mutually exclusive
  if [[ "$body_exists" == "true" ]] && [[ "$json_exists" == "true" ]]; then
    error_exit "Cannot specify both 'body' and 'json' fields - use only one"
  fi

  # Handle request body
  if [[ "$body_exists" == "true" ]] || [[ "$json_exists" == "true" ]]; then
    if [[ -z "$content_type" ]]; then
      error_exit "content_type is required when body or json is present"
    fi

    if [[ "$content_type" != "application/json" ]]; then
      error_exit "Only 'application/json' content_type is currently supported"
    fi

    local request_json
    if [[ "$body_exists" == "true" ]]; then
      request_json=$(yq e '.body' -o=json "$config")
    else
      request_json=$(yq e '.json' "$config")
    fi

    CURL_ARGS+=(
      -H "Content-Type: $content_type"
      -d "$request_json"
    )
  fi
}

# Execute HTTP request
execute_http_request() {
  local config="$1"
  local url="$2"

  # Build full URL
  local full_url
  full_url=$(http_build_url "$url" "$CONFIG_PATH" "$config" "$CONFIG_QUERY_EXISTS")

  # Validate method
  if [[ -z "$CONFIG_METHOD" ]]; then
    error_exit "HTTP method is required in config file"
  fi

  # Build curl args
  http_build_curl_args "$CONFIG_METHOD" "$full_url" "$config" "$CONFIG_CONTENT_TYPE" "$CONFIG_BODY_EXISTS" "$CONFIG_JSON_EXISTS"

  echo "Executing $CONFIG_METHOD request to $full_url" >&2

  local start_time end_time elapsed_ms response
  start_time=$(date +%s%N)
  response=$(curl -L "${CURL_ARGS[@]}")
  end_time=$(date +%s%N)
  elapsed_ms=$(( (end_time - start_time) / 1000000 ))

  if [[ -n "$CONFIG_JQ_FILTER" ]]; then
    response=$(echo "$response" | jq "$CONFIG_JQ_FILTER")
  fi

  print_response "$response"
  print_timing "$elapsed_ms"
}
