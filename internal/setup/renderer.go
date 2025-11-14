package setup

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type RenderContext struct {
	OS                     string
	Arch                   string
	WorkingDir             string
	Shell                  string
	Secrets                map[string]string
	GlobalEnv              map[string]string
	DefaultTimeout         time.Duration
	DefaultContinueOnError bool
}

type ExecutionPlan struct {
	Commands []ExecutableCommand
	Summary  ExecutionSummary
}

type ExecutableCommand struct {
	Phase           string
	StepName        string
	StepIndex       int
	Command         string
	Args            []string
	Env             map[string]string
	WorkingDir      string
	Timeout         time.Duration
	Description     string
	Retries         *RetryPolicy
	ContinueOnError bool
}

type RetryPolicy struct {
	Attempts int
	Backoff  time.Duration
}

type ExecutionSummary struct {
	TotalSteps    int
	TotalCommands int
	Phases        []PhaseSummary
}

type PhaseSummary struct {
	Name         string
	StepCount    int
	CommandCount int
}

func Render(config *SetupConfig, ctx *RenderContext) (*ExecutionPlan, error) {
	if err := validateMatrix(config, ctx); err != nil {
		return nil, err
	}

	conditionEval := NewConditionEvaluator(ctx.OS, ctx.Arch, mergeEnv(ctx.GlobalEnv, ctx.Secrets))

	plan, summary, err := processPhases(config, ctx, conditionEval)
	if err != nil {
		return nil, err
	}

	plan.Summary = *summary
	return plan, nil
}

func validateMatrix(config *SetupConfig, ctx *RenderContext) error {
	matcher := NewMatrixMatcher()
	matcher.OS = ctx.OS
	matcher.Arch = ctx.Arch

	matches, err := matcher.Matches(config)
	if err != nil {
		return fmt.Errorf("matrix matching failed: %w", err)
	}
	if !matches {
		return fmt.Errorf("current platform %s/%s does not match matrix constraints", ctx.OS, ctx.Arch)
	}

	return nil
}

func processPhases(
	config *SetupConfig,
	ctx *RenderContext,
	conditionEval *ConditionEvaluator,
) (*ExecutionPlan, *ExecutionSummary, error) {
	var plan ExecutionPlan
	summary := ExecutionSummary{
		Phases: []PhaseSummary{},
	}

	for _, phase := range GetAllPhases(config) {
		phaseSummary := PhaseSummary{
			Name:         phase.Name,
			StepCount:    len(phase.Steps),
			CommandCount: 0,
		}

		for stepIndex, step := range phase.Steps {
			commands, stepCommandsCount, err := processStep(
				phase.Name, stepIndex, step, config, ctx, conditionEval)
			if err != nil {
				return nil, nil, err
			}

			plan.Commands = append(plan.Commands, commands...)
			phaseSummary.CommandCount += stepCommandsCount
			summary.TotalCommands += stepCommandsCount
			summary.TotalSteps++
		}

		if phaseSummary.StepCount > 0 || phaseSummary.CommandCount > 0 {
			summary.Phases = append(summary.Phases, phaseSummary)
		}
	}

	return &plan, &summary, nil
}

func processStep(
	phaseName string,
	stepIndex int,
	step Step,
	config *SetupConfig,
	ctx *RenderContext,
	conditionEval *ConditionEvaluator,
) ([]ExecutableCommand, int, error) {
	shouldExecute, err := conditionEval.ShouldExecuteStep(&step)
	if err != nil {
		return nil, 0, fmt.Errorf("error evaluating condition for step %s[%d]: %w", phaseName, stepIndex, err)
	}

	if !shouldExecute {
		return nil, 0, nil
	}

	stepCtx, err := buildStepContext(step, config, ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("error building step context for step %s[%d]: %w", phaseName, stepIndex, err)
	}

	if step.Run == "" && step.Print == "" {
		return nil, 0, fmt.Errorf("step %s[%d] must have either 'run' or 'print'", phaseName, stepIndex)
	}

	if step.Run != "" && step.Print != "" {
		return nil, 0, fmt.Errorf("step %s[%d] cannot have both 'run' and 'print'", phaseName, stepIndex)
	}

	retryPolicy := ParseRetryPolicy(step.Retries)
	var execCommands []ExecutableCommand

	if step.Print != "" {
		cmd := buildPrintCommand(step, stepCtx, phaseName, stepIndex)
		execCommands = append(execCommands, cmd)
	} else if step.Run != "" {
		cmd := buildRunCommand(step, stepCtx, phaseName, stepIndex)
		execCommands = append(execCommands, cmd)
	}

	for i := range execCommands {
		execCommands[i].Retries = retryPolicy
		execCommands[i].ContinueOnError = stepCtx.ContinueOnError
	}

	return execCommands, len(execCommands), nil
}

