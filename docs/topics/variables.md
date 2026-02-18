# Variables

Variables let you parameterize requests — URLs, headers, bodies, assertions.
Use `${VAR}` syntax anywhere in your YAML files.

## Syntax

```yaml
url: ${BASE_URL}/api/users/${USER_ID}
headers:
  Authorization: Bearer ${API_KEY}
```

## Default Values

Provide fallbacks with `:-`:

```yaml
url: ${BASE_URL:-http://localhost:3000}/api/users
headers:
  X-Timeout: ${TIMEOUT:-30}
```

If `BASE_URL` is not set, `http://localhost:3000` is used.

## Resolution Order

Variables resolve in this priority (highest first):

1. **Chain step references**: `${step_name.field}` — data from previous chain steps
2. **Environment vars from yapi.config.yml** — vars defined in the active environment
3. **Shell environment variables** — from your OS/shell
4. **Default values** — specified with `:-`

## Type Preservation

When `${VAR}` is the entire value, the original type is preserved:

```yaml
body:
  count: ${step.count}          # Stays an integer if count is int
  label: "Count: ${step.count}" # String interpolation
```

## Environment Variable References in Assertions

In assertions, use `env.VAR_NAME` to compare against environment variables:

```yaml
expect:
  assert:
    - .owner == env.GITHUB_USER
    - .region == env.AWS_REGION
```

## Variables from Config

Define variables in `yapi.config.yml`:

```yaml
yapi: v1
default_environment: local

defaults:
  vars:
    API_VERSION: v1

environments:
  local:
    url: http://localhost:3000
    vars:
      API_KEY: dev_key
```

Then reference them in request files:

```yaml
yapi: v1
url: ${url}/api/${API_VERSION}/users
headers:
  X-Api-Key: ${API_KEY}
```

## Env Files

Load variables from `.env` files:

```yaml
# yapi.config.yml
environments:
  prod:
    url: https://api.example.com
    env_file: .env.prod
```

Or per-request:

```yaml
yapi: v1
env_files:
  - .env.local
url: ${BASE_URL}/api/users
```

## See Also

- `yapi docs environments` — Multi-environment configuration
- `yapi docs chain` — Chain step references
- `yapi docs assert` — Using env vars in assertions
