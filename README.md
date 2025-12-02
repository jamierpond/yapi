# 🐑 yapi

**The API client that lives in your terminal (and your git repo).**

Stop clicking through heavy Electron apps just to send a JSON body. **yapi** is a CLI-first, offline-first, git-friendly API client for HTTP, gRPC, and TCP. It uses simple YAML files to define requests, meaning you can commit them, review them, and run them anywhere.

[**Try the Playground**](https://yapi.run/playground) | [**View Source**](https://github.com/jamierpond/yapi)

-----

## ⚡ Install

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

### 3\. JQ Filtering (Built-in\!)

Don't grep output. Filter it right in the config.

```yaml
yapi: v1
url: https://jsonplaceholder.typicode.com/users
method: GET

# Only show me names and emails, sorted by name
jq_filter: "[.[] | {name, email}] | sort_by(.name)"
```

### 4\. gRPC (Reflection Support)

Stop hunting for `.proto` files. If your server supports reflection, **yapi** just works.

```yaml
yapi: v1
url: grpc://localhost:50051
service: helloworld.Greeter
rpc: SayHello

body:
  name: "yapi User"
```

### 5\. GraphQL

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
  * **LSP Support:** You get autocompletion for keys (`method`, `headers`, etc.) and validation right in your editor\!

### VS Code / Others

Since **yapi** includes a Language Server (`yapi lsp`), you can hook it up to any editor that supports LSP over stdio.

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
