#!/usr/bin/env bash
# sendexample2.sh - TCP and advanced examples using yapi send
# These mirror the equivalent .yapi.yml config files in examples/tcp/

set -euo pipefail

echo "=== TCP echo (mirrors tcp/echo-server.yapi.yml) ==="
yapi send 'tcp://tcpbin.com:4242' 'Hello from yapi!'

echo ""
echo "=== JSON output mode ==="
yapi send https://jsonplaceholder.typicode.com/posts/1 --json

echo ""
echo "=== Verbose mode ==="
yapi send -v https://httpbin.org/get

echo ""
echo "=== POST with explicit content type ==="
yapi send -X POST https://httpbin.org/post '{"id":42}' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer test-token'
