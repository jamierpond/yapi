# Bug: Request timeout field is ignored

## Summary

The `timeout` field in yapi request files is not being applied to the HTTP client. Requests always timeout after ~20 seconds regardless of the specified timeout value.

## Observed Behavior

```bash
$ time yapi run ./media-service/.yapi/process-video-mp4.yapi.yml
failed to read response body: context canceled
yapi run ./media-service/.yapi/process-video-mp4.yapi.yml  0.02s user 0.03s system 0% cpu 20.883 total
```

**Key observations:**
1. Timeout is set to `10m` in the yapi file
2. Request fails after only ~21 seconds
3. Error message is misleading: `failed to read response body: context canceled` (not a timeout error)

## Expected Behavior

- Request should wait for the full `10m` (or whatever timeout is specified) before timing out
- Error message should indicate a timeout occurred, not "context canceled"

## Reproducer

See `repro.yapi.yml` in this directory.

The equivalent curl command works correctly:
```bash
$ time curl -X POST http://localhost:3002/process/video \
  -H "Content-Type: application/json" \
  -d '{"url": "https://pond.audio/rick.mp4", "outputs": ["poster", "placeholder"]}' \
  --max-time 120 -s | jq '.durationSeconds'

212.092
curl ...  0.03s user 0.02s system 0% cpu 1.245 total
```

## Root Cause Hypothesis

The `timeout` field is either:
1. Not being parsed from the YAML file
2. Not being applied to the `http.Client.Timeout`
3. Being overridden by a hardcoded default (~20s)

## Impact

- Cannot test slow endpoints (video encoding, ML inference, etc.)
- Workaround requires falling back to curl scripts
- Misleading error message makes debugging difficult

## Environment

- yapi version: latest
- OS: macOS Darwin 24.6.0
- Go version: (check with `go version`)

## Suggested Fix

In the HTTP client setup code, ensure the timeout from the parsed request config is applied:

```go
// Pseudocode - actual implementation may vary
client := &http.Client{
    Timeout: parsedRequest.Timeout, // This might be missing or using a default
}
```

Also update the error handling to distinguish between:
- Actual context cancellation (user cancelled)
- Timeout exceeded (show "request timed out after X")
