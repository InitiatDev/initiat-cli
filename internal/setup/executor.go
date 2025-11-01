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
	fmt.Printf("🚀 Executing setup script: %d phases, %d steps, %d commands\n\n",
		len(plan.Summary.Phases), plan.Summary.TotalSteps, plan.Summary.TotalCommands)

	for _, cmd := range plan.Commands {
		fmt.Printf("[%s] %s", cmd.Phase, cmd.StepName)
		if cmd.Description != "" {
			fmt.Printf(": %s", cmd.Description)
		}
		fmt.Println()

		if err := e.executeCommand(cmd); err != nil {
			if cmd.ContinueOnError {
				fmt.Printf("⚠️  Command failed but continuing: %v\n", err)
				continue
			}
			return fmt.Errorf("command failed: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("✅ Setup script completed successfully!")
	return nil
}

func (e *Executor) executeCommand(cmd ExecutableCommand) error {
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

		err := e.runCommand(cmd)
		if err == nil {
			if attempt > 1 {
				fmt.Printf("  ✅ Command succeeded on retry\n")
			}
			return nil
		}

		lastErr = err
		if attempt < attempts {
			fmt.Printf("  ⚠️  Command failed, will retry: %v\n", err)
		}
	}

	return fmt.Errorf("command failed after %d attempts: %w", attempts, lastErr)
}

func (e *Executor) runCommand(cmd ExecutableCommand) error {
	env := make(map[string]string)
	for k, v := range cmd.Env {
		if e.shouldRedact(k) {
			env[k] = "[REDACTED]"
		} else {
			env[k] = v
		}
	}

	req := &CommandRequest{
		Command:    cmd.Command,
		Args:       cmd.Args,
		Env:        env,
		WorkingDir: cmd.WorkingDir,
		Timeout:    cmd.Timeout,
	}

	ctx := context.Background()
	return e.commandExecutor.Execute(ctx, req)
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
