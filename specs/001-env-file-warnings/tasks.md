# Tasks: Env File Warnings

**Input**: `specs/001-env-file-warnings/plan.md`

## Format

- `[P]` = Can run in parallel
- `[US#]` = User story reference
- Include file paths: `internal/config/loader.go`

---

## Phase 1: Setup & Data Structures

- [x] T001 [P] Add `EnvFileStatus` struct in `internal/config/loader.go`
- [x] T002 [P] Add `EnvFileLoadResult` struct in `internal/config/loader.go`
- [x] T003 Add `--strict-env` flag definition to run command in `internal/cli/commands/commands.go`

**Checkpoint**: Structs compile, flag is recognized by CLI

---

## Phase 2: Foundational - Variable Resolution Priority

These tasks must complete before user story implementation.

- [x] T004 Modify `buildEnvFileResolver()` in `internal/config/loader.go` to prioritize env_files over OS env
- [x] T005 Add resolution source tracking to resolver in `internal/config/loader.go` (track where each var came from)
- [x] T006 Modify `loadEnvFiles()` signature in `internal/config/loader.go` to return `(map[string]string, []string, error)` for vars, warnings, error
- [x] T007 Update all callers of `loadEnvFiles()` in `internal/config/loader.go` to handle new return signature
- [x] T008 Update `ResolveEnvFiles()` in `internal/config/project.go` to use new `loadEnvFiles()` signature

**Checkpoint**: `make build` passes, existing tests still pass

---

## Phase 3: User Story 1 - Missing Env File Warning (P1)

**Goal**: When env file is missing, show warning and continue execution.

- [x] T009 [US1] Implement file existence check in `loadEnvFiles()` in `internal/config/loader.go`
- [x] T010 [US1] On missing file: add warning to result, continue to next file in `internal/config/loader.go`
- [x] T011 [US1] Output warnings to stderr in CLI execution path in `cmd/yapi/main.go`
- [x] T012 [US1] Add unit test for missing env file warning in `internal/config/loader_test.go`

**Checkpoint**: `yapi run test.yapi.yml` with missing .env shows warning, continues execution

---

## Phase 4: User Story 2 - Strict Mode (P1)

**Goal**: `--strict-env` treats missing files as errors and disables OS fallback.

- [x] T013 [US2] Wire `--strict-env` flag through to resolver in `cmd/yapi/main.go`
- [x] T014 [US2] In strict mode: convert missing file warnings to errors in `internal/config/loader.go`
- [x] T015 [US2] In strict mode: skip OS env fallback in `buildEnvFileResolver()` in `internal/config/loader.go`
- [x] T016 [US2] In strict mode: error on undefined variables in `internal/config/loader.go`
- [x] T017 [US2] Add integration test for `--strict-env` flag in `cmd/yapi/main_test.go`

**Checkpoint**: `yapi run test.yapi.yml --strict-env` with missing .env exits with error

---

## Phase 5: User Story 3 & 4 - Multiple Files & Permission Errors (P2)

**Goal**: Handle multiple missing files with individual warnings; permission errors always halt.

- [x] T018 [US3] Ensure each missing file gets separate warning in `internal/config/loader.go`
- [x] T019 [US3] Deduplicate warnings for duplicate env_files entries in `internal/config/loader.go`
- [x] T020 [US4] Check file readability after existence check in `internal/config/loader.go`
- [x] T021 [US4] On permission error: return error immediately (not warning) in `internal/config/loader.go`
- [x] T022 [P] [US3] Add unit test for multiple missing files in `internal/config/loader_test.go`
- [x] T023 [P] [US4] Add unit test for permission error handling in `internal/config/loader_test.go`

**Checkpoint**: Multiple missing files show multiple warnings; permission error halts execution

---

## Phase 6: User Story 5 - LSP Go-to-Definition for Env Files (P1)

**Goal**: Ctrl+click on env file path navigates to that file at line 1.

- [ ] T024 [US5] Add `findEnvFilePathAtPosition()` helper in `internal/langserver/langserver.go`
- [ ] T025 [US5] Parse YAML to find env_files array node positions in `internal/langserver/langserver.go`
- [ ] T026 [US5] Extend `textDocumentDefinition()` to check for env file paths in `internal/langserver/langserver.go`
- [ ] T027 [US5] Return `protocol.Location` with line 1 of target file in `internal/langserver/langserver.go`
- [ ] T028 [US5] Handle missing file case: return nil (no navigation) in `internal/langserver/langserver.go`
- [ ] T029 [US5] Add unit test for go-to-definition on env file path in `internal/langserver/langserver_test.go`

**Checkpoint**: In VS Code, Ctrl+click on `.env.local` in env_files opens that file

---

## Phase 7: User Story 6 - LSP Go-to-Definition for Variables (P1)

