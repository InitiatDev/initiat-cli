package actions

import (
	"fmt"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/registry"
)

type EnsureRuntimeAction struct {
	*BaseAction
	runtimeName string
	version     string
	manager     *RuntimeManager
	pkgRegistry *registry.PackageManagerRegistry
}

type RuntimeManager struct {
	Asdf bool `yaml:"asdf,omitempty" json:"asdf,omitempty"`
}

func NewEnsureRuntimeAction(
	runtimeName, version string,
	manager *RuntimeManager,
) *EnsureRuntimeAction {
	return &EnsureRuntimeAction{
		BaseAction:  NewBaseAction(ActionTypeEnsureRuntime),
		runtimeName: runtimeName,
		version:     version,
		manager:     manager,
		pkgRegistry: registry.NewPackageManagerRegistry(),
	}
}

func (a *EnsureRuntimeAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.runtimeName) == "" {
		return nil, NewActionError(ActionTypeEnsureRuntime, "runtime name cannot be empty", nil)
	}

	// Check if any package manager is available for this OS
	if !a.pkgRegistry.HasAvailableManagers(ctx.OS) {
		return nil, NewActionError(ActionTypeEnsureRuntime,
			fmt.Sprintf("no package manager available for %s", ctx.OS), nil)
	}

	commands, err := a.getInstallCommands(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsureRuntime, "failed to generate install commands", err)
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

func (a *EnsureRuntimeAction) Validate() error {
	if strings.TrimSpace(a.runtimeName) == "" {
		return NewActionError(ActionTypeEnsureRuntime, "runtime name cannot be empty", nil)
	}

	// Validate runtime name
	validRuntimes := []string{
		RuntimeNode, RuntimePython, RuntimeGo, RuntimeElixir,
		RuntimeErlang, RuntimeJava, RuntimeRust, RuntimeDotnet,
	}
	if !contains(validRuntimes, a.runtimeName) {
		return NewActionError(
			ActionTypeEnsureRuntime,
			fmt.Sprintf("invalid runtime name '%s', must be one of: %s", a.runtimeName, strings.Join(validRuntimes, ", ")),
			nil,
		)
	}

	// Validate that at least one install method is configured
	if a.manager == nil {
		return NewActionError(ActionTypeEnsureRuntime, "at least one install method must be specified", nil)
	}

	return nil
}

type RuntimeCommand struct {
	Command     string
	Args        []string
	Description string
}

// getInstallCommands determines the best install method for the current platform
func (a *EnsureRuntimeAction) getInstallCommands(ctx *ActionContext) ([]RuntimeCommand, error) {
	manager := a.pkgRegistry.FindManager(ctx.OS)
	if manager == nil {
		return nil, fmt.Errorf("no suitable install method found for %s on %s", a.runtimeName, ctx.OS)
	}

	var commands []RuntimeCommand

	installCommands := manager.GetInstallCommands(a.runtimeName, a.version)
	for _, cmd := range installCommands {
		commands = append(commands, RuntimeCommand{
			Command:     cmd.Command,
			Args:        cmd.Args,
			Description: cmd.Description,
		})
	}

	return commands, nil
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
