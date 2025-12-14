# What is yapi?

Welcome to the yapi blog. This is where we'll share updates, tutorials, and thoughts about yapi development.

## What is yapi?

yapi is a YAML-powered API client that runs from your terminal. Define your requests in YAML, version them with git, and execute them anywhere.

### Key Features

- **Go Native Speed** - Written in Go, starts instantly, minimal RAM
- **Team Friendly** - Review API changes in Pull Requests
- **Built-in LSP** - Full Language Server with autocompletion

## Getting Started

Install yapi with a single command:

```bash
curl -fsSL https://yapi.sh/install | sh
```

Then create your first request:

```yaml
url: https://api.example.com/users
method: GET
headers:
  Authorization: Bearer ${API_KEY}
```

Run it:

```bash
yapi run users.yapi.yml
```

## Stay Tuned

More posts coming soon covering advanced features like request chaining, assertions, and team workflows.
