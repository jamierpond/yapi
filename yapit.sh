#!/bin/bash
# yapit Yaml API Testing
# requires: bash, curl, yq, fzf (optional, for interactive file selection)
set -e

# Default values
config=""
cli_url=""

# Display help message
show_help() {
  cat << EOF
yapit - YAML API Testing Tool

Usage: $(basename "$0") [OPTIONS]

Options:
  -c, --config FILE    Path to YAML config file (required)
  -u, --url URL        Override base URL from config file
  -h, --help           Display this help message

Examples:
  $(basename "$0") -c test.yaml
  $(basename "$0") --config test.yaml --url http://localhost:8080

EOF
  exit 0
}

# Error handler
error_exit() {
  echo "Error: $1" >&2
  echo "Use -h or --help for usage information" >&2
  exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -c|--config)
      if [[ -z "$2" ]] || [[ "$2" == -* ]]; then
        error_exit "Option $1 requires an argument"
      fi
      config="$2"
      shift 2
      ;;
    -u|--url)
      if [[ -z "$2" ]] || [[ "$2" == -* ]]; then
        error_exit "Option $1 requires an argument"
      fi
      cli_url="$2"
      shift 2
      ;;
    -h|--help)
      show_help
      ;;
    -*)
      error_exit "Unknown option: $1"
      ;;
    *)
      error_exit "Unexpected argument: $1"
      ;;
  esac
done

# Handle config file selection
if [[ -z "$config" ]]; then
  # Check if fzf is available
  if ! command -v fzf &>/dev/null; then
    error_exit "Config file is required (use -c or --config). Install fzf for interactive selection."
  fi

  # Find YAML files and let user select with fzf
  config=$(find . -type f \( -name "*.yml" -o -name "*.yaml" \) 2>/dev/null | fzf --prompt="Select config file: " --height=40% --border)

  # Exit if user cancelled fzf
  if [[ -z "$config" ]]; then
    error_exit "No config file selected"
  fi

  echo "Selected: $config"
fi

# Validate config file exists and is valid YAML
if [[ ! -f "$config" ]]; then
  error_exit "Config file '$config' does not exist"
fi

if ! yq e 'true' "$config" &>/dev/null; then
  error_exit "Config file '$config' is not a valid YAML file"
fi

# Extract values from config
config_url=$(yq e '.url // ""' "$config")
path=$(yq e '.path // ""' "$config")
method=$(yq e '.method // "GET"' "$config")
content_type=$(yq e '.content_type // ""' "$config")
body_exists=$(yq e 'has("body")' "$config")

# URL priority: CLI flag > YAML url (required if no CLI flag)
if [[ -n "$cli_url" ]]; then
  url="$cli_url"
elif [[ -n "$config_url" ]]; then
  url="$config_url"
else
  error_exit "URL is required: either provide 'url' in config file or use -u flag"
fi

# Build full URL
full_url="${url%/}${path}"
# echo "Requesting: $full_url"

# Validate method
if [[ -z "$method" ]]; then
  error_exit "HTTP method is required in config file"
fi

# Build curl command
curl_args=(
  -X "$method"
  "$full_url"
  -s
)

# Handle body if present
if [[ "$body_exists" == "true" ]]; then
  # Require content_type when body is present
  if [[ -z "$content_type" ]]; then
    error_exit "content_type is required when body is present"
  fi

  # Currently only support JSON
  if [[ "$content_type" != "application/json" ]]; then
    error_exit "Only 'application/json' content_type is currently supported"
  fi

  # Convert YAML body to JSON
  body_json=$(yq e '.body' -o=json "$config")

  # echo "Request Body: $body_json"

  curl_args+=(
    -H "Content-Type: $content_type"
    -d "$body_json"
  )
fi

# Execute request and capture output
#echo "Executing $method request to $full_url"
#echo "Curl command: curl ${curl_args[*]}"
response=$(curl "${curl_args[@]}")

# Try to format as JSON if possible, otherwise print as-is
if echo "$response" | jq . &>/dev/null; then
  echo "$response" | jq
else
  echo "$response"
fi



