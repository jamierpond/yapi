# gRPC Limitations in yapi

## Issue 1: `wait_for` Does Not Work with gRPC

### Problem

The `wait_for` feature does not poll when used with gRPC endpoints. It makes a single request, evaluates the `until` condition, and if false, immediately runs `expect` assertions (which fail).

### Expected Behavior

```yaml
- name: wait_for_embedding
  url: grpc://localhost:50051
  service: sample_search_service.SampleSearchService
  rpc: GetStatus
  plaintext: true
  body: {}
  wait_for:
    until:
      - .embeddingStatus.inProgress == false
    period: 5s
    timeout: 600s
```

Should poll every 5s until `inProgress == false` or timeout.

### Actual Behavior

Makes one request, sees `inProgress: true`, immediately fails with:
```
[FAIL] .embeddingStatus.inProgress == false
assertion failed: assertion failed
  Expected: .embeddingStatus.inProgress to equal false
  Actual:   .embeddingStatus.inProgress = true
```

No polling occurs.

### Impact

Cannot use yapi for async gRPC workflows that require polling (embeddings, job processing, etc).

---

## Issue 2: Status Code Handling in Expect Blocks

### Problem

Currently, `expect.status` assumes HTTP status codes (200, 201, etc). For gRPC, status codes are different:

- `0` = OK
- `1` = CANCELLED
- `2` = UNKNOWN
- `3` = INVALID_ARGUMENT
- `5` = NOT_FOUND
- `13` = INTERNAL
- etc.

When testing gRPC endpoints, `expect: status: 200` fails because gRPC returns `0` for success.

## Current Workaround

Remove `status` checks entirely and rely only on `assert` blocks:

```yaml
- name: set_scan_dirs
  url: grpc://localhost:50051
  service: foo.FooService
  rpc: SetScanDirs
  body:
    dirs:
      - /path/to/dir
  # No expect.status - just use assertions on response body
```

## Proposed Solutions

### Option A: Protocol-aware status mapping

Automatically map gRPC status 0 to "success" when checking `status: 200`:

```yaml
expect:
  status: 200  # Works for both HTTP 200 and gRPC OK (0)
```

### Option B: Explicit gRPC status support

Add `grpc_status` field:

```yaml
expect:
  grpc_status: OK  # or grpc_status: 0
```

### Option C: Generic success check

Add a `success: true` field that works across protocols:

```yaml
expect:
  success: true  # HTTP 2xx or gRPC 0
```

## Recommendation

Option C is cleanest for cross-protocol test files. Option B is most explicit for gRPC-specific tests.
