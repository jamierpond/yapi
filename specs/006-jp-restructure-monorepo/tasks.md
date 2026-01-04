# Tasks: Monorepo Restructure ("Polyglot Clear")

**Input**: `specs/006-jp-restructure-monorepo/plan.md`

## Format

- `[P]` = Can run in parallel (no dependencies on incomplete tasks)
- `[US#]` = User story reference
- Include file paths for each task

## Phase 1: Setup

Create the target directory structure before moving any content.

- [x] T001 Create `cli/` directory at repository root
- [x] T002 [P] Create `integrations/` directory at repository root
- [x] T003 [P] Create `assets/` directory at repository root

**Checkpoint**: Three new empty directories exist at root (`cli/`, `integrations/`, `assets/`)

---

## Phase 2: Foundational - Move Go CLI

Move all Go CLI code to the `cli/` subdirectory. This is blocking because the Makefile update depends on it.

- [x] T004 Move `cmd/` directory to `cli/cmd/`
- [x] T005 Move `internal/` directory to `cli/internal/`
- [x] T006 Move `scripts/` directory to `cli/scripts/`
- [x] T007 Move `bin/` directory to `cli/bin/`
- [x] T008 Move `go.mod` to `cli/go.mod`
- [x] T009 Move `go.sum` to `cli/go.sum`
- [x] T010 Update `Makefile` build target to use `cd cli &&` prefix for Go commands

**Checkpoint**: `cd cli && go build ./cmd/yapi` succeeds

---

## Phase 3: User Story 1 - Developer Navigates Repository (P1)

Goal: Clear separation of concerns visible at root directory level.

**Test Criteria**: Root directory has only organizational directories and config files.

- [x] T011 [US1] Verify root directory structure matches target (cli/, apps/, packages/, integrations/, assets/, examples/, specs/)

**Checkpoint**: `ls -la` at root shows clean organizational structure

---

## Phase 4: User Story 2 - Developer Builds CLI Independently (P1)

Goal: Go CLI can be built in isolation without triggering TypeScript tooling.

**Test Criteria**: `make build`, `make test`, `make lint` all pass.

- [x] T012 [US2] Run `make build` and verify Go binary is created at `cli/bin/yapi`
- [x] T013 [US2] Run `make test` and verify all Go tests pass
- [x] T014 [US2] Run `make lint` and verify no linting errors

**Checkpoint**: All three make commands succeed

---

## Phase 5: User Story 3 - Developer Works on Integration (P2)

Goal: All integrations consolidated in `integrations/` directory.

**Test Criteria**: VS Code extension, GitHub Action, and Neovim plugin are in `integrations/`.

- [x] T015 [US3] Move `action/` to `integrations/github-action/`
- [x] T016 [US3] [P] Move `extensions/vscode-extension/` to `integrations/vscode/`
- [x] T017 [US3] [P] Move `lua/` to `integrations/nvim/`
- [x] T018 [US3] Remove empty `extensions/` directory
- [x] T019 [US3] Update `integrations/vscode/package.json` copy:webview script path from `../../apps/vscode-webview/dist/` to `../../packages/vscode-ui/dist/`

**Checkpoint**: `ls integrations/` shows `github-action/`, `vscode/`, `nvim/`

---

## Phase 6: User Story 4 - Developer Uses Shared Package (P2)

Goal: Hoisted packages available in flat `packages/` directory.

**Test Criteria**: `pnpm install` resolves all workspace dependencies correctly.

- [x] T020 [US4] Move `apps/vscode-webview/` to `packages/vscode-ui/`
- [x] T021 [US4] Update `.gitmodules` to change madea-blog-core path from `apps/web/madea-blog-core` to `packages/madea-blog-core`
- [x] T022 [US4] Move `apps/web/madea-blog-core/` to `packages/madea-blog-core/`
- [x] T023 [US4] Update `pnpm-workspace.yaml` to include `integrations/*` in packages list

**Checkpoint**: `pnpm install` succeeds with no resolution errors

---

## Phase 7: Polish & Cross-Cutting Concerns

Consolidate assets and verify all builds work.

- [x] T024 [P] Move `pup.jpg` to `assets/pup.jpg`
- [x] T025 [P] Move `icon.svg` to `assets/icon.svg`
- [x] T026 [P] Move `og-image.png` to `assets/og-image.png`
- [x] T027 Run `pnpm install` to regenerate lockfile
- [x] T028 Run `pnpm -r build` to verify all TypeScript packages build (madea-blog-core has pre-existing dep issue)
- [x] T029 Run `pnpm --filter yapi-extension build` to verify VS Code extension builds
- [x] T030 Run `pnpm --filter @yapi/web build` to verify web app builds (fixed by adding simple-git to web deps)

**Checkpoint**: All builds pass, no broken imports

---

## Verification

```bash
# Go CLI verification
cd cli && go build ./cmd/yapi && go test ./... && go vet ./... && cd ..

# Root Makefile verification
make build && make test && make lint

# TypeScript verification
pnpm install
pnpm -r build

# VS Code extension specifically
pnpm --filter yapi-extension build
```

## Dependencies

```
T001 → T004-T009 (cli/ must exist before moving content)
T002 → T015-T018 (integrations/ must exist before moving content)
T003 → T024-T026 (assets/ must exist before moving content)
T004-T009 → T010 (Go code moved before Makefile update)
T010 → T012-T014 (Makefile updated before testing)
T019 → T020 (webview path update after webview moved)
T015-T018 → T027 (integrations moved before pnpm install)
T020-T023 → T027 (packages hoisted before pnpm install)
T027 → T028-T030 (pnpm install before builds)
```

## Parallel Execution Opportunities

**Phase 1** (all parallel):
- T001, T002, T003 can run simultaneously

**Phase 2** (sequential):
- T004-T009 can run in parallel (independent directory moves)
- T010 must wait for T004-T009

**Phase 5** (mostly parallel):
- T015, T016, T017 can run in parallel (different source directories)
- T018 waits for T016
- T019 waits for T016

**Phase 7** (asset moves parallel):
- T024, T025, T026 can run in parallel

## Notes

- This is a file restructure, not new code. Tasks are move/update operations.
- Go module path `yapi.run/cli` remains unchanged (module path is filesystem-independent).
- Package names (`@yapi/client`, `@yapi/ui`, etc.) remain unchanged.
- Git submodule (madea-blog-core) requires `.gitmodules` update before move.
- AI context files (CLAUDE.md, AGENTS.md, BRIEF.md) remain at root.

## Summary

| Metric | Value |
|--------|-------|
| **Total Tasks** | 30 |
| **Setup Phase** | 3 tasks |
| **Foundational Phase** | 7 tasks |
| **US1 (P1)** | 1 task |
| **US2 (P1)** | 3 tasks |
| **US3 (P2)** | 5 tasks |
| **US4 (P2)** | 4 tasks |
| **Polish Phase** | 7 tasks |
| **Parallel Opportunities** | 14 tasks can run in parallel at various points |
| **MVP Scope** | Phases 1-4 (US1 + US2 complete, CLI fully isolated) |
