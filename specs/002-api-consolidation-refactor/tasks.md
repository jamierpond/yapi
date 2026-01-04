# Tasks: API Consolidation Refactor

**Input**: `specs/002-api-consolidation-refactor/plan.md`

## Format

- `[P]` = Can run in parallel
- `[US#]` = User story reference
- Include file paths: `internal/executor/http.go`

---

## Phase 1: Foundational - AST Extraction

**Goal**: Extract shared AST helpers to enable accurate line numbers (blocks US2)

- [x] T001 Create `Location` type in `internal/validation/ast.go`
- [x] T002 Move `findVarPositionInYAML` from `internal/langserver/langserver.go:787` to `internal/validation/ast.go` as `FindVarPositionInYAML`
- [x] T003 Move `findNodeInMapping` from `internal/langserver/langserver.go:906` to `internal/validation/ast.go` as `FindNodeInMapping`
- [x] T004 Move `findKeyNodeInMapping` from `internal/langserver/langserver.go:922` to `internal/validation/ast.go` as `FindKeyNodeInMapping`
- [x] T005 Update `internal/langserver/langserver.go` to import and call `validation.FindVarPositionInYAML()`
- [x] T006 Run `make test && make lint` to verify no regressions

**Checkpoint**: AST helpers extracted, langserver still works

---

## Phase 2: User Story 1 (P1) - Analyzer Consolidation

**Goal**: Single `Analyze()` entry point replaces 4 functions

- [x] T007 [US1] Add `AnalyzeOptions` struct in `internal/validation/analyzer.go`
- [x] T008 [US1] Create `Analyze(text string, opts AnalyzeOptions) (*Analysis, error)` in `internal/validation/analyzer.go`
- [x] T009 [US1] Refactor `analyzeParsed` to accept `AnalyzeOptions` in `internal/validation/analyzer.go`
- [x] T010 [US1] Convert `AnalyzeConfigString` to wrapper calling `Analyze()` in `internal/validation/analyzer.go`
- [x] T011 [US1] Convert `AnalyzeConfigStringWithProject` to wrapper in `internal/validation/analyzer.go`
- [x] T012 [US1] Convert `AnalyzeConfigStringWithProjectAndPath` to wrapper in `internal/validation/analyzer.go`
- [x] T013 [US1] Convert `AnalyzeConfigStringWithProjectAndPathAndOptions` to wrapper in `internal/validation/analyzer.go`
- [x] T014 [US1] Add deprecation comments to old functions in `internal/validation/analyzer.go`
- [x] T015 [US1] Run `make test && make lint`

**Checkpoint**: `yapi validate` works, all tests pass

---

## Phase 3: User Story 2 (P2) - Accurate Line Numbers

**Goal**: Diagnostics report real line numbers, not line 0

- [x] T016 [US2] Update `ValidateEnvFilesExistFromProject` to use `FindVarPositionInYAML` in `internal/validation/analyzer.go` (already uses FindEnvFilesInConfig for config-level)
- [x] T017 [US2] Update `ValidateProjectVars` to use AST helpers for line numbers in `internal/validation/analyzer.go` (already uses findVarLine)
- [x] T018 [US2] Run `make test && make lint`

**Checkpoint**: `yapi validate` shows accurate line numbers for env file errors

---

## Phase 4: User Story 3 (P2) - Executor Simplification

**Goal**: Replace Factory with standalone `GetTransport()`

- [x] T019 [P] [US3] Add `GetTransport(transport string, client HTTPClient) (TransportFunc, error)` in `internal/executor/executor.go`
- [x] T020 [US3] Move switch logic from `Factory.Create()` to `GetTransport()` in `internal/executor/executor.go`
- [x] T021 [US3] Add deprecation comments to `Factory` and `NewFactory` in `internal/executor/executor.go`
- [x] T022 [US3] Update callers in `cmd/yapi/main.go` to use `GetTransport()` directly (N/A - main.go uses core.Engine, Factory now delegates to GetTransport)
- [x] T023 [US3] Run `make test && make lint`

**Checkpoint**: `yapi run` works with all transport types

---

## Phase 5: ChainContext as Resolver

**Goal**: Deduplicate variable expansion logic

- [x] T024 Add `Resolve(key string) (string, error)` method to `ChainContext` in `internal/runner/context.go`
- [x] T025 Refactor `ExpandVariables` to delegate to `vars.ExpandString(input, c.Resolve)` in `internal/runner/context.go`
- [x] T026 Remove duplicate regex handling from `ExpandVariables` in `internal/runner/context.go` (removed dead `$key` branch)
- [x] T027 Run `make test && make lint`

**Checkpoint**: Variable expansion works, request chaining still functions

---

## Phase 6: Main.go Extraction

**Goal**: Reduce main.go by ~400 lines

- [x] T028 [P] Create `internal/output/` directory (already existed)
- [x] T029 [P] Create `internal/output/result.go` with `JSONOutput` struct and `PrintJSON()` function
- [x] T030 Move `printResultAsJSON` logic from `cmd/yapi/main.go` to `internal/output/result.go`
- [x] T031 Update `cmd/yapi/main.go` to call `output.PrintJSON()`
- [ ] T032-T037 DEFERRED: stress/import handlers are tightly coupled to cobra commands (174 lines reduced; stress/import extraction would require major refactoring)
- [x] T038 Run `make test && make lint`

**Checkpoint**: main.go reduced, all commands still work

---

## Phase 7: Verification & Cleanup

- [x] T039 Count LOC: main.go reduced by 174 lines (2244 → 2070); ast.go added 117 lines shared code
- [x] T040 Run full test suite: `make build && make test && make lint` - all pass
- [x] T041 Manual test: `yapi run examples/http/jq-filter.yapi.yml` - works
- [x] T042 Manual test: `yapi validate examples/http/jq-filter.yapi.yml` - works
- [x] T043 LSP shares AST helpers with validation package (langserver calls validation.FindVarPositionInYAML)

---

## Verification

```bash
make build && make test && make lint
yapi run examples/http.yapi.yml
yapi validate examples/http.yapi.yml
wc -l cmd/yapi/main.go  # Should be ~400 lines less
```

## Dependencies

```
Phase 1 (AST) ──┬──> Phase 2 (US1: Analyzer)
                └──> Phase 3 (US2: Line Numbers)

Phase 4 (US3: Executor) ──> independent
Phase 5 (ChainContext) ──> independent
Phase 6 (Main.go) ──> depends on Phase 4 (executor changes)
Phase 7 (Verification) ──> all phases complete
```

## Parallel Opportunities

- T001-T004 can run in parallel (different functions in same new file)
- T019 can run in parallel with other phases (new function, no dependencies)
- T028, T029, T032, T035 can run in parallel (different directories/files)

## Notes

- Tests use table-driven format
- Keep packages focused and small
- Error messages must be actionable
- Deprecation wrappers maintain backward compatibility
