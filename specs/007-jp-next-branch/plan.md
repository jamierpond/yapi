# Implementation Plan: Next Branch Workflow

**Branch**: `007-jp-next-branch` | **Date**: 2026-01-04 | **Spec**: [spec.md](./spec.md)

## Summary

Establish a `next` branch as a staging/integration point by updating GitHub workflows to run CI on `next`, updating release scripts to allow releases from `next`, and documenting manual configuration steps for Vercel and GitHub branch protection.

## Technical Context

**Type**: Infrastructure/DevOps - no application code changes
**Files**: GitHub workflow YAML files, shell scripts, Makefile
**Testing**: Manual verification of workflow triggers
**Build**: N/A (configuration changes only)

## Constitution Check

*Must pass before implementation.*

| Principle | Status | Notes |
|-----------|--------|-------|
| CLI-First | [x] | N/A - no CLI changes |
| Git-Friendly | [x] | All changes are in plain text YAML/shell files |
| Protocol Agnostic | [x] | N/A - infrastructure only |
| Simplicity | [x] | Minimal changes: add branch name to existing triggers |
| Dogfooding | [x] | N/A - infrastructure only |
| Minimal Code | [x] | Adding single branch to existing arrays, no new logic |

## Affected Areas

```text
.github/workflows/
├── cli.yml                    # Add 'next' to branches
├── codecov.yml                # Add 'next' to branches
├── installer-tests.yml        # Add 'next' to branches
├── vscode-extension-build.yml # Add 'next' to branches
└── web-tests.yml              # Add 'next' to branches

cli/scripts/
└── bump.sh                    # Add 'next' to allowed branches

Makefile                       # Add 'next' to release target
```

## Implementation Approach

1. **Update all CI workflow files** to add `next` to branch triggers (both push and pull_request where applicable)
2. **Update `cli/scripts/bump.sh`** to allow releases from `next` branch
3. **Update `Makefile` release target** to allow releases from `next` branch
4. **Document manual steps** for Vercel and GitHub branch protection configuration

## Detailed Changes

### 1. GitHub Workflows (5 files)

Each file needs `next` added to the branches array:

**cli.yml** (lines 5-6, 10-11):
```yaml
on:
  push:
    branches: [ main, next ]
  pull_request:
    branches: [ main, next ]
```

**codecov.yml** (lines 5-6, 7-8):
```yaml
on:
  push:
    branches: [main, next]
  pull_request:
    branches: [main, next]
```

**installer-tests.yml** (lines 5-6, 9-10):
```yaml
on:
  push:
    branches: [main, next]
  pull_request:
    branches: [main, next]
```

**vscode-extension-build.yml** (lines 4-5, 6-7):
```yaml
on:
  push:
    branches: [ main, next ]
  pull_request:
    branches: [ main, next ]
```

**web-tests.yml** (lines 4-5, 10-11):
```yaml
on:
  push:
    branches: [ main, next ]
  pull_request:
    branches: [ main, next ]
```

### 2. cli/scripts/bump.sh (line 8)

Change:
```bash
if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "develop" ]]; then
```
To:
```bash
if [[ "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "develop" && "$CURRENT_BRANCH" != "next" ]]; then
```

Update error message on line 9:
```bash
echo "Error: Releases can only be made from 'main', 'develop', or 'next' branches"
```

### 3. Makefile (lines 105-106)

Change:
```makefile
if [ "$$BRANCH" != "main" ] && [ "$$BRANCH" != "develop" ]; then \
    echo "Error: Releases can only be made from 'main' or 'develop' branches"; \
```
To:
```makefile
if [ "$$BRANCH" != "main" ] && [ "$$BRANCH" != "develop" ] && [ "$$BRANCH" != "next" ]; then \
    echo "Error: Releases can only be made from 'main', 'develop', or 'next' branches"; \
```

## Manual Configuration Steps

After code changes are merged:

### Vercel Dashboard
1. Go to Project Settings > Git
2. Under "Production Branch", keep `main`
3. Under "Preview Deployments", ensure `next` branch deployments are enabled

### GitHub Repository Settings
1. Go to Settings > Branches
2. Add branch protection rule for `next`:
   - Require pull request reviews (optional)
   - Require status checks to pass
   - Select all CI checks that apply

## Verification

After implementation:
- [ ] Push to `next` triggers all CI workflows
- [ ] PR to `next` triggers all CI workflows
- [ ] `make release` works from `next` branch
- [ ] `bump.sh` works from `next` branch
- [ ] Vercel deploys preview for `next` branch
