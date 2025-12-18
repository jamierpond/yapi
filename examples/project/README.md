# Project Configuration Examples

These examples demonstrate the **project-aware environment system** in yapi, which enables comprehensive diagnostics and environment-specific variable management.

## Overview

When a `yapi.config.yml` file is present in your project, yapi becomes **project-aware** and provides:

- ✅ Cross-environment variable validation
- ⚠️ Smart diagnostics (e.g., "variable missing in prod")
- 🔄 Environment-specific execution with `--env` flag
- 📁 Support for `.env` files
- 🔍 LSP integration for real-time validation

## Examples

### Example 1: Simple Variable Configuration

**Location:** `examples/project/example1/`

Demonstrates basic environment configuration with inline variables.

```bash
# View the config
cat examples/project/example1/yapi.config.yml

# Run with different environments
cd examples/project/example1
yapi run .yapi/get-users.yapi.yml --env dev
yapi run .yapi/get-users.yapi.yml --env staging
yapi run .yapi/get-users.yapi.yml --env prod
```

**Key features:**
- Simple inline `vars` definitions
- Environment-specific overrides
- Default values

### Example 2: Using .env Files

**Location:** `examples/project/example2/`

Shows advanced configuration using `.env` files for secret management.

```bash
# View the config and env files
cat examples/project/example2/yapi.config.yml
cat examples/project/example2/.env.dev

# Run authenticated requests
cd examples/project/example2
yapi run .yapi/authenticated-request.yapi.yml --env dev
yapi run .yapi/authenticated-request.yapi.yml --env prod
```

**Key features:**
- `.env` file loading
- Environment-specific secrets
- Combined inline + file variables

### Example 3: Real API with GitHub

**Location:** `examples/project/example3/`

Working example using the GitHub API to query different users/organizations.

```bash
# View the config
cat examples/project/example3/yapi.config.yml

# Query Anthropic org (default)
cd examples/project/example3
yapi run .yapi/get-user.yapi.yml

# Query Torvalds repos
yapi run .yapi/list-repos.yapi.yml --env personal

# Get Linux kernel info
yapi run .yapi/get-repo.yapi.yml --env personal

# Query Octocat
yapi run .yapi/get-user.yapi.yml --env example
```

**Key features:**
- Real API integration (GitHub)
- No authentication required
- Multiple environments for different users/orgs
- JQ assertions on real data
- Default environment auto-selection

## Configuration Schema

### `yapi.config.yml` Structure

```yaml
yapi: v1
kind: project

# Default variables (applied to all environments)
defaults:
  env_files:
    - .env
  vars:
    TIMEOUT: "30"
    LOG_LEVEL: info

# Environment definitions
environments:
  dev:
    env_files:
      - .env.dev
    vars:
      API_URL: http://localhost:8080

  prod:
    env_files:
      - .env.prod
    vars:
      API_URL: https://api.example.com
      LOG_LEVEL: warn  # Override default
```

## Variable Resolution Order

Variables are resolved in this priority order (highest to lowest):

1. **OS Environment Variables** - `export API_KEY=xyz`
2. **Environment-specific vars** - `environments.prod.vars.API_KEY`
3. **Environment-specific .env files** - `environments.prod.env_files`
4. **Default vars** - `defaults.vars.API_KEY`
5. **Default .env files** - `defaults.env_files`

## Smart Diagnostics

The LSP and CLI provide comprehensive diagnostics:

### ❌ Error: Variable Not Defined Anywhere
```yaml
url: ${UNDEFINED_VAR}/api
```
**Diagnostic:** `Variable 'UNDEFINED_VAR' is not defined in any environment or defaults`

### ⚠️ Warning: Missing in Some Environments
```yaml
# yapi.config.yml
environments:
  dev:
    vars:
      API_KEY: dev-key
  prod:
    # API_KEY not defined!
    vars:
      OTHER_VAR: value
```

**Diagnostic:** `Variable 'API_KEY' is missing in environment(s): prod`

### ✅ Valid: Defined Everywhere
```yaml
# Either in defaults OR in all environments
defaults:
  vars:
    API_KEY: default-key  # ✅ Safe - all envs inherit
```

## CLI Usage

### Running with Environments

```bash
# Run a single request
yapi run request.yapi.yml --env dev

# Watch mode with environment
yapi watch request.yapi.yml --env staging

# Run all tests with environment
yapi test --env prod
```

### Project Root Detection

yapi automatically finds `yapi.config.yml` by walking up the directory tree from your request file:

```
project/
├── yapi.config.yml        # ← Found automatically
├── api/
│   └── users/
│       └── get.yapi.yml   # ← Run from here
```

## LSP Integration

When editing `.yapi.yml` files in your editor (VSCode, Neovim, etc.), the LSP provides:

1. **Real-time validation** - See diagnostics as you type
2. **Environment awareness** - Validates against all defined environments
3. **Hover information** - See variable values and definitions
4. **Auto-completion** - Suggests available variables

## Best Practices

### 1. Use Defaults for Common Values
```yaml
defaults:
  vars:
    TIMEOUT: "30"
    CONTENT_TYPE: application/json
```

### 2. Keep Secrets in .env Files
```yaml
environments:
  prod:
    env_files:
      - .env.prod  # Contains API_KEY, JWT_SECRET, etc.
```

### 3. Validate All Environments
Ensure variables are defined in all environments or in defaults to avoid warnings.

### 4. Use Descriptive Environment Names
- `dev` / `local` - Local development
- `staging` / `test` - Pre-production testing
- `prod` / `production` - Production environment

## Troubleshooting

### "No yapi.config.yml found"
The `--env` flag requires a project config. Make sure `yapi.config.yml` exists in your project root or a parent directory.

### "Environment 'xyz' not found"
Check available environments:
```bash
cat yapi.config.yml | grep -A 1 "environments:"
```

### Variables Not Resolving
1. Check variable name spelling
2. Verify it's defined in the target environment or defaults
3. Run with `--env` flag: `yapi run file.yapi.yml --env dev`

## Migration from Environment Variables

If you're currently using OS environment variables, you can gradually migrate:

```yaml
# yapi.config.yml
defaults:
  vars:
    # Define your existing vars here
    API_KEY: ${API_KEY}  # Falls back to OS env if not overridden
```

This allows mixing project-defined and OS-defined variables during migration.
