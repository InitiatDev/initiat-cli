# Setup Script Implementation Plan

## Overview
Build the `.initiat/setup.yml` interpreter using a test-driven, incremental approach. Focus on pure functions (YAML → command strings) before wiring up CLI commands.

## Folder Structure
```
internal/setup/
  ├── types.go           # Core data structures
  ├── parser.go          # YAML parsing
  ├── validator.go       # Schema validation
  ├── condition.go       # Condition DSL evaluation
  ├── renderer.go        # Convert actions to command strings
  ├── matrix.go          # OS/arch matching
  ├── schema.go          # JSON Schema generation
  └── actions/
      ├── runtime.go     # ensure_runtime action
      ├── tool.go        # ensure_tool action
      ├── database.go    # ensure_database action
      └── package.go     # ensure_package_manager action

schemas/
  └── setup-v1.json      # Generated JSON Schema
```

## Implementation Phases

### Phase 1: Foundation - Types & Parsing
**Goal**: Parse YAML into Go structs

#### Step 1.1: Define Core Types
- [x] Create `internal/setup/types.go`
- [x] Define structs for:
  - `SetupConfig` (top-level)
  - `Matrix` (os/arch constraints)
  - `Defaults` (timeout, shell, etc)
  - `Step` (name, if, timeout, etc)
  - `Retries` (attempts, backoff)
- [x] Add YAML tags to all structs
- [x] Write tests in `types_test.go` for struct instantiation

#### Step 1.2: YAML Parser
- [x] Create `internal/setup/parser.go`
- [x] Implement `Parse([]byte) (*SetupConfig, error)`
- [x] Handle YAML unmarshaling with validation
- [x] Write tests in `parser_test.go`:
  - Valid minimal YAML
  - Valid complete YAML
  - Invalid YAML (syntax errors)
  - Missing required fields

**Deliverable**: ✅ Can parse `.initiat/setup.yml` into Go structs

---

### Phase 2: Validation
**Goal**: Validate parsed config against schema rules

#### Step 2.1: JSON Schema Generation
- [ ] Create `internal/setup/schema.go`
- [ ] Implement `GenerateJSONSchema() ([]byte, error)`
- [ ] Generate comprehensive JSON Schema v7 for setup.yml
- [ ] Include:
  - All action types as oneOf schemas
  - Condition DSL patterns
  - OS/arch enum values
  - Runtime type enums
  - Database engine enums
  - Custom validation rules (duration format, path safety)
- [ ] Write tests in `schema_test.go`:
  - Schema is valid JSON Schema v7
  - All action types are represented
  - Required fields are marked
  - Enum values are complete

#### Step 2.2: Schema Validator
- [ ] Create `internal/setup/validator.go`
- [ ] Implement `Validate(*SetupConfig) error`
- [ ] Use JSON Schema for basic validation
- [ ] Additional Go-specific validation:
  - `env_from_secrets` references exist in `env.secrets`
  - Path safety (no escaping repo root)
  - Business logic rules
- [ ] Write tests in `validator_test.go`:
  - Valid configs pass
  - Invalid version fails
  - Missing action fails
  - Multiple actions fail
  - Invalid secret reference fails
  - Invalid timeout format fails

**Deliverable**: JSON Schema + Go validator for configs

---

### Phase 3: Matrix & Conditions
**Goal**: Evaluate OS/arch matching and condition DSL

#### Step 3.1: Matrix Matching
- [ ] Create `internal/setup/matrix.go`
- [ ] Implement `MatchesMatrix(matrix Matrix, os, arch string) bool`
- [ ] Write tests in `matrix_test.go`:
  - Empty matrix matches all
  - OS match
  - Arch match
  - Combined OS+arch match
  - No match cases

#### Step 3.2: Condition DSL Parser
- [ ] Create `internal/setup/condition.go`
- [ ] Implement condition evaluator:
  - `EvaluateCondition(expr string, ctx ConditionContext) (bool, error)`
  - Support: `os()`, `arch()`, `file_exists()`, `cmd_ok()`
  - Support: `&&`, `||`, `!`, parentheses
