# Environments

Environments let you switch between local, staging, and production configs
without changing your request files.

## Project Config File

Create `yapi.config.yml` in your project root:

```yaml
yapi: v1

default_environment: local

defaults:
  vars:
    API_VERSION: v1
    SHARED_VAR: shared_value

environments:
  local:
    url: http://localhost:3000
    vars:
      API_KEY: dev_key
      DEBUG: "true"

  staging:
    url: https://staging.api.example.com
    vars:
      API_KEY: ${STAGING_API_KEY}

  prod:
    url: https://api.example.com
    vars:
      API_KEY: ${PROD_API_KEY}
    env_file: .env.prod
```

## Key Fields

- **`default_environment`**: Used when `-e` flag is not specified
- **`defaults`**: Variables available in ALL environments
- **`environments`**: Environment-specific settings
  - **`url`**: Base URL for requests using `path:` instead of `url:`
  - **`vars`**: Environment-specific variables
  - **`env_file`**: Path to a `.env` file to load

## Selecting an Environment

```bash
yapi run request.yapi.yml -e staging
yapi test ./tests -e prod
```

Without `-e`, the `default_environment` is used.

## How URL Resolution Works

Request files can use `path:` instead of `url:`:

```yaml
# request.yapi.yml
yapi: v1
path: /api/users
method: GET
```

The `path` is appended to the environment's `url`. So with `-e local`,
the full URL becomes `http://localhost:3000/api/users`.

## Variable Precedence

When the same variable is defined in multiple places:

1. Chain step references (highest priority)
2. Environment-specific `vars`
3. `defaults.vars`
4. Shell environment variables
5. Default values (`${VAR:-default}`)

## Env Files

Load secrets from `.env` files per environment:

```yaml
environments:
  prod:
    env_file: .env.prod
```

```
# .env.prod
API_KEY=sk-prod-abc123
DB_URL=postgres://prod-host/db
```

## See Also

- `yapi docs variables` — Variable interpolation and resolution
- `yapi docs config` — Full YAML config field reference
- `yapi docs testing` — Running tests across environments
