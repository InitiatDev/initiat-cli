package actions

import "fmt"

type PrintAction struct {
	*BaseAction
	message string
}

func NewPrintAction(message string) *PrintAction {
	return &PrintAction{
		BaseAction: NewBaseAction(ActionTypePrint),
		message:    message,
	}
}

func (a *PrintAction) Render(ctx *ActionContext) ([]Command, error) {
	cmd := Command{
		Command:     "echo",
		Args:        []string{a.message},
		Env:         ctx.Env,
		WorkingDir:  ctx.WorkingDir,
		Timeout:     ctx.Timeout,
		Description: fmt.Sprintf("Print: %s", a.message),
	}

	return []Command{cmd}, nil
}

func (a *PrintAction) Validate() error {
	return nil
}