- [ ] Write tests in `condition_test.go`:
  - Simple `os("macos")`
  - Combined `os("macos") && arch("arm64")`
  - File existence checks
  - Command success checks
  - Negation `!file_exists("x")`
  - Parentheses grouping
  - Invalid syntax handling

**Deliverable**: Can evaluate whether steps should run

---

### Phase 4: Action Renderers
**Goal**: Convert action definitions to executable command strings

#### Step 4.1: Base Action Interface
- [ ] Create `internal/setup/actions/types.go`
- [ ] Define `Action` interface:
  ```go
  type Action interface {
      Render(ctx RenderContext) ([]Command, error)
      Validate() error
  }
  ```
- [ ] Define `Command` struct (shell command + metadata)
- [ ] Define `RenderContext` (OS, arch, cwd, env vars)

#### Step 4.2: Simple Actions
- [ ] Create `internal/setup/actions/simple.go`
- [ ] Implement actions:
  - `RunAction` (shell command)
  - `PrintAction` (echo/print)
  - `AssertCommandAction` (command that must succeed)
- [ ] Write tests in `simple_test.go`:
  - Run renders correct shell command
  - Print renders OS-appropriate print command
  - Assert wraps command with exit code check

#### Step 4.3: Package Manager Action
- [ ] Create `internal/setup/actions/package.go`
- [ ] Implement `EnsurePackageManagerAction`
- [ ] Logic:
  - `auto` → detect OS and choose (brew/apt/choco)
  - Render install commands if not present
- [ ] Write tests in `package_test.go`:
  - macOS → brew installation
  - Linux → apt installation
  - Windows → choco installation
  - Already installed → no-op

#### Step 4.4: Tool Action
- [ ] Create `internal/setup/actions/tool.go`
- [ ] Implement `EnsureToolAction`
- [ ] Logic:
  - Check if tool exists and version matches
  - Render install command per OS (brew/apt/choco)
  - Handle fallback URLs with checksum
- [ ] Write tests in `tool_test.go`:
  - Tool not present → install
  - Tool present, version OK → no-op
  - Tool present, version old → upgrade
  - Fallback URL usage

#### Step 4.5: Runtime Action
- [ ] Create `internal/setup/actions/runtime.go`
- [ ] Implement `EnsureRuntimeAction`
- [ ] Support runtimes: node, python, go, elixir, erlang, java, rust, dotnet
- [ ] Logic:
  - Prefer asdf if available
  - Fall back to OS package manager
  - Version pinning with .tool-versions
- [ ] Write tests in `runtime_test.go`:
  - Each runtime type
  - Version matching
  - Manager preference (asdf > brew > apt)
  - Multiple fallback installers

#### Step 4.6: Database Action
- [ ] Create `internal/setup/actions/database.go`
- [ ] Implement `EnsureDatabaseAction`
- [ ] Support: postgres, mysql, sqlite
- [ ] Logic:
  - Install database
  - Start service
  - Create databases
- [ ] Write tests in `database_test.go`:
  - Install postgres
  - Ensure service running
  - Create database commands

#### Step 4.7: HTTP Assert Action
- [ ] Create `internal/setup/actions/http.go`
- [ ] Implement `AssertHTTPAction`
- [ ] Logic:
  - Render curl/wget command with retry logic
  - Check expected status code
- [ ] Write tests in `http_test.go`:
  - Basic HTTP GET
  - Status code validation
  - Retry logic

**Deliverable**: Can convert all action types to command strings

---

### Phase 5: Command Renderer
**Goal**: Orchestrate full setup execution plan

#### Step 5.1: Main Renderer
- [ ] Create `internal/setup/renderer.go`
- [ ] Implement `Render(*SetupConfig, RenderContext) (*ExecutionPlan, error)`
- [ ] Logic:
  - Validate config
  - Check matrix constraints
  - Process phases in order
  - Evaluate conditions for each step
  - Apply defaults/overrides
  - Render each action to commands
- [ ] Return `ExecutionPlan`:
  - List of commands to execute
  - Metadata (timeouts, retries, cwd, env)
- [ ] Write tests in `renderer_test.go`:
  - Full config rendering
  - Step skipping based on conditions
  - Default application
  - Phase ordering
  - Environment variable merging

**Deliverable**: Can produce complete execution plan from YAML

---

