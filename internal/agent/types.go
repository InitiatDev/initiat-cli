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
	ActionListFiles  ProposedActionType = "list_files"
	ActionReadFiles  ProposedActionType = "read_files"
)

type DangerLevel string

const (
	DangerSafe      DangerLevel = "safe"
	DangerCaution   DangerLevel = "caution"
	DangerDangerous DangerLevel = "dangerous"
)

type ProposedAction struct {
	Type ProposedActionType `json:"type"`

	Summary      string      `json:"summary,omitempty"`
	Danger       DangerLevel `json:"danger,omitempty"`
	DangerReason string      `json:"danger_reason,omitempty"`

	Reason string `json:"reason,omitempty"`

	Command string            `json:"command,omitempty"`
	CWD     string            `json:"cwd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	Edits []FileEdit `json:"edits,omitempty"`

	Prompt string `json:"prompt,omitempty"`

	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths,omitempty"`
}

type FileEdit struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type Decision struct {
	Explanation string           `json:"explanation"`
	Actions     []ProposedAction `json:"actions"`
}

type AppliedActionResult struct {
	Type    ProposedActionType `json:"type"`
	Summary string             `json:"summary"`
	OK      bool               `json:"ok"`
	Error   string             `json:"error,omitempty"`
	Output  string             `json:"output,omitempty"`
}

type ApplyResult struct {
	Results []AppliedActionResult `json:"results"`
}
