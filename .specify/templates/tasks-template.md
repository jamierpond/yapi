# Tasks: [FEATURE NAME]

**Input**: `specs/[###-feature-name]/plan.md`

## Format

- `[P]` = Can run in parallel
- `[US#]` = User story reference
- Include file paths: `internal/executor/http.go`

## Phase 1: Setup

- [ ] T001 [P] Add CLI command in `cmd/yapi/`
- [ ] T002 [P] Add config struct in `internal/config/`

---

## Phase 2: Core Implementation

### User Story 1 (P1)

- [ ] T003 [US1] Implement core logic in `internal/[pkg]/`
- [ ] T004 [US1] Add YAML schema support in `internal/config/`
- [ ] T005 [US1] Wire up CLI command

**Checkpoint**: `yapi [command]` works for basic case

### User Story 2 (P2)

- [ ] T006 [US2] Extend for additional use case
- [ ] T007 [US2] Add error handling

---

## Phase 3: Protocol Support

- [ ] T008 [P] HTTP support in `internal/executor/http.go`
- [ ] T009 [P] gRPC support in `internal/executor/grpc.go`
- [ ] T010 [P] GraphQL support in `internal/executor/graphql.go`
- [ ] T011 [P] TCP support in `internal/executor/tcp.go`

---

## Phase 4: Testing & Polish

- [ ] T012 [P] Unit tests in `internal/[pkg]/*_test.go`
- [ ] T013 [P] Integration test in `tests/`
- [ ] T014 Add example in `examples/`
- [ ] T015 Update README if needed

---

## Verification

```bash
make build && make test && make lint
yapi [new-command] examples/[example].yapi.yml
```

## Notes

- Tests use table-driven format
- Keep packages focused and small
- Prefer composition over inheritance
- Error messages must be actionable
