# Research: API Consolidation Refactor

**Date**: 2026-01-03
**Feature**: 002-api-consolidation-refactor

## Phase 0: Investigation Findings

### 1. Dead Code Analysis (FR-002)

**Original Assumption**: `EnvFileStatus`, `EnvFileLoadResult`, and `ResolverOptions` are unused dead code.

**Finding**: All three structs are **actively used**:

| Struct | Location | Usage |
|--------|----------|-------|
| `EnvFileStatus` | loader.go:20-29 | Validates env file references with position info |
| `EnvFileLoadResult` | loader.go:31-36 | Container for env file load results |
| `ResolverOptions` | loader.go:243-246 | Configures StrictEnv mode through loader chain |

**Decision**: Do NOT delete these structs. FR-002 is **invalid** based on codebase analysis.

**Rationale**: These structs are integral to the env file loading system. Removing them would break the loader.

---

### 2. Analyzer API Analysis (FR-001)

**Current State**: Four functions exist in a nested call chain:

```
AnalyzeConfigString(text)
  └─> AnalyzeConfigStringWithProject(text, project, projectRoot)
      └─> AnalyzeConfigStringWithProjectAndPath(text, configPath, project, projectRoot)
          └─> AnalyzeConfigStringWithProjectAndPathAndOptions(text, configPath, project, projectRoot, opts)
```

**Locations**:
- Line 111: `AnalyzeConfigString(text string)`
- Line 118: `AnalyzeConfigStringWithProject(text, project, projectRoot)`
- Line 124: `AnalyzeConfigStringWithProjectAndPath(text, configPath, project, projectRoot)`
- Line 129: `AnalyzeConfigStringWithProjectAndPathAndOptions(text, configPath, project, projectRoot, opts)`

**Decision**: Consolidate to single `Analyze(text string, opts AnalyzeOptions)` function.

**Rationale**: Current API has combinatorial explosion. An options struct provides:
- Cleaner API surface
- Easier to add new options without new function variants
- Aligns with Constitution Principle IV (Simplicity)

**Migration Path**: Keep old functions as deprecated wrappers calling new `Analyze()` to avoid breaking changes.

---

### 3. Executor Factory Analysis (FR-003)

**Current State** (executor.go:22-51):
```go
type Factory struct {
    Client HTTPClient
}

func NewFactory(client HTTPClient) *Factory
func (f *Factory) Create(transport string) (TransportFunc, error)
```

**Decision**: Replace Factory with standalone `GetTransport(transport string, client HTTPClient)`.

**Rationale**:
- Factory holds single field (HTTPClient) with no state
- Single method (Create) makes it a disguised function
- Aligns with Constitution Principle IV (Simplicity)

**Alternatives Rejected**:
- Keep Factory: Added indirection with no benefit
- Multiple standalone functions per transport: Loses unified entry point

---

### 4. AST Logic Analysis (FR-004, FR-005)

**Current State** (langserver.go):
- `findVarPositionInYAML` (line 787)
- `findNodeInMapping` (line 906)
- `findKeyNodeInMapping` (line 922)
- Plus 4 additional navigation helpers

**Problem**: These functions are specific to langserver but could serve validation diagnostics.

**Decision**: Extract to `internal/validation/ast.go` with protocol-agnostic `Location` type.

**Rationale**:
- Enables accurate line numbers in validation diagnostics
- Aligns with Constitution Principle VII (Single Code Path)
- LSP and CLI validation share same position-finding logic

---

### 5. ChainContext Variable Expansion (FR-006, FR-007)

**Current State** (runner/context.go:57-102):
- Uses `vars.Expansion.ReplaceAllStringFunc`
- Priority: OS env > EnvOverrides > Chain results
- Inline regex matching

**Decision**: Have `ChainContext` implement `vars.Resolver` interface and delegate to `vars.ExpandString()`.

**Rationale**:
- Removes duplicate regex handling
- Uses vars package as single source of expansion logic
- Aligns with Constitution Principle VII (Single Code Path)

---

### 6. Compiler for Validation (FR-006)

**Current State**:
- `compiler.Compile()` (compiler.go:25) already handles config transformation
- Validation duplicates some compilation checks

**Decision**: Use `compiler.Compile()` with `vars.MockResolver` during validation.

**Rationale**:
- Compiler is source of truth for config interpretation
- Validation should use same path as runtime
- Aligns with Constitution Principle VII

---

### 7. Main.go Extraction Analysis (FR-008, FR-009, FR-010)

**Current State**: main.go is 2,244 lines with identifiable modules:

| Module | Lines | Extractable |
|--------|-------|-------------|
| JSON output | ~127 | Yes - `printResultAsJSON` |
| Stress testing | ~114 | Yes - worker pool logic |
| Import handling | ~180+ | Yes - Postman import |
| Config validation | ~180 | Partial - orchestration only |

**Decision**: Extract JSON output, stress testing, and import handling.

**Rationale**:
- Reduces main.go cognitive load
- Creates testable, reusable modules
- Aligns with Constitution Principle VI (Minimal Code)

---

## Research Conclusions

| Requirement | Feasible | Notes |
|-------------|----------|-------|
| FR-001: Consolidate Analyzer | Yes | Options struct pattern |
| FR-002: Remove dead structs | **No** | Structs are actively used |
| FR-003: Remove Factory | Yes | Standalone function |
| FR-004: Extract AST logic | Yes | New validation/ast.go |
| FR-005: Fix line 0 diagnostics | Yes | Use AST helpers |
| FR-006: Use compiler for validation | Yes | With MockResolver |
| FR-007: ChainContext as Resolver | Yes | Delegate to vars |
| FR-008: Extract JSON output | Yes | internal/output/ |
| FR-009: Extract stress test | Yes | internal/runner/stress.go |
| FR-010: Extract import CLI | Yes | internal/importer/cli.go |

**Net Impact**: 9 of 10 requirements are valid. FR-002 (remove dead code) is dropped as structs are actively used. This aligns with success criteria "more code deleted than added" through consolidation, not misguided deletion of functional code.
