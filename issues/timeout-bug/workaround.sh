#!/bin/bash

# Workaround for yapi timeout bug
# Use this curl script until the timeout field is fixed

BASE_URL="${1:-http://localhost:3002}"

echo "Testing /process/video (curl workaround)"
echo "========================================="

curl -X POST "$BASE_URL/process/video" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://pond.audio/rick.mp4", "outputs": ["poster", "placeholder"]}' \
  --max-time 600 \
  -w "\n\nTime: %{time_total}s | Status: %{http_code}\n" \
  -o /tmp/repro-output.json \
  -s

echo ""
echo "Assertions:"
jq -e '.poster != null' /tmp/repro-output.json > /dev/null && echo "[PASS] .poster != null" || echo "[FAIL] .poster != null"
jq -e '.placeholder != null' /tmp/repro-output.json > /dev/null && echo "[PASS] .placeholder != null" || echo "[FAIL] .placeholder != null"
