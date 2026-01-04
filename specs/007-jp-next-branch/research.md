# Research: Next Branch Workflow

## Current State Analysis

### GitHub Workflows

| Workflow | Current Branches | Trigger Type | Needs Update |
|----------|------------------|--------------|--------------|
| cli.yml | `main` only | push/PR | Yes - add `next` |
| codecov.yml | `main` only | push/PR | Yes - add `next` |
| github-action-dist.yml | any (path-based) | push/PR | No |
| installer-tests.yml | `main` only | push/PR | Yes - add `next` |
| release.yaml | tags only (`v*`) | push | No |
| vscode-extension-build.yml | `main` only | push/PR | Yes - add `next` |
| vscode-extension-publish.yml | tags only (`v*`) | push | No |
| web-tests.yml | `main` only | push/PR | Yes - add `next` |

### Release Process

**Current release flow** (from `cli/scripts/bump.sh` and `Makefile`):
1. `bump.sh` only allows releases from `main` or `develop`
2. Tags are created with format `v*.*.*`
3. Release workflow triggers on tag push
4. GoReleaser handles distribution
5. VS Code extension publishes on tag push

**Decision**: Add `next` to bump.sh allowed branches

### Vercel Configuration

**Current** (`vercel.json`):
- Framework: Next.js
- Build command: `bash ./cli/scripts/vercel-build.sh`
- Output: `./apps/web/.next`

Vercel branch deployments are configured via Vercel dashboard, not in `vercel.json`. The `next` branch will need to be configured as a preview branch in Vercel settings.

**Decision**: Document Vercel dashboard configuration (out of code scope)

### Makefile Release Target

**Current** (line 104-115):
```makefile
release:
    @BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
    if [ "$$BRANCH" != "main" ] && [ "$$BRANCH" != "develop" ]; then
        ...
```

**Decision**: Update to allow `next` branch

## Decisions

### 1. Workflow Branch Configuration

**Decision**: Add `next` to branch triggers alongside `main`

**Rationale**:
- CI should run on `next` to validate integrations before promotion to main
- All existing workflows that run on `main` should also run on `next`
- Tag-based workflows (release, vscode-publish) remain unchanged since releases only happen from main

**Alternatives considered**:
- Separate workflows for `next`: Rejected - would duplicate logic and violate Minimal Code principle
- Only run subset of tests: Rejected - `next` should have same validation as main

### 2. bump.sh Branch Restrictions

**Decision**: Update `bump.sh` to allow releases from `main`, `develop`, and `next`

**Rationale**:
- Allows flexibility if release workflow changes
- `next` is a promotion point that may need to create releases

**Note**: The Makefile `release` target has the same restriction and should also be updated.

### 3. Vercel Preview Deployments

**Decision**: Configure `next` as a preview branch in Vercel dashboard

**Rationale**:
- Automatic deployments for staging validation
- No code changes required - purely configuration
- Preview URL provides testing endpoint

**Implementation**: Manual configuration in Vercel dashboard (document in plan)

## Files to Modify

| File | Change |
|------|--------|
| `.github/workflows/cli.yml` | Add `next` to branches |
| `.github/workflows/codecov.yml` | Add `next` to branches |
| `.github/workflows/installer-tests.yml` | Add `next` to branches |
| `.github/workflows/vscode-extension-build.yml` | Add `next` to branches |
| `.github/workflows/web-tests.yml` | Add `next` to branches |
| `cli/scripts/bump.sh` | Add `next` to allowed branches |
| `Makefile` | Add `next` to release target |

## Out of Scope

- Vercel dashboard configuration (manual step)
- GitHub branch protection rules (manual step)
- Documentation updates (separate task)
