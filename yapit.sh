#!/bin/bash
# yapit Yaml API Testing
# requires: bash, curl, yq
set -e

config="$1"
url="$2"

default_url="http://localhost:3000"
usage_string="Usage: $0 <config> <url=$default_url>"
if [ -z "$config" ]; then
  echo "$usage_string"
  exit 1
fi

if [ -z "$url" ]; then
  url="$default_url"
fi

config_exists=$(yq e 'true' $config 2>/dev/null || echo "false")
if [ "$config_exists" != "true" ]; then
  echo "Config file $config does not exist or is not a valid YAML file."
  exit 1
fi

endpoint=$(yq e '.endpoint' $config)
json=$(yq e '.json' $config)
method=$(yq e '.method' $config)

url="${url%/}$endpoint"

curl                                                                           \
  -X "$method"                                                                 \
  "$url"                                                                       \
  -H "Content-Type: application/json"                                          \
  -d "$json"                                                                   \
  -s | jq