### Phase 6: Integration & Helpers
**Goal**: Utility functions for real execution

#### Step 6.1: Execution Context
- [ ] Create `internal/setup/context.go`
- [ ] Implement:
  - `NewContext() *Context` (detect OS, arch, cwd)
  - `WithSecrets(map[string]string) *Context`
  - `ShouldExecuteStep(step Step) (bool, error)`
- [ ] Write tests in `context_test.go`

#### Step 6.2: Timeout & Retry Helpers
- [ ] Create `internal/setup/execution.go`
- [ ] Implement:
  - `ParseDuration(string) (time.Duration, error)`
  - `ParseRetryPolicy(Retries) RetryPolicy`
- [ ] Write tests in `execution_test.go`

**Deliverable**: Helper utilities for execution

---

### Phase 7: CLI Integration
**Goal**: Wire up to actual CLI command

#### Step 7.1: Setup Command Structure
- [ ] Create `cmd/setup.go`
- [ ] Add subcommands:
  - `initiat setup validate` - validate YAML only
  - `initiat setup run` - execute setup
  - `initiat setup run --phase <phase>` - run up to phase
  - `initiat setup schema` - output JSON Schema

#### Step 7.2: Validate Subcommand
- [ ] Read `.initiat/setup.yml`
- [ ] Parse and validate
- [ ] Print success/errors

#### Step 7.3: Schema Subcommand
- [ ] Output generated JSON Schema to stdout
- [ ] Option to save to file: `--output schemas/setup-v1.json`

#### Step 7.4: Run Subcommand
- [ ] Read and parse YAML
- [ ] Create execution context
- [ ] Fetch secrets if needed
- [ ] Render execution plan
- [ ] Execute commands with:
  - Timeout enforcement
  - Retry logic
  - Secret redaction in logs
  - Progress output
- [ ] Handle failures per `continue_on_error`

**Deliverable**: Working CLI commands + JSON Schema output

---

## Testing Strategy

### Unit Tests
- Every function in isolation
- Mock file system operations
- Mock command execution checks
- Test both success and error paths

### Integration Tests
- Full YAML → execution plan
- Real file fixtures in `test/fixtures/setup_examples/`
- Example configs:
  - `minimal.yml`
  - `phoenix_basic.yml`
  - `node_basic.yml`
  - `multi_runtime.yml`
  - `conditional_steps.yml`

### Test Data
Create test fixtures directory:
```
test/fixtures/setup_examples/
  ├── minimal.yml
  ├── phoenix_basic.yml
  ├── node_basic.yml
  ├── invalid_version.yml
  ├── invalid_action.yml
  └── conditional_steps.yml
```

---

## Design Principles

1. **Pure Functions**: Most code is `YAML → data structures → commands`
2. **No Side Effects in Renderers**: Renderers only produce command strings, never execute
3. **Testability**: Every piece testable without filesystem/network
4. **Incremental**: Each phase builds on previous, independently testable
5. **Type Safety**: Leverage Go's type system, avoid `interface{}`
6. **Clear Errors**: Validation errors include line numbers and suggestions

---

## Success Criteria

- [x] Parse valid setup.yml files
- [ ] Generate comprehensive JSON Schema v7
- [ ] Validate against schema rules (JSON Schema + Go)
- [ ] Evaluate conditions correctly
- [ ] Render all action types to commands
- [ ] Full execution plan generation
- [x] >80% test coverage (91.7% for setup module)
- [ ] Working CLI commands
- [ ] JSON Schema output command
- [ ] No execution until Phase 7

## JSON Schema Benefits

### Reusability
- **Separate repo**: `initiat/setup-schema` for schema-only distribution
- **Language agnostic**: Any language can validate setup files
- **Version management**: Independent schema versioning

### Developer Experience
- **IDE support**: Auto-completion in VS Code, IntelliJ, etc.
- **Documentation**: Auto-generate docs from schema
- **Validation**: Pre-commit hooks, CI validation

### Ecosystem
- **Tool integration**: Other tools can understand setup format
- **Schema registry**: Centralized schema management
- **Migration tools**: Schema evolution support

---

## Next Steps

Start with **Phase 1: Foundation**. Create the types and parser, with comprehensive tests. Each phase should be complete and tested before moving to the next.

