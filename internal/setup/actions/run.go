package actions

import (
	"fmt"
	"strings"
)

type RunAction struct {
	*BaseAction
	command string
}

func NewRunAction(command string) *RunAction {
	return &RunAction{
		BaseAction: NewBaseAction(ActionTypeRun),
		command:    command,
	}
}

func (a *RunAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.command) == "" {
		return nil, NewActionError(ActionTypeRun, "command cannot be empty", nil)
	}

	parts := strings.Fields(a.command)
	if len(parts) == 0 {
		return nil, NewActionError(ActionTypeRun, "command cannot be empty", nil)
	}

	cmd := Command{
		Command:     parts[0],
		Args:        parts[1:],
		Env:         ctx.Env,
		WorkingDir:  ctx.WorkingDir,
		Timeout:     ctx.Timeout,
		Description: fmt.Sprintf("Run: %s", a.command),
	}

	return []Command{cmd}, nil
}

func (a *RunAction) Validate() error {
	if strings.TrimSpace(a.command) == "" {
		return NewActionError(ActionTypeRun, "command cannot be empty", nil)
	}
	return nil
}
