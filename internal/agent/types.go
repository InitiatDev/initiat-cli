package agent

import "context"

type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLM interface {
	Complete(ctx context.Context, req *CompleteRequest) (*CompleteResponse, error)
}

type CompleteRequest struct {
	Model    string
	System   string
	Messages []Message
}

type CompleteResponse struct {
	Text string
}

type ProposedActionType string

const (
	ActionStop       ProposedActionType = "stop"
	ActionRunCommand ProposedActionType = "run_command"
	ActionEditFiles  ProposedActionType = "edit_files"
	ActionAskUser    ProposedActionType = "ask_user"
)

type ProposedAction struct {
	Type ProposedActionType `json:"type"`

	Reason string `json:"reason,omitempty"`

	Command string            `json:"command,omitempty"`
	CWD     string            `json:"cwd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	Edits []FileEdit `json:"edits,omitempty"`

	Prompt string `json:"prompt,omitempty"`
}

type FileEdit struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type Decision struct {
	Explanation string           `json:"explanation"`
	Actions     []ProposedAction `json:"actions"`
}
