# Assertions

Assertions let you validate HTTP responses — status codes, body content, and headers.
They're the core of yapi's testing capability.

## Status Expectations

```yaml
expect:
  status: 200              # Exact match
  status: [200, 201, 204]  # Any of these
```

## Body Assertions

Body assertions are JQ expressions that must evaluate to `true`:

```yaml
expect:
  status: 200
  assert:
    - .id != null
    - .email != null
    - .active == true
```

### Operators

All JQ comparison operators work: `==`, `!=`, `>`, `>=`, `<`, `<=`

```yaml
assert:
  - . | length > 0          # Array has items
  - .count >= 10             # Numeric comparison
  - .name != ""              # Not empty string
```

### Array Operations

```yaml
assert:
  - . | type == "array"
  - . | length > 0
  - .[0].name != null        # First element
  - .[] | .status == "active" # All items match
```

### Type Checks

```yaml
assert:
  - . | type == "array"
  - .data | type == "object"
  - .count | type == "number"
```

## Environment Variable References

Compare response values against environment variables using `env.VAR_NAME`:

```yaml
assert:
  - .owner.login == env.GITHUB_USER
  - .email == env.EXPECTED_EMAIL
```

## Header Assertions

Use the grouped syntax to assert on response headers:

```yaml
expect:
  status: 200
  assert:
    headers:
      - .["content-type"] | startswith("application/json")
      - .["x-request-id"] != null
    body:
      - .id != null
```

When using the grouped syntax, body assertions go under `body:`.

## Flat vs Grouped Syntax

**Flat** (all assertions are body assertions):
```yaml
assert:
  - .id != null
```

**Grouped** (separate body and header assertions):
```yaml
assert:
  headers:
    - .["content-type"] | contains("json")
  body:
    - .id != null
```

## Assertions in Chains

Each chain step can have its own `expect` block:

```yaml
chain:
  - name: create
    url: /api/items
    method: POST
    body: { title: "test" }
    expect:
      status: 201
      assert:
        - .id != null

  - name: verify
    url: /api/items/${create.id}
    method: GET
    expect:
      status: 200
      assert:
        - .title == "test"
```

## See Also

- `yapi docs chain` — Multi-step request chaining
- `yapi docs jq` — JQ filtering and expressions
- `yapi docs variables` — Variable interpolation
- `yapi docs testing` — Test runner and CI/CD
