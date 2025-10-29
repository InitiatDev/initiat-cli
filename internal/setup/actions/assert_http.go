package actions

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/registry"
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

type AssertHTTPAction struct {
	*BaseAction
	url          string
	expectStatus int
	retries      *types.Retries
	httpRegistry *registry.HTTPClientRegistry
}

func NewAssertHTTPAction(url string, expectStatus int, retries *types.Retries) *AssertHTTPAction {
	return &AssertHTTPAction{
		BaseAction:   NewBaseAction(ActionTypeAssertHTTP),
		url:          url,
		expectStatus: expectStatus,
		retries:      retries,
		httpRegistry: registry.NewHTTPClientRegistry(),
	}
}

func (a *AssertHTTPAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.url) == "" {
		return nil, NewActionError(ActionTypeAssertHTTP, "URL cannot be empty", nil)
	}

	commands, err := a.getHTTPCommands(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeAssertHTTP, "failed to generate HTTP commands", err)
	}

	var result []Command
	for _, cmd := range commands {
		result = append(result, Command{
			Command:     cmd.Command,
			Args:        cmd.Args,
			Env:         ctx.Env,
			WorkingDir:  ctx.WorkingDir,
			Timeout:     ctx.Timeout,
			Description: cmd.Description,
		})
	}

	return result, nil
}

func (a *AssertHTTPAction) Validate() error {
	if strings.TrimSpace(a.url) == "" {
		return NewActionError(ActionTypeAssertHTTP, "URL cannot be empty", nil)
	}

	if !strings.HasPrefix(a.url, "http://") && !strings.HasPrefix(a.url, "https://") {
		return NewActionError(ActionTypeAssertHTTP, "URL must start with http:// or https://", nil)
	}

	if a.expectStatus == 0 {
		a.expectStatus = 200
	}

	if a.expectStatus < 100 || a.expectStatus > 599 {
		return NewActionError(ActionTypeAssertHTTP, "expected status code must be between 100 and 599", nil)
	}

	return nil
}

// getHTTPCommands generates HTTP assertion commands
func (a *AssertHTTPAction) getHTTPCommands(ctx *ActionContext) ([]Command, error) {
	var commands []Command

	httpClient := a.httpRegistry.FindClient(ctx.OS)
	if httpClient == nil {
		return nil, fmt.Errorf("no suitable HTTP client found for %s", ctx.OS)
	}

	checkCmd := httpClient.CheckURLCommand(a.url, a.expectStatus, a.retries)
	commands = append(commands, Command{
		Command:     checkCmd.Command,
		Args:        checkCmd.Args,
		Description: checkCmd.Description,
	})

	return commands, nil
}

// CheckHTTP performs the actual HTTP check
func (a *AssertHTTPAction) CheckHTTP(ctx *ActionContext) (bool, error) {
	httpClient := a.httpRegistry.FindClient(ctx.OS)
	if httpClient == nil {
		return false, fmt.Errorf("no suitable HTTP client found for %s", ctx.OS)
	}

	checkCmd := httpClient.CheckURLCommand(a.url, a.expectStatus, a.retries)
	cmd := exec.Command(checkCmd.Command, checkCmd.Args...) // #nosec G204

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	statusCode := strings.TrimSpace(string(output))
	expectedStatus := fmt.Sprintf("%d", a.expectStatus)

	return statusCode == expectedStatus, nil
}
