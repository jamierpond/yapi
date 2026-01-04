# Implementation Plan: Monorepo Restructure ("Polyglot Clear")

**Branch**: `006-jp-restructure-monorepo` | **Date**: 2026-01-03 | **Spec**: [spec.md](./spec.md)

## Summary

Reorganize the yapi polyglot monorepo by isolating the Go CLI into `cli/`, consolidating integrations into `integrations/`, and hoisting nested packages to a flat `packages/` structure. This is primarily a file/directory restructure with accompanying configuration updates.

## Technical Context

**Languages**: Go 1.25+, TypeScript 5.x, Lua
**Package Manager**: pnpm (workspace)
**Go Module**: `yapi.run/cli` (will change to `yapi.run/cli` from `cli/` subdirectory)
**Build Tools**: Make (Go), pnpm/vite/webpack (TypeScript)
**Testing**: `make test` (Go), pnpm test (TypeScript)

### Current Structure Analysis

| Component | Current Location | Target Location | Package Name |
|-----------|------------------|-----------------|--------------|
| Go CLI | `cmd/`, `internal/`, `go.mod` | `cli/` | `yapi.run/cli` |
| GitHub Action | `action/` | `integrations/github-action/` | `yapi-action` |
| VS Code Extension | `extensions/vscode-extension/` | `integrations/vscode/` | `yapi-extension` |
| VS Code Webview | `apps/vscode-webview/` | `packages/vscode-ui/` | `@yapi/vscode-webview` |
| Neovim Plugin | `lua/` | `integrations/nvim/` | N/A (Lua) |
| madea-blog-core | `apps/web/madea-blog-core/` | `packages/madea-blog-core/` | git submodule |
| Client | `packages/client/` | stays | `@yapi/client` |
| UI | `packages/ui/` | stays | `@yapi/ui` |
| Styles | `packages/styles/` | stays | `@yapi/styles` |
| Web | `apps/web/` | stays | N/A |

### Key Dependencies to Update

1. **VS Code Extension → Webview**: `copy:webview` script references `../../apps/vscode-webview/dist/`
2. **Webview → UI/Styles**: workspace dependencies `@yapi/ui`, `@yapi/styles`
3. **pnpm-workspace.yaml**: Currently references `apps/*`, `packages/*`, `extensions/*`
4. **Makefile**: References `cmd/yapi`, `internal/`

## Constitution Check

*Must pass before implementation.*

| Principle | Status | Notes |
|-----------|--------|-------|
| CLI-First | [x] Pass | Not affected - restructure only |
| Git-Friendly | [x] Pass | YAML configs unchanged |
| Protocol Agnostic | [x] Pass | Not affected |
| Simplicity | [x] Pass | Restructure reduces root noise, improves discoverability |
| Dogfooding | [x] Pass | Web app unaffected |
| Minimal Code | [x] Pass | No new code; just moving files and updating paths |
| Single Code Path | [x] Pass | Not affected |

## Affected Areas

```text
./                          # Root: Makefile, pnpm-workspace.yaml, .gitmodules
├── cli/                    # NEW: Go CLI moved here
│   ├── cmd/
│   ├── internal/
│   ├── scripts/
│   ├── bin/
│   ├── go.mod
│   └── go.sum
├── integrations/           # NEW: All integrations grouped
│   ├── vscode/             # Moved from extensions/vscode-extension
│   ├── github-action/      # Moved from action/
│   └── nvim/               # Moved from lua/
├── packages/               # UPDATED: Hoisted packages added
│   ├── vscode-ui/          # Moved from apps/vscode-webview
│   └── madea-blog-core/    # Moved from apps/web/madea-blog-core
└── apps/                   # UPDATED: Only web remains
    └── web/
```

## Implementation Approach

### Phase 1: Create Directory Structure
- Create `cli/` and `integrations/` directories
- This establishes the target structure before moving content

### Phase 2: Move Go CLI
- Move `cmd/`, `internal/`, `scripts/`, `bin/`, `go.mod`, `go.sum` to `cli/`
- Update `Makefile` to build from `cli/` directory
- The Go module path `yapi.run/cli` should remain unchanged (module path is independent of directory)

### Phase 3: Move Integrations
- Move `action/` → `integrations/github-action/`
- Move `extensions/vscode-extension/` → `integrations/vscode/`
- Move `lua/` → `integrations/nvim/`
- Update VS Code extension's `copy:webview` script path

### Phase 4: Hoist Packages
- Move `apps/vscode-webview/` → `packages/vscode-ui/`
- Update `.gitmodules` for `madea-blog-core` new path
- Move `apps/web/madea-blog-core/` → `packages/madea-blog-core/`

### Phase 5: Update Configuration
- Update `pnpm-workspace.yaml` to include `integrations/*`
- Update all relative path references in:
  - VS Code extension `package.json` (copy:webview script)
  - Any tsconfig.json path mappings
- Run `pnpm install` to regenerate lockfile

### Phase 6: Consolidate Assets
- Move `pup.jpg`, `icon.svg`, `og-image.png` to `packages/ui/assets/` or `assets/`
- Update any references to these files

### Phase 7: Validation
- Run `make build` from `cli/` (via updated root Makefile)
- Run `make test` and `make lint`
- Run `pnpm install` and verify workspace resolution
- Build VS Code extension
- Build web app

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Broken Go imports | Go module path stays `yapi.run/cli`; internal imports use relative paths |
| Broken TS imports | Package names (`@yapi/*`) unchanged; only workspace paths in pnpm update |
| Submodule issues | Update `.gitmodules` path; submodule URL unchanged |
| CI/CD breaks | Out of scope but noted; CI files may need path updates |

## Files to Update

1. **Makefile** - Change `go build ./cmd/yapi` to `cd cli && go build ./cmd/yapi`
2. **pnpm-workspace.yaml** - Add `integrations/*` to packages list
3. **.gitmodules** - Update `madea-blog-core` path
4. **extensions/vscode-extension/package.json** → **integrations/vscode/package.json** - Update `copy:webview` path
5. **Root config cleanup** - Remove empty `extensions/` directory after move

## Complexity Justification

> No concerns - this is a straightforward restructure with clear benefits.

| Concern | Why Needed | Simpler Alternative Rejected |
|---------|------------|------------------------------|
| N/A | N/A | N/A |
