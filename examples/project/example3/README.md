# Example 3: GitHub API with Environments

This example demonstrates using yapi with a real API (GitHub) and multiple environments to query different users/organizations.

## Structure

```
example3/
├── yapi.config.yml          # Project config with environments
└── .yapi/
    ├── get-user.yapi.yml    # Get user/org information
    ├── get-repo.yapi.yml    # Get specific repository
    └── list-repos.yapi.yml  # List repositories
```

## Environments

This example defines three environments:

- **anthropic** (default): Query Anthropic's GitHub organization
  - User: `anthropics`
  - Repo: `anthropic-sdk-python`

- **personal**: Query Linus Torvalds' repositories
  - User: `torvalds`
  - Repo: `linux`

- **example**: Query the Octocat example user
  - User: `octocat`
  - Repo: `hello-world`

## Usage

### Using default environment (anthropic)

```bash
yapi run .yapi/get-user.yapi.yml
yapi run .yapi/get-repo.yapi.yml
yapi run .yapi/list-repos.yapi.yml
```

### Switching environments

```bash
# Query Linus Torvalds' repos
yapi run .yapi/list-repos.yapi.yml --env personal

# Query Octocat
yapi run .yapi/get-user.yapi.yml --env example

# Get specific repo info
yapi run .yapi/get-repo.yapi.yml --env personal  # torvalds/linux
```

### Testing from repository root

```bash
cd /path/to/yapi

# Default environment
yapi run examples/project/example3/.yapi/get-user.yapi.yml

# Specific environment
yapi run examples/project/example3/.yapi/list-repos.yapi.yml --env personal
```

## What This Demonstrates

1. **Real API Integration**: Uses GitHub's public API (no auth required)
2. **Environment Variables**: Different GitHub users/repos per environment
3. **Default Values**: Common values (API base URL, headers) in defaults
4. **Automatic Environment Selection**: Uses `default_environment` when `--env` not specified
5. **Expectations**: All requests include assertions to validate responses

## Key Features

- **No authentication needed**: Uses GitHub's public API endpoints
- **Multiple environments**: Easily switch between different GitHub users/orgs
- **Shared defaults**: API base URL and headers defined once
- **JQ assertions**: Validates response structure and content
- **Real data**: Fetches actual repository information from GitHub

## Try It Out

```bash
cd examples/project/example3

# Get Anthropic org info (default)
yapi run .yapi/get-user.yapi.yml

# List Torvalds' repositories
yapi run .yapi/list-repos.yapi.yml --env personal

# Get the Linux kernel repo info
yapi run .yapi/get-repo.yapi.yml --env personal
```
