# Agent-Assisted Setup Improvements

Checklist for improving the agent-assisted setup functionality. Each section is ordered by implementation priority — later items depend on earlier ones.

---

## 1. Capture command output and user responses

The agent currently discards stdout from successful commands and throws away user answers to `ask_user` prompts. This blocks all downstream improvements (verification, richer conversation) because the agent can't reason about what actually happened.

### 1.1 Add `Output` field to `AppliedActionResult`

- [x] Add `Output string` field to `AppliedActionResult` in `internal/agent/types.go`
- [x] Update `ApplyResult` JSON serialization to include output

### 1.2 Return stdout/stderr from `RunCommand`

- [x] Change `ToolRunner.RunCommand` signature to return `(string, error)` instead of `error`
  - Return `res.Stdout` on success, `res.Stderr` on failure
- [x] Update `LocalToolRunner.RunCommand` in `internal/agent/local_tools.go` to return command output
- [x] Update the `ToolRunner` interface in `internal/agent/orchestrator.go`
- [x] Update `applyOne` in `orchestrator.go` to capture returned output into `AppliedActionResult.Output`
- [x] Update all test mocks implementing `ToolRunner` to match new signature

### 1.3 Capture user responses from `ActionAskUser`

- [x] Update `applyOne` to handle `ActionAskUser` — call `prompt.PromptInput(action.Prompt)` and store the response in `r.Output`
- [x] Ensure `ActionAskUser` no longer silently no-ops

### 1.4 Feed output back into LLM context

- [x] Update `formatApplyResults` in `cmd/setup.go` to include `Output` (truncated to a reasonable limit, e.g. 4KB) in the results sent back to the agent
- [x] Update `buildAgentContextMsg` to pass action outputs so the LLM can reason about command results and user answers across rounds

### 1.5 Return output from `ReadFiles` and `ListFiles` in results

- [x] Capture the string returned by `ReadFiles` and `ListFiles` into `AppliedActionResult.Output` in `applyOne`, so interrogation results are also available in the apply result trail

---

## 2. Structured output formatting

Replace raw `fmt.Println` calls with a structured, visually clear output format so users can follow what's happening at a glance.

### 2.1 Create an output formatter

- [x] Create `internal/agent/output.go` (or `internal/output/output.go` if shared with setup)
- [x] Define a `Formatter` struct with methods for each output type:
  - `PhaseStart(name string)` — e.g. `┌─ bootstrap`
  - `StepSuccess(name string, duration time.Duration)` — e.g. `│  ✓ Install deps  2.3s`
  - `StepFailure(name string, duration time.Duration, err string)` — e.g. `│  ✗ Run migrations  0.8s`
  - `StepSkipped(name string)` — e.g. `│  ○ Skipped (condition not met)`
  - `PhaseEnd()`
  - `PhasesSkipped(names []string)` — e.g. `└─ (setup, verify, post skipped)`
- [x] Support a `--no-color` / `NO_COLOR` env var fallback for CI/non-TTY environments
- [x] Detect non-TTY output and disable box-drawing characters automatically

### 2.2 Apply formatter to setup execution

- [ ] Update `Executor.Execute` in `internal/setup/executor.go` to use the formatter instead of raw `fmt.Printf`
- [ ] Show per-command durations on the same line as the step name
- [ ] On failure, show the failing command's stderr indented beneath the step
- [ ] Show skipped phases (phases that never ran due to early failure) at the bottom

### 2.3 Apply formatter to agent mode

- [ ] Add agent-mode header output:
  ```
  ╔══ Agent Mode ═══════════════════════════════════════╗
  ║ Diagnosing failure in provision → "Run migrations"  ║
  ╚═════════════════════════════════════════════════════╝
  ```
- [ ] Add round separator: `── Round 1 ──────────────────`
- [ ] Format the diagnosis explanation in a visually distinct block (indented or quoted)
- [ ] Format proposed actions as a numbered list with danger-level badges:
  ```
  Proposed actions:
    1. [safe]    Read db/migrate/ directory listing
    2. [caution] Run: rails db:migrate:status
  ```
- [ ] Show action results inline after execution:
  ```
    1. [safe]    Read db/migrate/ directory listing  ✓
    2. [caution] Run: rails db:migrate:status        ✓ (exit 0)
  ```

### 2.4 Improve approval UX

- [ ] Replace per-action y/n prompts with a batch approval prompt when all actions are safe/caution:
  ```
  Approve all 3 actions? [y/n/pick]
  ```
- [ ] When user chooses `pick`, show numbered list and accept comma-separated selections (e.g. `1,3`)
- [ ] Keep per-action approval for `dangerous`-level actions — always require explicit individual confirmation
- [ ] Update `Approver` interface:
  ```go
  type ApprovalChoice int
  const (
      ApproveAll ApprovalChoice = iota
      ApproveSome
      DenyAll
  )
  ```
- [ ] Update `PromptApprover` to support batch mode
- [ ] Ensure the orchestrator handles partial approval — only execute approved actions, skip denied ones (currently it hard-fails on any denial)

---

## 3. Post-setup verification

After setup succeeds (all phases exit 0), the agent should verify the project actually works — not just that install commands didn't crash.

### 3.1 Define verification check types

- [ ] Add a `VerifyCheck` struct in `internal/agent/types.go`:
  ```go
  type VerifyCheck struct {
      Type        string // "command", "http", "tcp"
      Target      string // command string, URL, or host:port
      Expect      string // expected substring, status code, "connected"
      Timeout     string // per-check timeout (e.g. "10s")
      Description string // human-readable label
  }
  ```
