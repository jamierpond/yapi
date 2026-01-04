# Research: Monorepo Restructure

**Feature**: 006-jp-restructure-monorepo
**Date**: 2026-01-03

## Research Topics

### 1. Go Module Path Handling After Directory Move

**Question**: Does moving Go code to a subdirectory require changing the module path in `go.mod`?

**Decision**: Keep module path as `yapi.run/cli`

**Rationale**:
- The module path in `go.mod` is independent of the directory structure
- Moving from `/go.mod` to `/cli/go.mod` does not require changing `module yapi.run/cli`
- All internal imports use the module path, not filesystem paths
- External consumers (if any) would reference `yapi.run/cli` regardless of directory

**Alternatives Considered**:
- Change to `yapi.run/yapi/cli`: Rejected - unnecessary change, breaks any existing references
- Keep go.mod at root: Rejected - defeats the purpose of isolating Go code

### 2. Git Submodule Path Update

**Question**: How to properly move a git submodule to a new location?

**Decision**: Update `.gitmodules` file with new path

**Rationale**:
- Git submodules are tracked in `.gitmodules` which maps paths to URLs
- Changing the path in `.gitmodules` is sufficient; URL stays the same
- The submodule content moves with a standard directory move
- No need to remove and re-add the submodule

**Implementation**:
```gitmodules
[submodule "packages/madea-blog-core"]
    path = packages/madea-blog-core
    url = https://github.com/jamierpond/madea-blog-core.git
```

**Alternatives Considered**:
- Remove and re-add submodule: Rejected - unnecessary complexity
- Convert to regular directory: Rejected - preserves git history benefits

### 3. pnpm Workspace Path Patterns

**Question**: What path patterns should `pnpm-workspace.yaml` include?

**Decision**: Use `apps/*`, `packages/*`, `integrations/*`

**Rationale**:
- pnpm workspace uses glob patterns to find packages
- Each directory containing a `package.json` becomes a workspace package
- The GitHub Action in `integrations/github-action/` has its own `package.json`
- VS Code extension in `integrations/vscode/` has its own `package.json`

**Updated Configuration**:
```yaml
packages:
  - 'apps/*'
  - 'packages/*'
  - 'integrations/*'
```

**Alternatives Considered**:
- Keep `extensions/*`: Rejected - directory no longer exists
- List each package explicitly: Rejected - glob patterns are cleaner

### 4. VS Code Extension Webview Path

**Question**: How should the VS Code extension reference the webview after it moves to packages?

**Decision**: Update `copy:webview` script to use new relative path

**Current Path**: `../../apps/vscode-webview/dist/`
**New Path**: `../../packages/vscode-ui/dist/`

**Rationale**:
- The VS Code extension copies built webview assets to its `media/` folder
- Path is relative from `integrations/vscode/` to `packages/vscode-ui/`
- Two levels up (`../../`) then into `packages/vscode-ui/dist/`

**Alternatives Considered**:
- Symlink: Rejected - doesn't work well with VS Code extension packaging
- Absolute paths: Rejected - not portable

### 5. Asset File Consolidation

**Question**: Where should root-level assets (`pup.jpg`, `icon.svg`, `og-image.png`) go?

**Decision**: Create `assets/` directory at root

**Rationale**:
- These are marketing/branding assets, not UI component assets
- `packages/ui/` is for React components, not static marketing images
- A dedicated `assets/` folder is clearer and matches common monorepo patterns
- The web app can reference them from there

**Alternatives Considered**:
- `packages/ui/assets/`: Rejected - conflates component library with branding
- `apps/web/public/`: Rejected - some assets are used by other apps too
- Leave at root: Rejected - adds to root noise (the problem we're solving)

### 6. Makefile Updates for CLI Subdirectory

**Question**: How should the root Makefile invoke Go builds in the `cli/` subdirectory?

**Decision**: Use `cd cli &&` prefix or `-C` flag for Go commands

**Rationale**:
- `go build` needs to run from the module root (where go.mod lives)
- Using `cd cli && go build ./cmd/yapi` is clear and portable
- Alternatively, `make -C cli` could invoke a cli-local Makefile

**Implementation Approach**:
```makefile
build:
    cd cli && go build -ldflags "$(LDFLAGS)" -o ./bin/yapi ./cmd/yapi
```

**Alternatives Considered**:
- Nested Makefile in cli/: Could add later if cli/ grows complex
- Go workspace (go.work): Not needed for single module

## Summary

All research questions resolved. No blockers identified. Implementation can proceed with the documented decisions.
