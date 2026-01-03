# Feature Specification: Env File Warnings

**Branch**: `001-env-file-warnings` | **Created**: 2026-01-03 | **Status**: Draft

## Overview

When a YAPI configuration references env files via the `env_files` field, users should be warned or receive an error when those files are missing or unreadable. This prevents silent failures where environment variables are expected but unavailable.

## User Stories

### US1 - Missing Env File Warning (P1)

A developer runs a YAPI request that references an env file (e.g., `.env.local`) which doesn't exist on their machine. The developer receives a clear warning message indicating which file is missing, allowing them to either create the file or remove the reference.

**CLI Usage**:
```bash
yapi run my-request.yapi.yml
```

**Acceptance**:
- Given a YAPI file references `.env.local` in `env_files`, when `.env.local` does not exist, then YAPI displays a warning message: "Warning: env file '.env.local' not found"
- Given the warning is displayed, when the request is executed, then YAPI continues execution (warning does not block)

---

### US2 - Strict Mode for Env Files (P1)

A developer wants to ensure all env files are present before execution to prevent partial configuration issues. They enable strict mode to treat missing env files as errors rather than warnings.

**CLI Usage**:
```bash
yapi run my-request.yapi.yml --strict-env
```

**Acceptance**:
- Given a YAPI file references `.env.local` in `env_files`, when `.env.local` does not exist and `--strict-env` flag is used, then YAPI exits with an error before executing the request
- Given `--strict-env` is not specified, when an env file is missing, then YAPI only warns and continues

---

### US3 - Multiple Env File Validation (P2)

A developer has a configuration with multiple env files where some exist and some don't. They should see warnings for each missing file, with clear identification of which files are problematic.

**CLI Usage**:
```bash
yapi run my-request.yapi.yml
```

**Acceptance**:
- Given a YAPI file references `[".env", ".env.local", ".env.secrets"]` in `env_files`, when `.env` exists but `.env.local` and `.env.secrets` do not, then YAPI displays two separate warnings for the missing files
- Given multiple files are missing, when the warnings are displayed, then each warning clearly identifies the specific file path

---

### US4 - Unreadable Env File Error (P2)

A developer has an env file that exists but cannot be read due to permission issues. They should receive a clear error explaining the access problem.

**CLI Usage**:
```bash
yapi run my-request.yapi.yml
```

**Acceptance**:
- Given a YAPI file references `.env.local` in `env_files`, when `.env.local` exists but is not readable, then YAPI displays an error: "Error: cannot read env file '.env.local': permission denied"
- Given a file permission error occurs, when `--strict-env` is used, then YAPI exits with non-zero status

## Requirements

### Functional

- **FR-001**: `yapi` MUST check for the existence of all files listed in `env_files` before executing a request
- **FR-002**: `yapi` MUST display a warning message for each env file that does not exist, including the file path
- **FR-003**: `yapi` MUST continue execution after displaying warnings for missing env files (unless strict mode enabled)
- **FR-004**: `yapi` MUST support a `--strict-env` CLI flag that treats missing env files as errors
- **FR-005**: `yapi` MUST display an error and halt execution when an env file exists but cannot be read
- **FR-006**: `yapi` MUST output warnings/errors to stderr to distinguish from normal output

### YAML Schema (if applicable)

```yaml
yapi: v1
# Existing field - no schema changes required
env_files:
  - .env.local
  - .env.secrets
```

### Protocol Support

| Protocol | Supported | Notes                        |
|----------|-----------|------------------------------|
| HTTP     | [x]       | Env files used in headers, URLs, body |
| gRPC     | [x]       | Env files used in metadata, messages |
| GraphQL  | [x]       | Env files used in variables, headers |
| TCP      | [x]       | Env files used in connection params |

## Edge Cases

- What happens when env_files is an empty array? - YAPI proceeds normally with no warnings
- What happens when an env file path is absolute vs relative? - Both are supported; relative paths resolved from YAPI file directory
- What happens when env file exists but is empty? - YAPI proceeds normally (empty file is valid)
- What happens when the same env file is listed twice? - Only validate/warn once per unique file
- What happens on Windows vs Unix file paths? - Use OS-appropriate path handling

## Assumptions

- Warning messages are written to stderr
- Default behavior (without --strict-env) is to warn and continue, matching common tool behavior
- Env file lookup is relative to the YAPI configuration file location, not the current working directory
- Permission errors are always treated as errors (not warnings) since they indicate a system configuration issue

## Success Criteria

- [ ] Feature works via CLI without GUI
- [ ] Configuration stored in `.yapi.yml` files
- [ ] Works across all applicable protocols
- [ ] Minimal implementation, no unnecessary complexity
- [ ] Tests pass: `make test`
- [ ] Lint passes: `make lint`
- [ ] 100% of users encountering missing env files see a clear warning message
- [ ] Users can distinguish between missing file warnings and permission errors
- [ ] Strict mode exits with non-zero status code when env files are missing
