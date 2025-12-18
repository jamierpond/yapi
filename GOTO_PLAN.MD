# Implementation Plan: Goto Definition for Project Variables in LSP

## Overview
Add textDocument/definition support to the yapi LSP server to enable "goto definition" functionality for project variables. When a user places their cursor on a variable reference like `${API_URL}` in a `.yapi.yml` file and triggers "goto definition" (F12 or Cmd+Click), they should jump to where that variable is defined in `yapi.config.yml`.

## Current State
- LSP server is implemented in `/Users/jamiepond/projects/yapi/internal/langserver/langserver.go`
- Already supports hover (shows variable values), completion, and diagnostics
- Uses `github.com/tliron/glsp` v0.2.2 library and `gopkg.in/yaml.v3` for YAML parsing
- Variables are defined in `yapi.config.yml` under `environments.[env_name].vars` or `defaults.vars`
- Variables are referenced in `.yapi.yml` files using `${var}` or `$var` syntax
- The `FindEnvVarRefs()` function already extracts variable references with position info

## Implementation Approach

### 1. Register the Definition Handler
**File:** `/Users/jamiepond/projects/yapi/internal/langserver/langserver.go`

Add `TextDocumentDefinition: textDocumentDefinition,` to the handler map in the `Run()` function (around line 53, after `TextDocumentHover`).

No changes needed to the `initialize()` function - the glsp library automatically advertises the definition capability when the handler is registered.

### 2. Implement the Definition Handler
**File:** `/Users/jamiepond/projects/yapi/internal/langserver/langserver.go`

Create a new function `textDocumentDefinition(ctx *glsp.Context, params *protocol.DefinitionParams) (any, error)` that:

1. Gets the document from the URI
2. Extracts the cursor position from params
3. Uses `validation.FindEnvVarRefs(doc.Text)` to find all variable references (reuse existing code)
4. Checks if cursor position falls within any variable reference
5. Skips chain references (containing dots like `${step.field}`)
6. Checks if project context exists (`doc.Project != nil`)
7. Calls a helper function to find the variable definition location
8. Returns `protocol.Location` or `nil`

### 3. Find Variable Definition in Config
**File:** `/Users/jamiepond/projects/yapi/internal/langserver/langserver.go`

Create a helper function `findVariableDefinition(varName string, project *config.ProjectConfigV1, projectRoot string) (*protocol.Location, error)` that:

1. Determines which environment to use (default_environment or first alphabetically)
2. Resolves environment variables using `project.ResolveEnvFiles()` to check if variable exists
3. Returns `nil` if variable only exists in OS environment (can't jump to it)
4. Determines which YAML section contains the variable:
   - `environments.[env_name].vars`
   - `defaults.vars`
5. Finds the `yapi.config.yml` file path (try `.yml` then `.yaml`)
6. Calls a YAML parser helper to find the exact position
7. Converts to a `protocol.Location` with file URI and range
8. Returns the location

### 4. Parse YAML for Position Info
**File:** `/Users/jamiepond/projects/yapi/internal/langserver/langserver.go`

Create a helper function `findVarPositionInYAML(yamlPath string, varName string, section []string) (*protocol.Range, error)` that:

1. Reads the `yapi.config.yml` file
2. Unmarshals into `yaml.Node` (which preserves line/column info)
3. Traverses the node tree following the section path (e.g., `["environments", "dev", "vars"]`)
4. Finds the key node matching the variable name
5. Extracts `Line` and `Column` from the yaml.Node (1-indexed)
6. Converts to `protocol.Range` (0-indexed)
7. Calculates end position as start + variable name length
8. Returns the range

**Helper functions needed:**
- `findNodeInMapping(mapping *yaml.Node, key string) *yaml.Node` - finds value node for a key
- `findKeyNodeInMapping(mapping *yaml.Node, key string) *yaml.Node` - finds the key node itself (for position)

### 5. Environment Selection Helper
**File:** `/Users/jamiepond/projects/yapi/internal/langserver/langserver.go`

Create `getEffectiveEnvironment(project *config.ProjectConfigV1) string` that:
- Returns `project.DefaultEnvironment` if set
- Otherwise returns first environment name alphabetically
- Used to determine which environment's variables to search when multiple exist

## Key Technical Details

**YAML Node Traversal:**
- Root node is a `DocumentNode` containing a `MappingNode`
- MappingNodes have `Content` as alternating `[key, value, key, value, ...]`
- Navigate to `environments.dev.vars.API_URL` by following key→value→key→value path
- The key node contains the `Line` and `Column` we need

**Position Conversion:**
- `yaml.Node`: Line and Column are 1-indexed
- LSP `protocol.Position`: Line and Character are 0-indexed
- Convert: `Line: node.Line - 1, Character: node.Column - 1`

**URI Format:**
- Convert file path to URI: `protocol.DocumentUri("file://" + configPath)`

## Edge Cases to Handle

1. **Variable only in .env files**: Return `nil` (can't provide precise navigation)
2. **Chain references** (`${step.field}`): Skip these - already filtered by dot check
3. **OS environment variables only**: Return `nil` (no file to jump to)
4. **Config file not found**: Return `nil` gracefully
5. **YAML parsing errors**: Return `nil` to avoid breaking LSP
6. **No project context**: Return `nil`
7. **Variable not found in config**: Return `nil`

## Critical Files

- `/Users/jamiepond/projects/yapi/internal/langserver/langserver.go` - Main implementation file
- `/Users/jamiepond/projects/yapi/internal/validation/analyzer.go` - Reference for `FindEnvVarRefs()` (lines 478-566)
- `/Users/jamiepond/projects/yapi/internal/config/project.go` - Reference for `ProjectConfigV1` and `ResolveEnvFiles()`
- `/Users/jamiepond/projects/yapi/examples/project/example1/yapi.config.yml` - Test file for verification

## Testing Approach

Manual testing with the example project:
1. Open `/Users/jamiepond/projects/yapi/examples/project/example1/.yapi/get-users.yapi.yml`
2. Place cursor on `${API_URL}` and trigger goto definition
3. Should jump to line 16 in `yapi.config.yml` (dev environment)
4. Test with `${TIMEOUT}` - should jump to line 10 (defaults section)
5. Test with undefined variable - should do nothing
6. Test with chain reference `${step.field}` - should do nothing

## Dependencies
No new dependencies required. Uses existing:
- `github.com/tliron/glsp` v0.2.2
- `gopkg.in/yaml.v3`
- Standard library packages
