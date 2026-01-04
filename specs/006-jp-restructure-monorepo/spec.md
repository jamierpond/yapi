# Feature Specification: Monorepo Restructure ("Polyglot Clear")

**Branch**: `006-jp-restructure-monorepo` | **Created**: 2026-01-03 | **Status**: Draft

## Overview

Reorganize the yapi polyglot monorepo (Go, TypeScript/Next.js, Lua) to improve discoverability, build caching, and separation of concerns by isolating the Go CLI core from consumers (web, VS Code, GitHub Action) and shared libraries.

## User Stories

### US1 - Developer Navigates Repository (P1)

A new developer joins the project and needs to understand the codebase structure. They can clearly identify where the core CLI code lives (`cli/`), where web applications are (`apps/`), where integrations exist (`integrations/`), and where shared TypeScript packages are (`packages/`).

**Acceptance**:
- Given a developer opens the repository root, when they look at top-level directories, then they can immediately understand the separation between CLI core, web apps, integrations, and shared packages
- Given a developer wants to modify the CLI, when they navigate to the `cli/` directory, then all Go source code is contained there

---

### US2 - Developer Builds CLI Independently (P1)

A developer needs to build only the Go CLI without triggering TypeScript builds. The CLI codebase is isolated in its own directory with its own `go.mod`.

**Acceptance**:
- Given a developer is in the `cli/` directory, when they run `make build`, then only Go code compiles
- Given a developer modifies CLI code, when they run tests, then only CLI-related tests execute

---

### US3 - Developer Works on Integration (P2)

A developer maintaining the VS Code extension, GitHub Action, or Neovim plugin can find all integration code in a single `integrations/` directory.

**Acceptance**:
- Given a developer wants to update the VS Code extension, when they navigate to `integrations/vscode/`, then all extension code is there
- Given a developer wants to update the GitHub Action, when they navigate to `integrations/github-action/`, then all action code is there

---

### US4 - Developer Uses Shared Package (P2)

A developer creating new TypeScript functionality can import shared packages from a flat `packages/` directory accessible to all apps and integrations.

**Acceptance**:
- Given `madea-blog-core` is hoisted to `packages/`, when the web app imports it, then the import resolves correctly
- Given the VS Code webview package is in `packages/vscode-ui/`, when the VS Code extension builds, then it can consume the webview

## Requirements

### Functional

- **FR-001**: The Go CLI source code (currently `cmd/`, `internal/`, `go.mod`, `go.sum`, `scripts/`, `bin/`) MUST be relocated to a `cli/` directory
- **FR-002**: The GitHub Action (currently `action/`) MUST be relocated to `integrations/github-action/`
- **FR-003**: The VS Code extension (currently `extensions/vscode-extension/`) MUST be relocated to `integrations/vscode/`
- **FR-004**: The Neovim plugin (currently `lua/`) MUST be relocated to `integrations/nvim/`
- **FR-005**: The vscode-webview app (currently `apps/vscode-webview/`) MUST be relocated to `packages/vscode-ui/`
- **FR-006**: The madea-blog-core package (currently `apps/web/madea-blog-core/`) MUST be hoisted to `packages/madea-blog-core/`
- **FR-007**: The `pnpm-workspace.yaml` MUST be updated to reflect new package locations (`apps/*`, `packages/*`, `integrations/*`)
- **FR-008**: All Go import paths within `cli/` MUST be updated to reflect the new module location
- **FR-009**: All TypeScript import paths for moved packages MUST be updated to reflect new locations
- **FR-010**: The root `Makefile` MUST be updated to reference `cli/` for Go build commands
- **FR-011**: Asset files (`pup.jpg`, `icon.svg`, `og-image.png`) MUST be consolidated into `packages/ui/assets/` or a dedicated `assets/` directory
- **FR-012**: AI context files (`CLAUDE.md`, `AGENTS.md`, `BRIEF.md`) MUST remain at root or be moved to a consistent location

### Directory Structure

```text
.
├── apps/                    # Deployable web applications
│   ├── web/                 # Main yapi website/dashboard
│   ├── docs/                # Documentation site (if separate)
│   └── playground/          # Interactive playground (if separate)
├── cli/                     # Go CLI Core Engine
│   ├── cmd/                 # Entry points
│   ├── internal/            # Private library code
│   ├── scripts/             # Build scripts
│   ├── bin/                 # Build output
│   ├── go.mod               # Go module definition
│   └── go.sum               # Go dependency lock
├── integrations/            # Editor/CI integrations
│   ├── vscode/              # VS Code extension
│   ├── github-action/       # GitHub Action
│   └── nvim/                # Neovim plugin
├── packages/                # Shared TypeScript/JS libraries
│   ├── client/              # API client
│   ├── ui/                  # Shared UI components
│   ├── styles/              # Shared styles
│   ├── vscode-ui/           # VS Code webview UI
│   └── madea-blog-core/     # Blog functionality
├── examples/                # Example .yapi.yml files
├── specs/                   # Design docs and RFCs
├── .specify/                # Speckit tooling
├── pnpm-workspace.yaml      # Updated package paths
├── go.work                  # Multi-module Go workspace (optional)
└── README.md
```

## Edge Cases

- What happens if a package has circular dependencies after relocation?
  - The relocation preserves existing dependency relationships; circular dependencies would have existed before
- What happens if TypeScript package names in `package.json` conflict with new paths?
  - Package names in `package.json` are independent of file paths; only `pnpm-workspace.yaml` needs updating
- What happens if the `madea-blog-core` submodule has its own git history?
  - It appears to be a git submodule; it should be moved as-is, preserving the submodule relationship or converted to a regular directory

## Assumptions

- The `madea-blog-core` directory containing a `.git` folder is a git submodule and will be handled appropriately during the move
- The `docs/` and `playground/` apps mentioned in the proposed structure do not currently exist and are optional future additions
- The `go.work` file is optional and only needed if multiple Go modules exist
- Existing CI/CD pipelines will need updates (out of scope for this restructure but noted)
- The `imported/` and `postman-collections/` directories at root can remain or be moved to `examples/` (decision deferred)

## Success Criteria

- All Go code is contained within the `cli/` directory
- All integrations (VS Code, GitHub Action, Neovim) are contained within `integrations/`
- All shared TypeScript packages are in a flat `packages/` directory
- `make build` successfully compiles the CLI from the new location
- `make test` passes for the CLI
- `make lint` passes for the CLI
- `pnpm install` resolves all dependencies correctly
- Each relocated package builds and runs correctly
- No broken imports exist in TypeScript or Go code
- Root directory contains only orchestration files (README, config, workspace files) and top-level organizational directories
