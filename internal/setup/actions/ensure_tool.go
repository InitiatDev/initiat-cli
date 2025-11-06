package actions

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/registry"
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

type EnsureToolAction struct {
	*BaseAction
	toolName      string
	version       string
	installConfig *types.ToolInstallConfig
	pkgRegistry   *registry.PackageManagerRegistry
}

func NewEnsureToolAction(toolName, version string, installConfig *types.ToolInstallConfig) *EnsureToolAction {
	return &EnsureToolAction{
		BaseAction:    NewBaseAction(ActionTypeEnsureTool),
		toolName:      toolName,
		version:       version,
		installConfig: installConfig,
		pkgRegistry:   registry.NewPackageManagerRegistry(),
	}
}

func (a *EnsureToolAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.toolName) == "" {
		return nil, NewActionError(ActionTypeEnsureTool, "tool name cannot be empty", nil)
	}

	installed, err := a.IsInstalled(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsureTool, "failed to check if tool is installed", err)
	}

	if installed {
		return []Command{
			{
				Command:     "which",
				Args:        []string{a.toolName},
				Env:         ctx.Env,
				WorkingDir:  ctx.WorkingDir,
				Timeout:     ctx.Timeout,
				Description: fmt.Sprintf("Verify %s is installed", a.toolName),
			},
		}, nil
	}

	commands, err := a.getInstallCommands(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsureTool, "failed to generate install commands", err)
	}

	return wrapCommandsWithContext(commands, ctx), nil
}

func (a *EnsureToolAction) Validate() error {
	if strings.TrimSpace(a.toolName) == "" {
		return NewActionError(ActionTypeEnsureTool, "tool name cannot be empty", nil)
	}

	if a.installConfig == nil {
		return NewActionError(ActionTypeEnsureTool, "install configuration is required", nil)
	}

	hasInstallMethod := a.installConfig.Brew != nil ||
		a.installConfig.Apt != nil ||
		a.installConfig.Choco != nil ||
		a.installConfig.FallbackURL != ""

	if !hasInstallMethod {
		return NewActionError(ActionTypeEnsureTool, "at least one install method must be specified", nil)
	}

	return nil
}

type ToolCommand struct {
	Command     string
	Args        []string
	Description string
}

func (t ToolCommand) GetCommand() string     { return t.Command }
func (t ToolCommand) GetArgs() []string      { return t.Args }
func (t ToolCommand) GetDescription() string { return t.Description }

func (a *EnsureToolAction) getInstallCommands(ctx *ActionContext) ([]ToolCommand, error) {
	if a.installConfig != nil {
		manager := a.findManagerWithConfig(ctx.OS)
		if manager != nil {
			return a.getCommandsFromConfig(manager)
		}
	}

	systemManager := a.pkgRegistry.FindSystemPackageManager(ctx.OS)
	if systemManager != nil {
		installCmd := systemManager.InstallCommand(a.toolName, a.version)
		return []ToolCommand{
			{
				Command:     installCmd.Command,
				Args:        installCmd.Args,
				Description: installCmd.Description,
			},
		}, nil
	}

	if a.installConfig != nil && a.installConfig.FallbackURL != "" {
		return a.getFallbackCommands(ctx)
	}

	return nil, fmt.Errorf("no suitable install method found for %s on %s", a.toolName, ctx.OS)
}

func (a *EnsureToolAction) findManagerWithConfig(os string) registry.SystemPackageManager {
	systemManager := a.pkgRegistry.FindSystemPackageManager(os)
	if systemManager == nil {
		return nil
	}

	_, ok := systemManager.ExtractToolInstallCommands(a.installConfig)
	if ok {
		return systemManager
	}

	return nil
}

func (a *EnsureToolAction) getCommandsFromConfig(
	pkgManager registry.SystemPackageManager,
) ([]ToolCommand, error) {
	commands, ok := pkgManager.ExtractToolInstallCommands(a.installConfig)
	if !ok {
		return nil, fmt.Errorf("install config not properly specified")
	}

	var result []ToolCommand
	for _, cmd := range commands {
		result = append(result, ToolCommand{
			Command:     cmd.Command,
			Args:        cmd.Args,
			Description: cmd.Description,
		})
	}

	return result, nil
}

// getFallbackCommands generates fallback installation commands for direct download
func (a *EnsureToolAction) getFallbackCommands(_ *ActionContext) ([]ToolCommand, error) {
	if a.installConfig.FallbackURL == "" {
		return nil, fmt.Errorf("no fallback URL provided")
	}

	commands := []ToolCommand{
		{
			Command:     "curl",
			Args:        []string{"-L", "-o", fmt.Sprintf("/tmp/%s", a.toolName), a.installConfig.FallbackURL},
			Description: fmt.Sprintf("Download %s from fallback URL", a.toolName),
		},
		{
			Command:     "chmod",
			Args:        []string{"+x", fmt.Sprintf("/tmp/%s", a.toolName)},
			Description: fmt.Sprintf("Make %s executable", a.toolName),
		},
		{
			Command:     "sudo",
			Args:        []string{"mv", fmt.Sprintf("/tmp/%s", a.toolName), "/usr/local/bin/"},
			Description: fmt.Sprintf("Install %s to /usr/local/bin", a.toolName),
		},
	}

	return commands, nil
}

// IsInstalled checks if the tool is already installed
func (a *EnsureToolAction) IsInstalled(ctx *ActionContext) (bool, error) {
	_, err := exec.LookPath(a.toolName)
	return err == nil, nil
}

func wrapCommandsWithContext[T interface {
	GetCommand() string
	GetArgs() []string
	GetDescription() string
}](commands []T, ctx *ActionContext) []Command {
	result := make([]Command, len(commands))
	for i, cmd := range commands {
		result[i] = Command{
			Command:     cmd.GetCommand(),
			Args:        cmd.GetArgs(),
			Env:         ctx.Env,
			WorkingDir:  ctx.WorkingDir,
			Timeout:     ctx.Timeout,
			Description: cmd.GetDescription(),
		}
	}
	return result
}
