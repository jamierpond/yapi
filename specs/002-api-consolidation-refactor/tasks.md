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

- [ ] T016 [US2] Update `ValidateEnvFilesExistFromProject` to use `FindVarPositionInYAML` in `internal/validation/analyzer.go`
- [ ] T017 [US2] Update `ValidateProjectVars` to use AST helpers for line numbers in `internal/validation/analyzer.go`
- [ ] T018 [US2] Run `make test && make lint`

**Checkpoint**: `yapi validate` shows accurate line numbers for env file errors

---

## Phase 4: User Story 3 (P2) - Executor Simplification

**Goal**: Replace Factory with standalone `GetTransport()`

- [ ] T019 [P] [US3] Add `GetTransport(transport string, client HTTPClient) (TransportFunc, error)` in `internal/executor/executor.go`
- [ ] T020 [US3] Move switch logic from `Factory.Create()` to `GetTransport()` in `internal/executor/executor.go`
- [ ] T021 [US3] Add deprecation comments to `Factory` and `NewFactory` in `internal/executor/executor.go`
- [ ] T022 [US3] Update callers in `cmd/yapi/main.go` to use `GetTransport()` directly
- [ ] T023 [US3] Run `make test && make lint`

**Checkpoint**: `yapi run` works with all transport types

---

## Phase 5: ChainContext as Resolver

**Goal**: Deduplicate variable expansion logic

- [ ] T024 Add `Resolve(key string) (string, error)` method to `ChainContext` in `internal/runner/context.go`
- [ ] T025 Refactor `ExpandVariables` to delegate to `vars.ExpandString(input, c.Resolve)` in `internal/runner/context.go`
- [ ] T026 Remove duplicate regex handling from `ExpandVariables` in `internal/runner/context.go`
- [ ] T027 Run `make test && make lint`

**Checkpoint**: Variable expansion works, request chaining still functions

---

## Phase 6: Main.go Extraction

**Goal**: Reduce main.go by ~400 lines

- [ ] T028 [P] Create `internal/output/` directory
- [ ] T029 [P] Create `internal/output/result.go` with `JSONOutput` struct and `PrintJSON()` function
- [ ] T030 Move `printResultAsJSON` logic from `cmd/yapi/main.go` to `internal/output/result.go`
- [ ] T031 Update `cmd/yapi/main.go` to call `output.PrintJSON()`
- [ ] T032 [P] Create `internal/runner/stress.go` with `RunStress()` function
- [ ] T033 Move stress test worker pool logic from `cmd/yapi/main.go` to `internal/runner/stress.go`
- [ ] T034 Update `cmd/yapi/main.go` to call `runner.RunStress()`
- [ ] T035 [P] Create `internal/importer/cli.go` with `RunImport()` function
- [ ] T036 Move import CLI handler logic from `cmd/yapi/main.go` to `internal/importer/cli.go`
- [ ] T037 Update `cmd/yapi/main.go` to call `importer.RunImport()`
- [ ] T038 Run `make test && make lint`

**Checkpoint**: main.go reduced, all commands still work

---

## Phase 7: Verification & Cleanup

- [ ] T039 Count LOC: verify more code deleted than added
- [ ] T040 Run full test suite: `make build && make test && make lint`
- [ ] T041 Manual test: `yapi run examples/http.yapi.yml`
- [ ] T042 Manual test: `yapi validate examples/http.yapi.yml`
- [ ] T043 Manual test: LSP still provides accurate diagnostics

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
