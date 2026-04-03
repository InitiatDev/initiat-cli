package setup

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/output"
)

type Executor struct {
	secrets         map[string]string
	commandExecutor CommandExecutor
	formatter       *output.Formatter
}

func NewExecutor(secrets map[string]string) *Executor {
	return &Executor{
		secrets:         secrets,
		commandExecutor: NewRealCommandExecutor(),
		formatter:       output.NewFormatter(os.Stdout),
	}
}

func NewExecutorWithCommandExecutor(secrets map[string]string, executor CommandExecutor) *Executor {
	return &Executor{
		secrets:         secrets,
		commandExecutor: executor,
		formatter:       output.NewFormatter(os.Stdout),
	}
}

func NewExecutorWithFormatter(secrets map[string]string, executor CommandExecutor, formatter *output.Formatter) *Executor {
	return &Executor{
		secrets:         secrets,
		commandExecutor: executor,
		formatter:       formatter,
	}
}

func (e *Executor) Execute(plan *ExecutionPlan) error {
	report := &ExecutionReport{
		StartedAt: time.Now(),
		Summary:   plan.Summary,
		Commands:  []CommandExecutionRecord{},
	}

	f := e.formatter
	currentPhase := ""
	completedPhases := map[string]bool{}

	for _, cmd := range plan.Commands {
		if cmd.Phase != currentPhase {
			if currentPhase != "" {
				f.PhaseEnd()
			}
			currentPhase = cmd.Phase
			f.PhaseStart(currentPhase)
		}

		record, err := e.executeCommandWithReport(cmd)
		report.Commands = append(report.Commands, record)

		duration := lastAttemptDuration(record)
		stepLabel := cmd.StepName
		if stepLabel == "" {
			stepLabel = cmd.Description
		}

		if err != nil {
			stderr := lastAttemptStderr(record)
			f.StepFailure(stepLabel, duration, stderr)

			if cmd.ContinueOnError {
				continue
			}

			f.PhaseEnd()
			completedPhases[currentPhase] = true

			skipped := skippedPhaseNames(plan.Summary.Phases, completedPhases)
			if len(skipped) > 0 {
				f.PhasesSkipped(skipped)
			}

			report.FinishedAt = time.Now()
			return &SetupExecutionError{
				Report:        report,
				FailedCommand: record,
				Err:           fmt.Errorf("command failed: %w", err),
			}
		}

		f.StepSuccess(stepLabel, duration)
		completedPhases[cmd.Phase] = true
	}

	if currentPhase != "" {
		f.PhaseEnd()
	}

	report.FinishedAt = time.Now()
	return nil
}

func lastAttemptDuration(record CommandExecutionRecord) time.Duration {
	if len(record.Attempts) == 0 {
		return 0
	}
	return record.Attempts[len(record.Attempts)-1].Duration
}

func lastAttemptStderr(record CommandExecutionRecord) string {
	if len(record.Attempts) == 0 {
		return ""
	}
	return record.Attempts[len(record.Attempts)-1].Stderr
}

func skippedPhaseNames(phases []PhaseSummary, completed map[string]bool) []string {
	var skipped []string
	for _, p := range phases {
		if !completed[p.Name] {
			skipped = append(skipped, p.Name)
		}
	}
	return skipped
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
			if cmd.Retries != nil && cmd.Retries.Backoff > 0 {
				time.Sleep(cmd.Retries.Backoff)
			}
		}

		res, err := e.runCommand(cmd)
		record.Attempts = append(record.Attempts, e.toAttemptRecord(attempt, res, err))

		if err == nil {
			record.Success = true
			return record, nil
		}

		lastErr = err
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
