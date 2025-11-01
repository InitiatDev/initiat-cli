package setup

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions"
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
	if err := Validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	if err := validateMatrix(config, ctx); err != nil {
		return nil, err
	}

	conditionEval := NewConditionEvaluator(ctx.OS, ctx.Arch, mergeEnv(ctx.GlobalEnv, ctx.Secrets))
	actionFactory := NewActionFactory()

	plan, summary, err := processPhases(config, ctx, conditionEval, actionFactory)
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
	actionFactory *ActionFactory,
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
				phase.Name, stepIndex, step, config, ctx, conditionEval, actionFactory)
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
	actionFactory *ActionFactory,
) ([]ExecutableCommand, int, error) {
	shouldExecute, err := conditionEval.ShouldExecuteStep(&step)
	if err != nil {
		return nil, 0, fmt.Errorf("error evaluating condition for step %s[%d]: %w", phaseName, stepIndex, err)
	}

	if !shouldExecute {
		return nil, 0, nil
	}

	stepCtx, err := buildActionContext(step, config, ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("error building action context for step %s[%d]: %w", phaseName, stepIndex, err)
	}

	action, err := actionFactory.BuildFromStep(&step)
	if err != nil {
		return nil, 0, fmt.Errorf("error building action for step %s[%d]: %w", phaseName, stepIndex, err)
	}

	if err := action.Validate(); err != nil {
		return nil, 0, fmt.Errorf("action validation failed for step %s[%d]: %w", phaseName, stepIndex, err)
	}

	commands, err := action.Render(stepCtx)
	if err != nil {
		return nil, 0, fmt.Errorf("action rendering failed for step %s[%d]: %w", phaseName, stepIndex, err)
	}

	retryPolicy := buildRetryPolicy(step.Retries)
	var execCommands []ExecutableCommand

	for _, cmd := range commands {
		execCmd := ExecutableCommand{
			Phase:           phaseName,
			StepName:        step.Name,
			StepIndex:       stepIndex,
			Command:         cmd.Command,
			Args:            cmd.Args,
			Env:             cmd.Env,
			WorkingDir:      cmd.WorkingDir,
			Timeout:         cmd.Timeout,
			Description:     cmd.Description,
			Retries:         retryPolicy,
			ContinueOnError: stepCtx.ContinueOnError,
		}

		execCommands = append(execCommands, execCmd)
	}

	return execCommands, len(execCommands), nil
}

func buildActionContext(step Step, config *SetupConfig, ctx *RenderContext) (*actions.ActionContext, error) {
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

	return &actions.ActionContext{
		OS:              ctx.OS,
		Arch:            ctx.Arch,
		Env:             env,
		Secrets:         secrets,
		WorkingDir:      workingDir,
		Shell:           shell,
		Timeout:         timeout,
		ContinueOnError: continueOnError,
	}, nil
}

func buildRetryPolicy(retries *Retries) *RetryPolicy {
	return ParseRetryPolicy(retries)
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
