package actions

import "fmt"

type BaseAction struct {
	actionType ActionType
}

func NewBaseAction(actionType ActionType) *BaseAction {
	return &BaseAction{
		actionType: actionType,
	}
}

func (a *BaseAction) Type() ActionType {
	return a.actionType
}

func (a *BaseAction) Render(ctx *ActionContext) ([]Command, error) {
	return nil, fmt.Errorf("BaseAction.Render not implemented")
}

func (a *BaseAction) Validate() error {
	return nil
}
