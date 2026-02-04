#!/usr/bin/env bash
# sendexample1.sh - HTTP examples using yapi send
# These mirror the equivalent .yapi.yml config files in examples/http/

set -euo pipefail

echo "=== Simple GET (mirrors http/simple-get.yapi.yml) ==="
yapi send https://jsonplaceholder.typicode.com/posts/1

echo ""
echo "=== POST JSON (mirrors http/post-json.yapi.yml) ==="
yapi send https://httpbin.org/post '{"title":"Hello from yapi"}' --jq '.json'

echo ""
echo "=== GET with custom headers (mirrors http/custom-headers.yapi.yml) ==="
yapi send https://httpbin.org/headers \
  -H 'X-Custom-Header: my-custom-value' \
  -H 'Accept: application/json'

echo ""
echo "=== Explicit method with -X ==="
yapi send -X PUT https://httpbin.org/put '{"updated":true}'

echo ""
echo "=== GET with jq filter ==="
yapi send https://jsonplaceholder.typicode.com/posts/1 --jq '.title'
