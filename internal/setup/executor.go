package setup

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Executor struct {
	secrets         map[string]string
	commandExecutor CommandExecutor
}

func NewExecutor(secrets map[string]string) *Executor {
	return &Executor{
		secrets:         secrets,
		commandExecutor: NewRealCommandExecutor(),
	}
}

func NewExecutorWithCommandExecutor(secrets map[string]string, executor CommandExecutor) *Executor {
	return &Executor{
		secrets:         secrets,
		commandExecutor: executor,
	}
}

func (e *Executor) Execute(plan *ExecutionPlan) error {
	report := &ExecutionReport{
		StartedAt: time.Now(),
		Summary:   plan.Summary,
		Commands:  []CommandExecutionRecord{},
	}

	fmt.Printf("🚀 Executing setup script: %d phases, %d steps, %d commands\n\n",
		len(plan.Summary.Phases), plan.Summary.TotalSteps, plan.Summary.TotalCommands)

	for _, cmd := range plan.Commands {
		fmt.Printf("[%s] %s", cmd.Phase, cmd.StepName)
		if cmd.Description != "" {
			fmt.Printf(": %s", cmd.Description)
		}
		fmt.Println()

		record, err := e.executeCommandWithReport(cmd)
		report.Commands = append(report.Commands, record)

		if err != nil {
			if cmd.ContinueOnError {
				fmt.Printf("⚠️  Command failed but continuing: %v\n", err)
				continue
			}
			report.FinishedAt = time.Now()
			return &SetupExecutionError{
				Report:        report,
				FailedCommand: record,
				Err:           fmt.Errorf("command failed: %w", err),
			}
		}
	}

	fmt.Println()
	fmt.Println("✅ Setup script completed successfully!")
	report.FinishedAt = time.Now()
	return nil
}

func (e *Executor) executeCommandWithReport(cmd ExecutableCommand) (CommandExecutionRecord, error) {
	record := CommandExecutionRecord{
		Phase:           cmd.Phase,
		StepName:        cmd.StepName,
		StepIndex:       cmd.StepIndex,
		Command:         cmd.Command,
		Args:            cmd.Args,
		WorkingDir:      cmd.WorkingDir,
		Timeout:         cmd.Timeout,
		ContinueOnError: cmd.ContinueOnError,
		Retries:         cmd.Retries,
		EnvRedacted:     e.redactEnv(cmd.Env),
		Attempts:        []CommandAttemptRecord{},
	}

	var lastErr error
	attempts := 1
	if cmd.Retries != nil && cmd.Retries.Attempts > 0 {
		attempts = cmd.Retries.Attempts
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			fmt.Printf("  ↻ Retry attempt %d/%d...\n", attempt, attempts)
			if cmd.Retries != nil && cmd.Retries.Backoff > 0 {
				time.Sleep(cmd.Retries.Backoff)
			}
		}

		res, err := e.runCommand(cmd)
		record.Attempts = append(record.Attempts, e.toAttemptRecord(attempt, res, err))

		if err == nil {
			if attempt > 1 {
				fmt.Printf("  ✅ Command succeeded on retry\n")
			}
			record.Success = true
			return record, nil
		}

		lastErr = err
		if attempt < attempts {
			fmt.Printf("  ⚠️  Command failed, will retry: %v\n", err)
		}
	}

	record.Success = false
	return record, fmt.Errorf("command failed after %d attempts: %w", attempts, lastErr)
}

func (e *Executor) runCommand(cmd ExecutableCommand) (*CommandResult, error) {
	req := &CommandRequest{
		Command:    cmd.Command,
		Args:       cmd.Args,
		Env:        cmd.Env,
		WorkingDir: cmd.WorkingDir,
		Timeout:    cmd.Timeout,
	}

	ctx := context.Background()
	return e.commandExecutor.Execute(ctx, req)
}

func (e *Executor) redactEnv(env map[string]string) map[string]string {
	if env == nil {
		return nil
	}
	out := make(map[string]string, len(env))
	for k, v := range env {
		if e.shouldRedact(k) {
			out[k] = "[REDACTED]"
		} else {
			out[k] = v
		}
	}
	return out
}

func (e *Executor) toAttemptRecord(attempt int, res *CommandResult, err error) CommandAttemptRecord {
	rec := CommandAttemptRecord{
		Attempt: attempt,
	}
	if res != nil {
		rec.Duration = res.Duration
		rec.ExitCode = res.ExitCode
		rec.Stdout = e.redactOutput(res.Stdout)
		rec.Stderr = e.redactOutput(res.Stderr)
		rec.TimedOut = res.TimedOut
	} else {
		rec.ExitCode = -1
	}
	if err != nil {
		rec.ErrorText = err.Error()
	}
	return rec
}

func (e *Executor) redactOutput(s string) string {
	if e.secrets == nil || s == "" {
		return s
	}
	redacted := s
	for _, v := range e.secrets {
		if v == "" {
			continue
		}
		redacted = strings.ReplaceAll(redacted, v, "[REDACTED]")
	}
	return redacted
}

func (e *Executor) shouldRedact(key string) bool {
	keyLower := strings.ToLower(key)
	if e.secrets == nil {
		return isRedactable(keyLower)
	}

	for secretKey := range e.secrets {
		if strings.EqualFold(key, secretKey) {
			return true
		}
	}

	return isRedactable(keyLower)
}

func isRedactable(key string) bool {
	keyLower := strings.ToLower(key)
	return strings.Contains(keyLower, "secret") ||
		strings.Contains(keyLower, "password") ||
		strings.Contains(keyLower, "token") ||
		strings.Contains(keyLower, "api_key") ||
		strings.Contains(keyLower, "access_token") ||
		strings.Contains(keyLower, "refresh_token")
}
