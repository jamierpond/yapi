# Request Chaining

Chains let you execute multiple requests sequentially, passing data between steps.
Login, use the token, verify the result — all in one file.

## Basic Chain

```yaml
yapi: v1
chain:
  - name: login
    url: https://api.example.com/auth/login
    method: POST
    body:
      username: ${USERNAME}
      password: ${PASSWORD}
    expect:
      status: 200
      assert:
        - .token != null

  - name: get_profile
    url: https://api.example.com/me
    method: GET
    headers:
      Authorization: Bearer ${login.token}
    expect:
      status: 200
```

## Step References

Reference data from previous steps with `${step_name.field}`:

```yaml
- name: create_user
  url: /api/users
  method: POST
  body: { name: "Alice" }
  expect:
    assert:
      - .id != null

- name: get_user
  url: /api/users/${create_user.id}
  method: GET
```

### Nested Paths

Access nested JSON fields with dot notation:

```yaml
${step_name.data.user.email}     # Nested object
${step_name.items[0].id}         # Array indexing
```

### Where References Work

Step references work in any string field: URLs, headers, body values, assertions.

```yaml
- name: verify
  url: /api/users/${create.id}
  headers:
    Authorization: Bearer ${login.token}
  expect:
    assert:
      - .email == env.EXPECTED_EMAIL
```

## Type Preservation

When a `${ref}` is the entire value (not part of a larger string), its type is preserved:

```yaml
body:
  user_id: ${get_user.id}      # Stays an integer
  name: "User ${get_user.id}"  # String interpolation
```

## Fail-Fast Behavior

Chains stop on the first failure:
- If a request fails (network error, timeout), the chain stops
- If an assertion fails, the chain stops
- Subsequent steps are skipped

## Step Config Inheritance

Each step can override any base-level config field. Steps inherit from the
top-level config (URL, headers, timeout, etc.):

```yaml
yapi: v1
timeout: 10s                    # Default for all steps

chain:
  - name: fast_check
    url: /health
    timeout: 2s                  # Override for this step

  - name: slow_op
    url: /process
    timeout: 30s                 # Override for this step
```

## See Also

- `yapi docs assert` — Assertions on status, body, and headers
- `yapi docs variables` — Variable interpolation and resolution
- `yapi docs polling` — Polling with wait_for
