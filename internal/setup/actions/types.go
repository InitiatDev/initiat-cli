package actions

import (
	"fmt"
	"time"
)

type ActionType string

const (
	ActionTypeRun                  ActionType = "run"
	ActionTypePrint                ActionType = "print"
	ActionTypeEnsurePackageManager ActionType = "ensure_package_manager"
	ActionTypeEnsureTool           ActionType = "ensure_tool"
	ActionTypeEnsureRuntime        ActionType = "ensure_runtime"
	ActionTypeEnsureDatabase       ActionType = "ensure_database"
	ActionTypeAssertCommand        ActionType = "assert_command"
	ActionTypeAssertHTTP           ActionType = "assert_http"
)

type Action interface {
	Type() ActionType
	Render(ctx *ActionContext) ([]Command, error)
	Validate() error
}

type ActionContext struct {
	OS              string
	Arch            string
	Env             map[string]string
	Secrets         map[string]string
	WorkingDir      string
	Shell           string
	Timeout         time.Duration
	ContinueOnError bool
}

type Command struct {
	Command     string
	Args        []string
	Env         map[string]string
	WorkingDir  string
	Timeout     time.Duration
	Description string
}

type ActionError struct {
	Type    ActionType
	Message string
	Err     error
}

func (e *ActionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s action error: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s action error: %s", e.Type, e.Message)
}

func (e *ActionError) Unwrap() error {
	return e.Err
}

func NewActionError(actionType ActionType, message string, err error) *ActionError {
	return &ActionError{
		Type:    actionType,
		Message: message,
		Err:     err,
	}
}
