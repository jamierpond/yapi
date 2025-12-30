Based on the `internal/vars` and `internal/validation` codebase, here is the technical breakdown of why the current "lazy" (`$VAR`) vs. "strict" (`${VAR}`) expansion creates complexity and the recommended long-term strategy for the variable system.

### The Current Conflict

The system currently treats `$VAR` and `${VAR}` as functionally identical for **OS/Project Environment variables**, but this causes significant collisions in specific contexts:

1. **GraphQL Collisions:** GraphQL uses `$variable` natively for operations. Currently, `yapi` has to hard-code an "ignore" rule for lazy variables in GraphQL blocks to avoid breaking valid queries.
2. **Ambiguity in Chains:** In request chains, a reference like `${auth.token}` is clear. However, a lazy reference like `$auth.token` is harder to parse accurately because the system has to guess where the "variable" ends and the "property path" begins.
3. **Cross-Platform Inconsistency:** Shell environments treat `$VAR` and `${VAR}` differently regarding word splitting and concatenation, but `yapi` tries to unify them, leading to unexpected behavior when users copy-paste from their terminal.

---

### Long-Term Recommendation: **"Strict-Only" Deprecation Path**

To resolve these complexities, the system should move toward a **Strict-Only** expansion model.

#### 1. Deprecate `$VAR` (Lazy Syntax)

* **Action:** Mark `$VAR` as deprecated in the next minor version.
* **Why:** It removes the ambiguity in GraphQL and Chain references. It forces users to be explicit about boundaries, which is safer for token concatenation (e.g., `Bearer ${TOKEN}` vs. `Bearer $TOKEN_INVALID`).
* **Implementation:** The `vars.Expansion` regex should be updated to prioritize `${}` and eventually only support it.

#### 2. Reserve `$` for "Dynamic Context"

* **Strategy:** Reserve the single `$` prefix for internal, built-in "dynamic" variables (similar to Postman's `$guid` or `$timestamp`).
* **Benefit:** This distinguishes between **User/Env Variables** (`${MY_VAR}`) and **System Helpers** (`$guid`).

#### 3. Introduce Explicit Namespacing

Currently, the system guesses if a variable is a chain reference by checking for a dot (`.`). This is brittle.

* **Proposed Long-term Syntax:**
* **Environment:** `${env.VAR_NAME}` (instead of just `${VAR_NAME}`)
* **Chain Results:** `${steps.step_name.field}`
* **Built-ins:** `$now`, `$uuid`


* **Why:** This allows the compiler to validate variables *before* execution, providing better LSP diagnostics for missing environment keys vs. missing chain steps.

#### 4. "Safe-String" GraphQL Handling

Instead of just ignoring `$` in GraphQL, the long-term fix is a **template literal approach**.

* **Action:** Only expand variables inside GraphQL if they are wrapped in `${}`.
* **Result:** Standard GraphQL variables (like `$id: ID!`) remain untouched, while `yapi` injected variables (like `${USER_ID}`) are swapped out before the request is sent.

### Summary Table: Long-Term Variable Map

| Type | Current (Mixed) | Long-Term (Proposed) |
| --- | --- | --- |
| **Env Variable** | `$API_KEY` or `${API_KEY}` | `${env.API_KEY}` |
| **Chain Ref** | `${auth.token}` | `${steps.auth.token}` |
| **Built-in** | N/A | `$uuid`, `$timestamp` |
| **GraphQL** | Mixed (Collision prone) | `${var}` for inject, `$var` for native |

By moving to **Explicit Strict Expansion**, you eliminate the "guessing" logic in `internal/vars/expand.go` and make the system predictable across HTTP, gRPC, and GraphQL transports.