type StepContext struct {
	OS              string
	Arch            string
	Env             map[string]string
	WorkingDir      string
	Shell           string
	Timeout         time.Duration
	ContinueOnError bool
}

func buildStepContext(step Step, config *SetupConfig, ctx *RenderContext) (*StepContext, error) {
	timeout := ctx.DefaultTimeout
	if step.Timeout != "" {
		parsed, err := ParseDuration(step.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format: %w", err)
		}
		timeout = parsed
	} else if config.Defaults != nil && config.Defaults.Timeout != "" {
		parsed, err := ParseDuration(config.Defaults.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid default timeout format: %w", err)
		}
		timeout = parsed
	}

	workingDir := ctx.WorkingDir
	if step.CWD != "" {
		workingDir = filepath.Join(ctx.WorkingDir, step.CWD)
	} else if config.Defaults != nil && config.Defaults.CWD != "" {
		workingDir = filepath.Join(ctx.WorkingDir, config.Defaults.CWD)
	}

	shell := ctx.Shell
	if config.Defaults != nil && config.Defaults.Shell != "" {
		shell = config.Defaults.Shell
	}

	continueOnError := ctx.DefaultContinueOnError
	if step.ContinueOnError {
		continueOnError = step.ContinueOnError
	} else if config.Defaults != nil {
		continueOnError = config.Defaults.ContinueOnError
	}

	env := make(map[string]string)

	if ctx.GlobalEnv != nil {
		for k, v := range ctx.GlobalEnv {
			env[k] = v
		}
	}

	if config.Env != nil && config.Env.Vars != nil {
		for k, v := range config.Env.Vars {
			env[k] = v
		}
	}

	if step.Env != nil {
		for k, v := range step.Env {
			env[k] = v
		}
	}

	secrets := make(map[string]string)
	if ctx.Secrets != nil {
		for k, v := range ctx.Secrets {
			secrets[k] = v
		}
	}

	if step.EnvFromSecrets != nil {
		for _, secretName := range step.EnvFromSecrets {
			value, exists := secrets[secretName]
			if !exists && !step.OptionalSecrets {
				return nil, fmt.Errorf("required secret '%s' not found", secretName)
			}
			if exists {
				env[secretName] = value
			}
		}
	}

	return &StepContext{
		OS:              ctx.OS,
		Arch:            ctx.Arch,
		Env:             env,
		WorkingDir:      workingDir,
		Shell:           shell,
		Timeout:         timeout,
		ContinueOnError: continueOnError,
	}, nil
}

func buildRunCommand(step Step, stepCtx *StepContext, phaseName string, stepIndex int) ExecutableCommand {
	shell, args := determineShellCommand(stepCtx.Shell, step.Run)

	return ExecutableCommand{
		Phase:           phaseName,
		StepName:        step.Name,
		StepIndex:       stepIndex,
		Command:         shell,
		Args:            args,
		Env:             stepCtx.Env,
		WorkingDir:      stepCtx.WorkingDir,
		Timeout:         stepCtx.Timeout,
		Description:     fmt.Sprintf("Run: %s", step.Run),
		ContinueOnError: stepCtx.ContinueOnError,
	}
}

func buildPrintCommand(step Step, stepCtx *StepContext, phaseName string, stepIndex int) ExecutableCommand {
	var command string
	var args []string

	const (
		powershellCmd = "powershell"
		echoCmd       = "echo"
	)

	if stepCtx.OS == osWindows {
		command = powershellCmd
		args = []string{"-Command", fmt.Sprintf("Write-Host '%s'", strings.ReplaceAll(step.Print, "'", "''"))}
	} else {
		command = echoCmd
		args = []string{step.Print}
	}

	return ExecutableCommand{
		Phase:           phaseName,
		StepName:        step.Name,
		StepIndex:       stepIndex,
		Command:         command,
		Args:            args,
		Env:             stepCtx.Env,
		WorkingDir:      stepCtx.WorkingDir,
		Timeout:         stepCtx.Timeout,
		Description:     fmt.Sprintf("Print: %s", step.Print),
		ContinueOnError: stepCtx.ContinueOnError,
	}
}

func determineShellCommand(shell string, command string) (string, []string) {
	const (
		powershellCmd = "powershell"
		bashCmd       = "/bin/bash"
	)

	if shell == "auto" || shell == "" {
		if runtime.GOOS == goOSWindows {
			return powershellCmd, []string{"-Command", command}
		}
		return bashCmd, []string{"-c", command}
	}

	if strings.Contains(shell, powershellCmd) || strings.Contains(shell, "pwsh") {
		return shell, []string{"-Command", command}
	}

	return shell, []string{"-c", command}
}

func mergeEnv(envs ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, env := range envs {
		for k, v := range env {
			result[k] = v
		}
	}
	return result
}
