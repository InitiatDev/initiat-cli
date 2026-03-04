package setup

import (
	"fmt"
	"time"
)

type ExecutionReport struct {
	StartedAt  time.Time
	FinishedAt time.Time
	Summary    ExecutionSummary
	Commands   []CommandExecutionRecord
}

type CommandExecutionRecord struct {
	Phase      string
	StepName   string
	StepIndex  int
	Command    string
	Args       []string
	WorkingDir string
	Timeout    time.Duration

	ContinueOnError bool
	Retries         *RetryPolicy

	EnvRedacted map[string]string
	Attempts    []CommandAttemptRecord
	Success     bool
}

type CommandAttemptRecord struct {
	Attempt   int
	Duration  time.Duration
	ExitCode  int
	Stdout    string
	Stderr    string
	TimedOut  bool
	ErrorText string
}

type SetupExecutionError struct {
	Report        *ExecutionReport
	FailedCommand CommandExecutionRecord
	Err           error
}

func (e *SetupExecutionError) Error() string {
	if e == nil {
		return "setup execution failed"
	}
	if e.Err == nil {
		return "setup execution failed"
	}
	stepName := e.FailedCommand.StepName
	if stepName == "" {
		stepName = fmt.Sprintf("step[%d]", e.FailedCommand.StepIndex)
	}
	return fmt.Sprintf("setup execution failed at %s/%s: %v", e.FailedCommand.Phase, stepName, e.Err)
}

func (e *SetupExecutionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
