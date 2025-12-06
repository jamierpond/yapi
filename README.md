# 🐑 yapi

[![CLI](https://github.com/jamierpond/yapi/actions/workflows/cli.yml/badge.svg)](https://github.com/jamierpond/yapi/actions/workflows/cli.yml)
[![Playground](https://yapi.run/badge.svg)](https://yapi.run/playground)

**The API client that lives in your terminal (and your git repo).**

Stop clicking through heavy Electron apps just to send a JSON body. **yapi** is a CLI-first, offline-first, git-friendly API client for HTTP, gRPC, and TCP. It uses simple YAML files to define requests, meaning you can commit them, review them, and run them anywhere.

[**Try the Playground**](https://yapi.run/playground) | [**View Source**](https://github.com/jamierpond/yapi)

-----

## ⚡ Install

**Using Homebrew (macOS):**

```bash
brew tap jamierpond/yapi
brew install --cask yapi
```

**Linux (from GitHub Releases):**

```bash
# For amd64
curl -sL https://github.com/jamierpond/yapi/releases/latest/download/yapi_linux_amd64.tar.gz | tar xz
sudo mv yapi /usr/local/bin/

# For arm64
curl -sL https://github.com/jamierpond/yapi/releases/latest/download/yapi_linux_arm64.tar.gz | tar xz
sudo mv yapi /usr/local/bin/
```

**Windows (PowerShell):**

```powershell
# For amd64
Invoke-WebRequest -Uri "https://github.com/jamierpond/yapi/releases/latest/download/yapi_windows_amd64.zip" -OutFile yapi.zip
Expand-Archive yapi.zip -DestinationPath .
Move-Item yapi.exe "$env:LOCALAPPDATA\Microsoft\WindowsApps\"

# For arm64
Invoke-WebRequest -Uri "https://github.com/jamierpond/yapi/releases/latest/download/yapi_windows_arm64.zip" -OutFile yapi.zip
Expand-Archive yapi.zip -DestinationPath .
Move-Item yapi.exe "$env:LOCALAPPDATA\Microsoft\WindowsApps\"
```

**Using Go:**

```bash
go install yapi.run/cli/cmd/yapi@latest
```

**From Source:**

```bash
git clone https://github.com/jamierpond/yapi
cd yapi
make install
```

-----

## 🚀 Quick Start

1.  **Create a request file** (e.g., `get-user.yapi.yml`):

    ```yaml
    yapi: v1
    url: https://jsonplaceholder.typicode.com/users/1
    method: GET
    ```

2.  **Run it:**

    ```bash
    yapi run get-user.yapi.yml
    ```

3.  **See the magic:** You get a beautifully highlighted, formatted response.

> **Note:** The `yapi: v1` version tag is required at the top of all config files. This enables future schema evolution while maintaining backwards compatibility.

-----

## 📚 Examples

**yapi** speaks many protocols. Here is how you define them.

### 1\. HTTP with JSON Body

No more escaping quotes in curl.

```yaml
yapi: v1
url: https://api.example.com/posts
method: POST
content_type: application/json

# yapi handles the JSON encoding for you
body:
  title: "Hello World"
  tags:
    - cli
    - testing
  author:
    id: 123
    active: true
```

### 2\. Environment Variables & Chaining

Use shell environment variables directly. Perfect for CI/CD or keeping secrets out of git.

```yaml
yapi: v1
url: https://api.example.com/secure-data
method: GET

headers:
  Authorization: Bearer ${API_TOKEN}
  X-Request-ID: ${REQUEST_ID:-default_id}
```

### 3\. Request Chaining

Chain multiple requests together, referencing data from previous steps. Build authentication flows, integration tests, or multi-step workflows.

```yaml
yapi: v1
chain:
  - name: get_todo
    url: https://jsonplaceholder.typicode.com/todos/1
    method: GET
    expect:
      status: 200
      assert:
        - .userId != null
        - .id == 1

  - name: get_user
    url: https://jsonplaceholder.typicode.com/users/${get_todo.userId}
    method: GET
    expect:
      status: 200
      assert:
        - .name != null
```

**Key features:**
- Reference previous step data with `${step_name.field}` syntax
- Access nested JSON properties: `${login.data.token}`
- Assertions use JQ expressions that must evaluate to true
- Chains stop on first failure (fail-fast)

### 4\. Assertions & Expectations

Validate responses inline with JQ-powered assertions. No separate test framework needed.

```yaml
yapi: v1
url: https://api.example.com/users
method: GET
expect:
  status: 200              # or [200, 201] for multiple valid codes
  assert:
    - . | length > 0       # array has items
    - .[0].email != null   # first item has email
    - .[] | .active == true # all items are active
```

### 5\. JQ Filtering (Built-in\!)

Don't grep output. Filter it right in the config.

```yaml
yapi: v1
url: https://jsonplaceholder.typicode.com/users
method: GET

# Only show me names and emails, sorted by name
jq_filter: "[.[] | {name, email}] | sort_by(.name)"
```

### 6\. gRPC (Reflection Support)

Stop hunting for `.proto` files. If your server supports reflection, **yapi** just works.

```yaml
yapi: v1
url: grpc://localhost:50051
service: helloworld.Greeter
rpc: SayHello

body:
  name: "yapi User"
```

### 7\. GraphQL

First-class support for queries and variables.

```yaml
yapi: v1
url: https://countries.trevorblades.com/graphql

graphql: |
  query getCountry($code: ID!) {
    country(code: $code) {
      name
      capital
    }
  }

variables:
  code: "BR"
```

-----

## 🎛️ Interactive Mode (TUI)

Don't remember the file name? Just run `yapi` without arguments.

```bash
yapi
```

This launches the **Interactive TUI**. You can fuzzy-search through all your `.yapi.yml` files in the current directory (and subdirectories) and execute them instantly.

### 👀 Watch Mode

Tired of `Alt-Tab` -\> `Up Arrow` -\> `Enter`? Use watch mode to re-run the request every time you save the file.

```bash
yapi watch ./my-request.yapi.yml
```

-----

## 🧠 Editor Integration

### Language Server

**yapi** ships with a full Language Server Protocol (LSP) implementation that brings IDE-level intelligence to your `.yapi.yml` files:

**Real-time Validation** - Errors and warnings appear as you type, with precise line/column positions. The LSP runs the same validation as the CLI, catching issues before you execute.

**Intelligent Autocompletion** - Context-aware suggestions for:
- Top-level keys (`url`, `method`, `headers`, `body`, `query`, `graphql`, `jq_filter`, etc.)
- HTTP methods (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS`)
- Encoding types (`text`, `hex`, `base64`)
- Boolean fields (`insecure`, `plaintext`, `close_after_send`)

**Hover Information** - Hover over `${VAR}` references to see environment variable status and (redacted) values.

### Neovim

**yapi** was built with Neovim in mind. We have a native plugin in `lua/yapi_nvim`.

**Lazy.nvim setup:**

```lua
{
  dir = "~/path/to/yapi/lua/yapi_nvim", -- or point to your installed path
  config = function()
    require("yapi_nvim").setup({
      lsp = true,    -- Enables the yapi Language Server
      pretty = true, -- Uses the TUI renderer in the popup
    })
  end
}
```

  * `:YapiRun` - Run the current buffer.
  * `:YapiWatch` - Open a split and watch the current buffer.

### VS Code / Others

The LSP runs over stdio and works with any editor that supports the protocol:

```bash
yapi lsp
```

-----

## 📂 Project Structure

  * `cmd/yapi`: The main CLI entry point.
  * `internal/executor`: The brains. HTTP, gRPC, TCP, and GraphQL logic.
  * `internal/tui`: The BubbleTea-powered interactive UI.
  * `examples/`: **Look here for a ton of practical YAML examples\!**
  * `webapp/`: The Next.js code for [yapi.run](https://yapi.run).

-----

## 🤝 Contributing

Found a bug? Want to add WebSocket support? PRs are welcome\!

1.  Fork it.
2.  `make build` to ensure it compiles.
3.  `make test` to run the suite.
4.  Ship it.

-----

*Made with ☕ and Go.*