**Goal**: Ctrl+click on `${VAR}` navigates to variable definition in env file.

- [ ] T030 [US6] Verify existing `textDocumentDefinition()` handles `${VAR}` references in `internal/langserver/langserver.go`
- [ ] T031 [US6] Ensure navigation goes to correct line in env file via `findVarPositionInEnvFile()` in `internal/langserver/langserver.go`
- [ ] T032 [US6] Handle variable defined in multiple files: use first per env_files order in `internal/langserver/langserver.go`
- [ ] T033 [US6] Add unit test for go-to-definition on variable reference in `internal/langserver/langserver_test.go`

**Checkpoint**: In VS Code, Ctrl+click on `${GITHUB_PAT}` opens .env file at definition line

---

## Phase 8: LSP Diagnostics (FR-011, FR-012, FR-017, FR-018)

**Goal**: Show warnings/info in editor for missing files, undefined vars, OS fallback, resolution source.

- [ ] T034 [P] Add `validateEnvFilesExist()` function in `internal/validation/analyzer.go`
- [ ] T035 [P] Add `validateVariablesDefined()` function in `internal/validation/analyzer.go`
- [ ] T036 Call new validation functions from `AnalyzeConfigString()` in `internal/validation/analyzer.go`
- [ ] T037 Call new validation functions from `AnalyzeConfigStringWithProject()` in `internal/validation/analyzer.go`
- [ ] T038 Generate warning diagnostic for missing env files with line/col in `internal/validation/analyzer.go`
- [ ] T039 Generate warning diagnostic for undefined variables in `internal/validation/analyzer.go`
- [ ] T040 Generate warning diagnostic for OS env fallback in `internal/validation/analyzer.go`
- [ ] T041 Generate info diagnostic showing resolution source for each variable in `internal/validation/analyzer.go`
- [ ] T042 [P] Add unit test for missing env file diagnostic in `internal/validation/analyzer_test.go`
- [ ] T043 [P] Add unit test for undefined variable diagnostic in `internal/validation/analyzer_test.go`
- [ ] T044 [P] Add unit test for OS fallback warning in `internal/validation/analyzer_test.go`
- [ ] T045 [P] Add unit test for resolution source info in `internal/validation/analyzer_test.go`

**Checkpoint**: Open YAPI file with missing .env in VS Code, see squiggly underlines

---

## Phase 9: Polish & Cross-Cutting

- [ ] T046 Add example YAPI file with env_files in `examples/env-files.yapi.yml`
- [ ] T047 Verify `make test` passes with all new tests
- [ ] T048 Verify `make lint` passes
- [ ] T049 Manual test: CLI warning output format
- [ ] T050 Manual test: LSP diagnostics in VS Code

---

## Verification

```bash
make build && make test && make lint

# CLI tests
yapi run examples/env-files.yapi.yml  # Should show warnings
yapi run examples/env-files.yapi.yml --strict-env  # Should error

# LSP tests (in VS Code)
# 1. Open file with missing .env - see warning underline
# 2. Ctrl+click on env file path - navigates to file
# 3. Ctrl+click on ${VAR} - navigates to definition
```

---

## Dependencies

```
T001, T002, T003 (parallel)
    └── T004, T005, T006, T007, T008 (sequential - foundational)
        ├── T009-T012 (US1 - missing file warning)
        │   └── T013-T017 (US2 - strict mode, depends on US1)
        │       └── T018-T023 (US3/US4 - multiple files, permission)
        ├── T024-T029 (US5 - LSP go-to-def env files, parallel with US1)
        ├── T030-T033 (US6 - LSP go-to-def vars, parallel with US1)
        └── T034-T045 (LSP diagnostics, parallel with US5/US6)
            └── T046-T050 (polish, after all)
```

## Parallel Execution Opportunities

**Phase 1**: T001, T002, T003 can run in parallel
**Phase 5**: T022, T023 can run in parallel
**Phase 8**: T034, T035 can run in parallel; T042-T045 can run in parallel
**Cross-phase**: US5, US6, and LSP Diagnostics (Phases 6-8) can run in parallel after Phase 2

## Summary

- **Total tasks**: 50
- **US1 (Missing file warning)**: 4 tasks
- **US2 (Strict mode)**: 5 tasks
- **US3/US4 (Multiple files, permissions)**: 6 tasks
- **US5 (LSP go-to-def env files)**: 6 tasks
- **US6 (LSP go-to-def vars)**: 4 tasks
- **LSP Diagnostics**: 12 tasks
- **Setup/Foundational**: 8 tasks
- **Polish**: 5 tasks

## MVP Scope

**Recommended MVP**: Complete Phases 1-4 (US1 + US2)
- CLI shows warnings for missing env files
- `--strict-env` flag works
- 17 tasks total for MVP
