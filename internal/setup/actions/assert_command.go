package actions

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type AssertCommandAction struct {
	*BaseAction
	command string
}

func NewAssertCommandAction(command string) *AssertCommandAction {
	return &AssertCommandAction{
		BaseAction: NewBaseAction(ActionTypeAssertCommand),
		command:    command,
	}
}

func (a *AssertCommandAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.command) == "" {
		return nil, NewActionError(ActionTypeAssertCommand, "command cannot be empty", nil)
	}

	parts := strings.Fields(a.command)
	if len(parts) == 0 {
		return nil, NewActionError(ActionTypeAssertCommand, "command cannot be empty", nil)
	}

	cmd := Command{
		Command:     parts[0],
		Args:        parts[1:],
		Env:         ctx.Env,
		WorkingDir:  ctx.WorkingDir,
		Timeout:     ctx.Timeout,
		Description: fmt.Sprintf("Assert: %s", a.command),
	}

	return []Command{cmd}, nil
}

func (a *AssertCommandAction) Validate() error {
	if strings.TrimSpace(a.command) == "" {
		return NewActionError(ActionTypeAssertCommand, "command cannot be empty", nil)
	}
	return nil
}

func (a *AssertCommandAction) Execute(ctx *ActionContext) error {
	parts := strings.Fields(a.command)
	if len(parts) == 0 {
		return NewActionError(ActionTypeAssertCommand, "command cannot be empty", nil)
	}

	cmd := exec.Command(parts[0], parts[1:]...) // #nosec G204
	cmd.Env = os.Environ()
	for k, v := range ctx.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Dir = ctx.WorkingDir

	if ctx.Timeout > 0 {
		timeout := time.After(ctx.Timeout)
		done := make(chan error, 1)

		go func() {
			done <- cmd.Run()
		}()

		select {
		case err := <-done:
			if err != nil {
				return NewActionError(ActionTypeAssertCommand, fmt.Sprintf("assertion failed: %s", a.command), err)
			}
		case <-timeout:
			return NewActionError(
				ActionTypeAssertCommand,
				fmt.Sprintf("assertion timed out: %s", a.command),
				fmt.Errorf("timeout after %v", ctx.Timeout),
			)
		}
	} else {
		if err := cmd.Run(); err != nil {
			return NewActionError(ActionTypeAssertCommand, fmt.Sprintf("assertion failed: %s", a.command), err)
		}
	}

	return nil
}