- [ ] Add `ActionVerify` to `ProposedActionType`
- [ ] Add `Checks []VerifyCheck` field to `ProposedAction`

### 3.2 Implement verification runners

- [ ] Add `RunVerifyChecks(ctx context.Context, checks []VerifyCheck) []VerifyResult` to `LocalToolRunner`
- [ ] Implement **command** checks: run command, check exit code and stdout contains `Expect`
- [ ] Implement **HTTP** checks: `GET` the URL, check status code matches `Expect` (e.g. "200")
  - Support waiting with retries (server may take a moment to start)
  - Configurable timeout per check
- [ ] Implement **TCP** checks: dial `host:port`, check connection succeeds
  - Useful for databases, Redis, etc.
- [ ] Return structured results:
  ```go
  type VerifyResult struct {
      Check   VerifyCheck
      Passed  bool
      Output  string // stdout or response body snippet
      Error   string
      Elapsed time.Duration
  }
  ```

### 3.3 Add verification system prompt

- [ ] Create `buildVerifySystemPrompt()` in `orchestrator.go` that instructs the LLM:
  - "Setup completed successfully. Based on the project type and structure, propose verification checks."
  - For CLIs: run the binary with `--version` or `--help`
  - For web apps: start the dev server, check a health/root endpoint, then stop it
  - For libraries: run the test suite
  - For Docker projects: check containers are running, ports are accessible
  - Keep checks lightweight — verification should take seconds, not minutes

### 3.4 Add `Orchestrator.Verify` method

- [ ] Add `Verify(ctx context.Context, snapshot *ProjectSnapshot, report *ExecutionReport) (*VerifyDecision, error)`
- [ ] This calls the LLM with the verify system prompt + project snapshot
- [ ] LLM returns proposed verification checks
- [ ] Run checks, collect results, report pass/fail

### 3.5 Wire verification into the setup command

- [ ] In `cmd/setup.go` `runSetupRun`, after successful `runner.Run()`:
  ```go
  if cfg.Agent.Enabled {
      maybeRunAgentVerification(ctx, setupConfig, wd)
  }
  ```
- [ ] `maybeRunAgentVerification`:
  - Build project snapshot
  - Call `orchestrator.Verify()`
  - Display results using the formatter:
    ```
    ── Verification ─────────────────────────────────
      ✓ `myapp --version` → "v1.2.3"          0.1s
      ✓ GET http://localhost:3000/health → 200  1.2s
      ✗ TCP localhost:5432 → connection refused 0.0s
    
    2/3 checks passed. 1 failed.
    ```
  - If any checks fail, offer to enter agent mode to diagnose
- [ ] Add `--skip-verify` flag to `setup run` to opt out
- [ ] Add `--verify-only` flag to run only verification against an already-set-up project

### 3.6 Handle long-running verification targets

- [ ] For web apps/servers, the agent needs to start the server as a background process, verify, then stop it
- [ ] Add a `BackgroundCommand` concept:
  - Start process, wait for readiness (poll HTTP/TCP), run checks, send SIGTERM
  - Capture stdout/stderr for diagnostics if startup fails
- [ ] Set a global verification timeout (e.g. 60s) to prevent hanging on a server that won't start

---

## 4. Better error context for the LLM

Improve the information the agent receives to make diagnoses more accurate.

### 4.1 Include phase context in diagnosis

- [ ] Add the failed phase name (bootstrap/provision/setup/verify) to the system prompt or context message
- [ ] Add guidance per phase:
  - `bootstrap` failures → likely missing system dependencies or wrong runtime version
  - `provision` failures → likely database/service connectivity or state issues
  - `setup` failures → likely app configuration or build errors
  - `verify` failures → likely the app isn't working despite install succeeding

### 4.2 Detect first-run vs re-run

- [ ] Check for markers of a previous setup attempt (e.g. `node_modules/` exists, database already created)
- [ ] Pass this to the agent: "This appears to be a re-run. Previous state may be causing conflicts."
- [ ] Guide the agent to consider dirty state (stale lock files, partially migrated DB, leftover containers)

### 4.3 Include platform context

- [ ] Pass `runtime.GOOS` and `runtime.GOARCH` to the agent context
- [ ] If matrix is defined in setup.yml, include which matrix variant is being executed
- [ ] Guide platform-specific suggestions (e.g. `brew install` on macOS, `apt-get` on Linux)

---

## 5. Proactive agent mode

Allow the agent to participate throughout setup, not just on failure.

### 5.1 Add `--agent` flag to `setup run`

- [ ] Add `--agent` flag to `setupRunCmd`
- [ ] When set, enable agent involvement at phase boundaries (not just on failure)

### 5.2 Pre-phase review

- [ ] Before each phase starts, send the phase's steps to the LLM
- [ ] LLM can flag potential issues (e.g. "This step runs `npm install` but no `package.json` was found")
- [ ] Show warnings to user, let them continue or abort
- [ ] Keep this fast — use a lightweight prompt, don't over-interrogate

### 5.3 Post-phase validation

- [ ] After each phase completes successfully, optionally ask the LLM to validate outcomes
- [ ] E.g. after `bootstrap`: "Dependencies installed. Check that expected binaries are on PATH."
- [ ] This is a lighter version of the full verification in section 3 — quick sanity checks between phases

### 5.4 Conversational mode

- [ ] After all phases + verification, drop into an optional interactive session:
  ```
  Setup complete. All checks passed.
  
  Ask the agent anything about this project, or press Enter to exit:
  > How do I run the test suite?
  ```
- [ ] Use the accumulated project context (snapshot, README, execution results) to answer questions
- [ ] Exit on empty input or `quit`/`exit`
