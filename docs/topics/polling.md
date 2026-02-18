# Polling

The `wait_for` block lets you poll an endpoint until conditions are met.
Useful for async operations, health checks, and eventual consistency.

## Basic Polling

```yaml
yapi: v1
url: https://api.example.com/jobs/123
method: GET

wait_for:
  until:
    - .status == "completed"
  period: 2s
  timeout: 30s
```

This sends GET requests every 2 seconds until `.status == "completed"` or 30 seconds elapse.

## Fields

- **`until`**: Array of JQ assertions. ALL must pass to stop polling.
- **`period`**: Fixed delay between attempts (e.g., `"1s"`, `"500ms"`).
- **`backoff`**: Exponential backoff (mutually exclusive with `period`).
- **`timeout`**: Maximum wait time. Default: `30s`.

## Fixed Period

Retry at a constant interval:

```yaml
wait_for:
  until:
    - .ready == true
  period: 1s
  timeout: 60s
```

## Exponential Backoff

Start with a short delay, grow exponentially:

```yaml
wait_for:
  until:
    - .status == "done"
  backoff:
    seed: 1s             # First wait: 1s
    multiplier: 2        # Then 2s, 4s, 8s, 16s...
  timeout: 60s
```

The delay doubles each attempt: seed, seed*multiplier, seed*multiplier^2, etc.
`multiplier` must be greater than 1.

## Multiple Conditions

All `until` conditions must pass simultaneously:

```yaml
wait_for:
  until:
    - .status == "completed"
    - .result != null
    - .errors | length == 0
  period: 2s
  timeout: 30s
```

## Polling in Chains

Each chain step can have its own `wait_for`:

```yaml
chain:
  - name: start_job
    url: /api/jobs
    method: POST
    body: { task: "process" }
    expect:
      status: 202
      assert:
        - .job_id != null

  - name: wait_done
    url: /api/jobs/${start_job.job_id}
    method: GET
    wait_for:
      until:
        - .status == "completed"
      period: 2s
      timeout: 60s
    expect:
      assert:
        - .result != null
```

## Behavior

- If the request itself fails (network error), polling continues until timeout
- If assertions don't pass, polling retries
- On timeout, the test fails with an error
- On success, execution continues normally (chain proceeds to next step)

## See Also

- `yapi docs assert` — Assertions used in until conditions
- `yapi docs chain` — Multi-step request chaining
