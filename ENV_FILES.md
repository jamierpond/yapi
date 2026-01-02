# env_files Support in Request Files

## Overview

You can now use `env_files` directly in `.yapi.yml` request files to load environment variables from `.env` files without needing a `yapi.config.yml` project configuration.

## Usage

```yaml
yapi: v1
url: ${BASE_URL}/api/v1
method: POST
env_files:
  - .env.local
headers:
  Authorization: Bearer ${API_KEY}
body:
  secret: ${SECRET_VALUE}
```

## Features

- **Multiple env files**: List multiple files; later files override earlier ones
- **Relative paths**: Paths are resolved relative to the config file directory
- **Works with chains**: Each chain step inherits env_files from the base config
- **LSP support**: Autocompletion for `env_files` key in VS Code/Neovim

## Variable Precedence

1. OS environment variables (highest)
2. Project config vars (`yapi.config.yml` environments)
3. Request-level `env_files`
4. Empty string fallback

## Example

```
project/
├── .env.local          # API_KEY=dev-key, BASE_URL=http://localhost:3000
├── .env.secrets        # SECRET_VALUE=my-secret
└── requests/
    └── create-user.yapi.yml
```

```yaml
# requests/create-user.yapi.yml
yapi: v1
url: ${BASE_URL}/users
method: POST
env_files:
  - ../.env.local
  - ../.env.secrets
headers:
  Authorization: Bearer ${API_KEY}
body:
  name: "Test User"
  token: ${SECRET_VALUE}
```

## Changes

### `internal/config/v1.go`
- Added `EnvFiles []string` field to `ConfigV1` struct
- Added `"env_files"` to `knownV1Keys` validation map

### `internal/config/loader.go`
- Added `LoadFromStringWithPath()` for path-aware config loading
- Added `loadEnvFiles()` to load variables from .env files using godotenv
- Added `buildEnvFileResolver()` to create combined resolver with correct precedence

### `internal/validation/analyzer.go`
- Added `AnalyzeConfigStringWithProjectAndPath()` for path-aware analysis
- Added `validateEnvVarsWithEnvFiles()` to suppress false warnings for env_files vars
- Added `extractEnvFileVarNames()` helper

### `internal/core/core.go`
- Updated to use `AnalyzeConfigStringWithProjectAndPath()` with config file path

### `internal/langserver/langserver.go`
- Added `env_files` to LSP autocompletion suggestions

### `internal/config/loader_test.go`
- Added `TestLoadFromStringWithPath_EnvFiles` - basic env file loading
- Added `TestLoadFromStringWithPath_EnvFiles_MultipleFiles` - multiple file merging
- Added `TestLoadFromStringWithPath_EnvFiles_MissingFile` - error handling
- Added `TestLoadFromStringWithPath_EnvFiles_OSEnvTakesPrecedence` - precedence order
