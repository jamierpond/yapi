# Plan: `wait_for` Feature - DX Design

## Overview

A new `wait_for` block that repeatedly polls an endpoint until a condition is satisfied. Designed for async server operations that require time to complete (job processing, webhooks, eventual consistency, etc.).

## DX Design

### Basic Syntax

```yaml
yapi: v1
url: ${url}/jobs/${job_id}
method: GET

wait_for:
  until:
    - .status == "completed"
  period: 2s
  timeout: 60s
```

### Fixed Period (Simple)

```yaml
wait_for:
  until:
    - .status == "completed"
  period: 2s          # Fixed time between attempts
  timeout: 60s        # Total time limit
```

### Exponential Backoff

```yaml
wait_for:
  until:
    - .status == "completed"
  backoff:
    seed: 1s          # Initial wait time
    multiplier: 2     # Each attempt waits multiplier * previous
  timeout: 60s        # Total time limit
```

Backoff example with `seed: 1s, multiplier: 2`:
- Attempt 1 → wait 1s
- Attempt 2 → wait 2s
- Attempt 3 → wait 4s
- Attempt 4 → wait 8s
- ...continues until timeout

### Behavior

1. Execute the request
2. If `until` conditions pass → success, stop polling
3. If `until` conditions fail OR request errors (5xx, network) → wait (period or backoff), retry
4. If `timeout` exceeded → fail with timeout error

**Error handling**: Intermediate failures (5xx, network errors, 4xx) are treated as "not ready yet" and polling continues. Only timeout causes failure.

**Timing**: Either `period` OR `backoff` must be specified (mutually exclusive).

---

## Use Cases

### 1. Single Request - Job Completion

```yaml
yapi: v1
url: ${url}/jobs/${job_id}
method: GET

wait_for:
  until:
    - .status == "completed" or .status == "failed"
  period: 2s
  timeout: 120s

expect:
  status: 200
  assert:
    - .status == "completed"  # Final assertion after wait_for succeeds
```

### 2. Chain Step - Async Workflow

```yaml
yapi: v1
chain:
  - name: create_job
    url: ${url}/jobs
    method: POST
    body:
      type: "data_export"
    expect:
      status: 202
      assert:
        - .job_id != null

  - name: wait_for_job
    url: ${url}/jobs/${create_job.job_id}
    method: GET
    wait_for:
      until:
        - .status == "completed"
      backoff:
        seed: 1s
        multiplier: 2
      timeout: 300s
    expect:
      status: 200
      assert:
        - .download_url != null

  - name: download_result
    url: ${wait_for_job.download_url}
    method: GET
    output_file: ./export.csv
```

### 3. Webhook/Callback Waiting

```yaml
yapi: v1
chain:
  - name: trigger_webhook
    url: ${url}/webhooks/trigger
    method: POST

  - name: check_received
    url: ${url}/webhooks/received
    method: GET
    wait_for:
      until:
        - . | length > 0
        - .[0].payload.event == "user.created"
      period: 1s
      timeout: 30s
```

### 4. Database Eventual Consistency

```yaml
yapi: v1
chain:
  - name: create_user
    url: ${url}/users
    method: POST
    body:
      email: "test@example.com"
    expect:
      status: 201

  - name: verify_searchable
    url: ${url}/users/search?email=test@example.com
    method: GET
    wait_for:
      until:
        - . | length == 1
      period: 500ms
      timeout: 10s
```

---

## Interaction with Existing Features

### With `expect`

`wait_for` runs first. Once `until` conditions pass, `expect` runs on the final response:

```yaml
wait_for:
  until:
    - .status != "pending"  # Wait until not pending
  period: 1s
  timeout: 30s

expect:
  status: 200
  assert:
    - .status == "completed"  # Then verify it's completed (not failed)
```

### With `timeout`

The existing `timeout` field is per-request. `wait_for.timeout` is total polling time:

```yaml
timeout: 5s  # Each poll attempt times out after 5s

wait_for:
  until:
    - .ready == true
  period: 2s
  timeout: 60s  # Total polling time limit
```

### With `delay`

`delay` happens before `wait_for` starts:

```yaml
delay: 5s  # Wait 5s before starting to poll

wait_for:
  until:
    - .status == "done"
  period: 2s
  timeout: 30s
```

---

## Output During Polling

When running with verbose/default output:

```
[POLL] Attempt 1 - conditions not met, retrying in 2s...
[POLL] Attempt 2 - conditions not met, retrying in 2s...
[POLL] Attempt 3 - request failed (503), retrying in 2s...
[POLL] Attempt 4 - conditions met!
```

---

## Config Schema

```go
type Backoff struct {
    Seed       string  `yaml:"seed"`       // Initial wait, e.g., "1s"
    Multiplier float64 `yaml:"multiplier"` // e.g., 2
}

type WaitFor struct {
    Until   []string `yaml:"until"`             // Required: JQ assertions
    Period  string   `yaml:"period,omitempty"`  // Fixed interval, e.g., "2s"
    Backoff *Backoff `yaml:"backoff,omitempty"` // Exponential backoff
    Timeout string   `yaml:"timeout"`           // Required: total time limit
}
```

Added to `ConfigV1`:
```go
type ConfigV1 struct {
    // ... existing fields ...
    WaitFor *WaitFor `yaml:"wait_for,omitempty"`
}
```

---

## Validation Rules

1. `until` is required and must have at least one assertion
2. `timeout` is required and must be valid Go duration
3. Exactly one of `period` OR `backoff` must be specified (mutually exclusive)
4. If `period`: must be valid Go duration
5. If `backoff`: `seed` must be valid Go duration, `multiplier` must be > 1
6. All `until` expressions must be valid JQ
