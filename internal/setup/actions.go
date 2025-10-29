package setup

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions"
)

type ActionRegistry struct {
	actions map[actions.ActionType]actions.Action
}

func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[actions.ActionType]actions.Action),
	}
}

func (r *ActionRegistry) Register(action actions.Action) {
	r.actions[action.Type()] = action
}

func (r *ActionRegistry) Get(actionType actions.ActionType) (actions.Action, bool) {
	action, exists := r.actions[actionType]
	return action, exists
}

func (r *ActionRegistry) List() []actions.ActionType {
	var types []actions.ActionType
	for actionType := range r.actions {
		types = append(types, actionType)
	}
	return types
}

func (r *ActionRegistry) RenderStep(step *Step, ctx *actions.ActionContext) ([]actions.Command, error) {
	action, err := r.getActionFromStep(step)
	if err != nil {
		return nil, err
	}

	if err := action.Validate(); err != nil {
		return nil, actions.NewActionError(action.Type(), "validation failed", err)
	}

	return action.Render(ctx)
}

func (r *ActionRegistry) getActionFromStep(step *Step) (actions.Action, error) {
	var actionType actions.ActionType

	switch {
	case step.Run != "":
		actionType = actions.ActionTypeRun
	case step.Print != "":
		actionType = actions.ActionTypePrint
	case step.EnsurePackageManager != nil:
		actionType = actions.ActionTypeEnsurePackageManager
	case step.EnsureTool != nil:
		actionType = actions.ActionTypeEnsureTool
	case step.EnsureRuntime != nil:
		actionType = actions.ActionTypeEnsureRuntime
	case step.EnsureDatabase != nil:
		actionType = actions.ActionTypeEnsureDatabase
	case step.AssertCommand != "":
		actionType = actions.ActionTypeAssertCommand
	case step.AssertHTTP != nil:
		actionType = actions.ActionTypeAssertHTTP
	default:
		return nil, fmt.Errorf("no action found in step")
	}

	action, exists := r.Get(actionType)
	if !exists {
		return nil, fmt.Errorf("action type %s not registered", actionType)
	}

	return action, nil
}
