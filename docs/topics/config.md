# Config Reference

Every yapi request file starts with `yapi: v1`. This page lists all available fields.

## Project Layout

Place a `yapi.config.yml` at your project root to define environments and base URLs. Request files (`.yapi.yml`) can live anywhere beneath it — yapi walks up the directory tree to find the nearest config.

```
my-project/
├── yapi.config.yml          # project config (environments, base URLs)
└── yapi/
    ├── homepage.yapi.yml    # GET /
    ├── sitemap.yapi.yml     # GET /sitemap.xml
    └── health.yapi.yml      # GET /healthz
```

### Example: `yapi.config.yml`

```yaml
yapi: v1

default_environment: local

environments:
  local:
    url: http://localhost:3000
  prod:
    url: https://api.example.com
```

### Example: `yapi/homepage.yapi.yml`

```yaml
yapi: v1
path: /
method: GET

headers:
  User-Agent: yapi-cli

expect:
  status: 200
```

Because the request file uses `path: /` instead of a full `url`, yapi resolves it against the active environment's base URL. Running `yapi run yapi/homepage.yapi.yml` hits `http://localhost:3000/` by default, or `https://api.example.com/` with `-e prod`.

## Request Fields

| Field | Type | Description |
|---|---|---|
| `yapi` | string | **Required.** Version tag. Always `v1`. |
| `url` | string | Full request URL |
| `path` | string | Path appended to environment base URL |
| `method` | string | HTTP method: GET, POST, PUT, PATCH, DELETE |
| `headers` | map | Request headers |
| `query` | map | Query parameters |
| `timeout` | string | Request timeout (e.g., `"4s"`, `"100ms"`) |
| `delay` | string | Wait before executing (e.g., `"5s"`) |
| `insecure` | bool | Skip TLS verification |

`url` and `path` are mutually exclusive. Use `path` when `yapi.config.yml` provides a base URL.

## Body Fields

These are mutually exclusive — use only one:

| Field | Type | Description |
|---|---|---|
| `body` | map | JSON object body |
| `json` | string | Raw JSON string body |
| `form` | map | Form-encoded body |

## Response Processing

| Field | Type | Description |
|---|---|---|
| `jq_filter` | string | JQ expression to transform response |
| `output_file` | string | Save response body to file |
| `content_type` | string | Override content type |

## GraphQL Fields

| Field | Type | Description |
|---|---|---|
| `graphql` | string | GraphQL query or mutation |
| `variables` | map | GraphQL variables |

## gRPC Fields

| Field | Type | Description |
|---|---|---|
| `service` | string | gRPC service name |
| `rpc` | string | RPC method name |
| `proto` | string | Path to .proto file |
| `proto_path` | string | Proto import path |
| `plaintext` | bool | No TLS for gRPC |

## TCP Fields

| Field | Type | Description |
|---|---|---|
| `data` | string | Raw data to send |
| `encoding` | string | `text` (default), `hex`, `base64` |
| `read_timeout` | int | Seconds to wait for response |
| `idle_timeout` | int | Milliseconds before response is considered complete |
| `close_after_send` | bool | Close connection after sending |

## Testing Fields

| Field | Type | Description |
|---|---|---|
| `expect` | object | Status and assertion expectations |
| `expect.status` | int/[]int | Expected status code(s) |
| `expect.assert` | list/map | Body and header assertions |
| `wait_for` | object | Polling configuration |
| `chain` | list | Multi-step request chain |

## Environment Fields

| Field | Type | Description |
|---|---|---|
| `env_files` | []string | Paths to .env files to load |

## Project Config (`yapi.config.yml`)

| Field | Type | Description |
|---|---|---|
| `default_environment` | string | Environment used without `-e` |
| `defaults.vars` | map | Variables for all environments |
| `environments` | map | Environment definitions |
| `environments.{name}.url` | string | Base URL |
| `environments.{name}.vars` | map | Environment-specific variables |
| `environments.{name}.env_file` | string | .env file path |

## See Also

- `yapi docs protocols` — Protocol-specific details and examples
- `yapi docs variables` — Variable interpolation
- `yapi docs assert` — Assertion syntax
