# Bug: Request timeout not respected

## Summary

The `timeout` field in yapi request files is not being respected. Requests timeout much earlier than the specified duration.

## Steps to Reproduce

1. Create a yapi file with a long timeout (e.g., `timeout: 5m`)
2. Make a request to an endpoint that takes longer than the default timeout but less than the specified timeout
3. Observe that the request fails with `context deadline exceeded` before 5 minutes

## Expected Behavior

The request should wait for the full duration specified in the `timeout` field (5 minutes in this case) before timing out.

## Actual Behavior

The request times out much sooner than 5 minutes, likely using a hardcoded default timeout instead of the user-specified value.

## Error Message

```
failed to execute request: Post "http://localhost:3002/process/video": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
```

## Environment

- yapi version: (current)
- OS: macOS

## Reproducer File

See `timeout-bug-repro.yapi.yml` in this directory.

## Notes

- The endpoint works correctly when tested with curl using `--max-time 300`
- The timeout duration string format (`5m`) appears to match the documented format
- This suggests the timeout value is either not being parsed correctly or not being applied to the HTTP client
