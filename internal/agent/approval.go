package agent

import (
	"fmt"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/prompt"
)

type Approver interface {
	Approve(action ProposedAction) (bool, error)
}

type PromptApprover struct{}

func (a *PromptApprover) Approve(action ProposedAction) (bool, error) {
	switch action.Type {
	case ActionRunCommand:
		assess := AssessSafety(action)
		msg := approvalMessage(action, assess, fmt.Sprintf("Run command:\n%s", action.Command))
		return prompt.PromptYesNo(msg)
	case ActionEditFiles:
		assess := AssessSafety(action)
		var b strings.Builder
		b.WriteString("Edit files:\n")
		for _, e := range action.Edits {
			b.WriteString("- ")
			b.WriteString(e.Path)
			b.WriteString("\n")
		}
		msg := approvalMessage(action, assess, strings.TrimSpace(b.String()))
		return prompt.PromptYesNo(msg)
	case ActionListFiles, ActionReadFiles:
		return true, nil
	case ActionAskUser:
		return true, nil
	case ActionStop:
		return true, nil
	default:
		return false, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func approvalMessage(action ProposedAction, assess SafetyAssessment, body string) string {
	effective := MaxDanger(action.Danger, assess.EffectiveDanger)

	var b strings.Builder
	b.WriteString("Agent proposal\n\n")
	b.WriteString("Summary: ")
	b.WriteString(action.Summary)
	b.WriteString("\nDanger: ")
	b.WriteString(string(effective))
	b.WriteString("\nDanger reason: ")
	b.WriteString(action.DangerReason)
	if strings.TrimSpace(action.Reason) != "" {
		b.WriteString("\nWhy: ")
		b.WriteString(action.Reason)
	}
	if len(assess.Signals) > 0 {
		b.WriteString("\nLocal safety signals: ")
		b.WriteString(strings.Join(assess.Signals, "; "))
	}
	b.WriteString("\n\n")
	b.WriteString(body)
	b.WriteString("\n\nApprove?")
	return b.String()
}
