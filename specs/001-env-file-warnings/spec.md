# Feature Specification: Env File Warnings

**Branch**: `001-env-file-warnings` | **Created**: 2026-01-03 | **Status**: Draft

## Overview

When a YAPI configuration references env files via the `env_files` field, users should be warned or receive an error when those files are missing or unreadable. This prevents silent failures where environment variables are expected but unavailable.

**Guiding principle**: The LSP should provide as much diagnostic help as possible, surfacing issues in real-time so developers catch problems before running requests. This includes go-to-definition navigation, warnings for missing files, and alerts for undefined variables.

## Clarifications

### Session 2026-01-03

- Q: Should "go to definition" work only on `env_files` entries, or also on `${VAR}` variable references? → A: Both. `env_files` entries navigate to top of file; `${VAR}` references navigate to the variable definition line in the env file.
- Q: Should missing env files and undefined variables show as LSP diagnostics? → A: Yes, both.
- Q: What is the variable resolution priority? → A: env_files take priority over OS env. If a variable falls back to OS env (when env_files exist), warn about it.
- Q: What does --strict-env do? → A: Only resolve from env_files, no OS env fallback. Error if any env file missing or variable undefined.

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

---

### US5 - Go to Definition for Env Files (P1)

A developer is editing a YAPI file in their IDE and wants to quickly navigate to an env file referenced in `env_files`. They use "go to definition" on the file path and are taken to the top of that file.

**IDE Usage**:
- Cursor on `.env.local` in `env_files` array → Go to Definition → Opens `.env.local` at line 1

**Acceptance**:
- Given a YAPI file has `env_files: [".env.local"]`, when user invokes "go to definition" on `.env.local`, then the IDE opens `.env.local` at line 1
- Given the env file does not exist, when user invokes "go to definition", then no navigation occurs (or appropriate feedback shown)

---

### US6 - Go to Definition for Variable References (P1)

A developer sees a `${GITHUB_PAT}` variable reference in their YAPI file and wants to see where it's defined. They use "go to definition" and are taken to the line in the env file where that variable is declared.

**IDE Usage**:
- Cursor on `${GITHUB_PAT}` in headers → Go to Definition → Opens `.env.local` at line where `GITHUB_PAT=...` is defined

**Acceptance**:
- Given `GITHUB_PAT` is defined on line 5 of `.env.local`, when user invokes "go to definition" on `${GITHUB_PAT}`, then the IDE opens `.env.local` at line 5
- Given `GITHUB_PAT` is defined in multiple env files, when user invokes "go to definition", then the IDE navigates to the first matching definition (per env_files order)
- Given `GITHUB_PAT` is not defined in any env file, when user invokes "go to definition", then no navigation occurs

## Requirements

### Functional

- **FR-001**: `yapi` MUST check for the existence of all files listed in `env_files` before executing a request
- **FR-002**: `yapi` MUST display a warning message for each env file that does not exist, including the file path
- **FR-003**: `yapi` MUST continue execution after displaying warnings for missing env files (unless strict mode enabled)
- **FR-004**: `yapi` MUST support a `--strict-env` CLI flag that:
  - Resolves variables ONLY from env_files (no OS env fallback)
  - Treats missing env files as errors (not warnings)
  - Treats undefined variables as errors
- **FR-005**: `yapi` MUST display an error and halt execution when an env file exists but cannot be read
- **FR-006**: `yapi` MUST output warnings/errors to stderr to distinguish from normal output

### Variable Resolution

- **FR-013**: Variables MUST be resolved in this priority order (highest to lowest):
  1. `environments.[name].vars`
  2. `environments.[name].env_files`
  3. `defaults.vars`
  4. `defaults.env_files`
  5. Per-request `env_files` (in individual .yapi.yml)
  6. OS environment (fallback only)
- **FR-014**: When a variable is resolved from OS env (and any config-level vars/env_files exist), `yapi` MUST display a warning: "variable 'X' resolved from OS environment, not configuration"
- **FR-015**: `yapi` MUST error if a variable is undefined after checking all sources
- **FR-016**: Missing env files in `defaults.env_files` or `environments.[name].env_files` MUST generate warnings (or errors in strict mode)

### LSP Features

The LSP should be maximally helpful - surface every issue that could cause a request to fail before the user runs it.

- **FR-007**: The LSP MUST support "go to definition" on env file paths in the `env_files` array, navigating to line 1 of the target file
- **FR-008**: The LSP MUST support "go to definition" on `${VAR}` references, navigating to the line where the variable is defined in the env file
- **FR-009**: When a variable is defined in multiple env files, the LSP MUST navigate to the first definition in `env_files` order
- **FR-010**: When the target file or variable definition does not exist, the LSP MUST return no definition (no navigation)
- **FR-011**: The LSP MUST show a warning diagnostic on env file paths that do not exist
- **FR-012**: The LSP MUST show a warning diagnostic on `${VAR}` references where the variable is not defined in the current configuration (env_files + project vars)
- **FR-017**: The LSP MUST show a warning diagnostic on `${VAR}` references that would resolve from OS env (not configuration)
- **FR-018**: The LSP MUST show an info diagnostic on `${VAR}` references indicating where the variable was resolved from (e.g., ".env.local:5", "defaults.vars")

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
- What happens when go-to-definition is invoked on a missing env file? - No navigation occurs
- What happens when `${VAR}` references a variable not in any env file? - Diagnostic warning shown, no navigation on go-to-definition

## Assumptions

- Warning messages are written to stderr
- Default behavior (without --strict-env) is to warn and continue, matching common tool behavior
- Env file lookup is relative to the YAPI configuration file location, not the current working directory
- Permission errors are always treated as errors (not warnings) since they indicate a system configuration issue
- LSP diagnostics should be comprehensive and proactive - catch every potential issue before runtime

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
- [ ] LSP go-to-definition works for env file paths and variable references
- [ ] LSP diagnostics appear for missing env files and undefined variables
