package agent

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/prompt"
)

type Approver interface {
	Approve(action ProposedAction) (bool, error)
}

type PromptApprover struct{}

func (a *PromptApprover) Approve(action ProposedAction) (bool, error) {
	switch action.Type {
	case ActionRunCommand:
		return prompt.PromptYesNo(fmt.Sprintf("Agent wants to run: %s. Approve?", action.Command))
	case ActionEditFiles:
		return prompt.PromptYesNo("Agent wants to edit files. Approve?")
	case ActionAskUser:
		return true, nil
	case ActionStop:
		return true, nil
	default:
		return false, fmt.Errorf("unknown action type: %s", action.Type)
	}
}
