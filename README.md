# yapi 🐏

Tired of `curl` novels and heavyweight API tools for one tiny request? **yapi** is a small, Bash-powered YAML API client that speaks HTTP, gRPC, and raw TCP.

You write clean YAML, yapi does the ugly shell work.

---

## Quick Taste

Instead of this `curl` command:

```bash
curl -X POST 'https://httpbin.org/post' \
  -H 'Content-Type: application/json' \
  -d '{"title":"Testing yapi","description":"...","userId":123,"isPublished":true,"tags":["testing","api","yaml"]}'
```

You write this `yapi` config:

```yaml
# examples/create-post.yapi.yml
# yaml-language-server: $schema=https://yapit.dev/schema/v1

url: https://httpbin.org
path: /post
method: POST
content_type: application/json

body:
  title: "Testing yapi - YAML API Testing Tool"
  description: "This demo shows nested objects, arrays, and various data types"
  userId: 123
  isPublished: true
  tags:
    - testing
    - api
    - yaml
```

Then just run `yapi`:

```bash
yapi -c examples/create-post.yapi.yml
```

---

## Installation

1.  **Clone the repo** somewhere permanent, like `~/.config/yapi`.
    ```bash
    git clone https://github.com/jamierpond/yapi.git ~/.config/yapi
    ```

2.  **Make scripts executable**.
    ```bash
    chmod +x ~/.config/yapi/yapi ~/.config/yapi/lib/*.sh
    ```

3.  **Add to your `$PATH`** in `~/.zshrc` or `~/.bashrc`.
    ```sh
    export PATH="$HOME/.config/yapi:$PATH"
    ```

4.  **Reload your shell** and you're good to go.
    ```bash
    source ~/.zshrc
    yapi -h
    ```

---

## Features

*   **Declarative YAML** configs for clean, version-controllable API requests.
*   **Multi-Protocol**: Support for HTTP/REST, gRPC, and raw TCP.
*   **`fzf` Integration**: Interactive file picker for `*.yapi.yml` files.
*   **History**: All runs are logged to `~/.yapi_history` for easy re-use.
*   **Editor Support**: JSON Schema for autocomplete and validation in your editor.
*   **Lightweight**: Just a few shell scripts and common command-line tools.

---

## Usage

```bash
# Interactive selection (fzf over git tracked *.yapi.yml)
yapi

# Explicit config file
yapi -c examples/google.yapi.yml

# Override base URL at runtime
yapi -c examples/google.yapi.yml -u "https://httpbin.org/get"

# Search all YAML files, not just git tracked
yapi --all
```

---

## Supported Protocols

### HTTP / REST

Supports standard verbs, query params, and JSON bodies.

```yaml
url: https://httpbin.org
path: /post
method: POST
query: { search: "yapi" }
body:
  name: "yapi demo"
  isPublished: true
```

### gRPC (via `grpcurl`)

Uses server reflection by default. Can also use local `.proto` files.

```yaml
url: grpc://grpcb.in:9000
method: grpc
service: hello.HelloService
rpc: SayHello
body:
  greeting: "yapi"
```

### TCP (via `nc`)

Sends raw data over TCP. Supports text, hex, and base64 encoding.

```yaml
url: tcp://tcpbin.com:4242
method: tcp
data: "Hello from yapi!\n"
encoding: text
```

---

## Dependencies

-   **Core**: `bash`, `curl`, `jq`, `yq` (mikefarah), `git`
-   **Optional**: `fzf` (for picker), `grpcurl` (for gRPC), `nc` (for TCP)

---

## Development

-   **Tests**: Written in `bats`. Run with `bats test/*.bats`.
-   **Schema**: If you change `yapi` behavior, keep `yapi.schema.json` in sync.
-   **Dependencies**: If you add a new dependency, add it to the `.depends` file.

Contributions are welcome!
