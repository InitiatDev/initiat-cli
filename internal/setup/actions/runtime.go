package actions

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/registry"
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
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

	if !a.pkgRegistry.HasAvailableManagers(ctx.OS) {
		return nil, NewActionError(ActionTypeEnsureRuntime,
			fmt.Sprintf("no package manager available for %s", ctx.OS), nil)
	}

	installed, err := a.IsInstalled(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsureRuntime, "failed to check if runtime is installed", err)
	}

	if installed {
		return []Command{
			{
				Command:     a.runtimeName,
				Args:        []string{"version"},
				Env:         ctx.Env,
				WorkingDir:  ctx.WorkingDir,
				Timeout:     ctx.Timeout,
				Description: fmt.Sprintf("Verify %s is installed", a.runtimeName),
			},
		}, nil
	}

	commands, err := a.getInstallCommands(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsureRuntime, "failed to generate install commands", err)
	}

	return wrapCommandsWithContext(commands, ctx), nil
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

func (r RuntimeCommand) GetCommand() string     { return r.Command }
func (r RuntimeCommand) GetArgs() []string      { return r.Args }
func (r RuntimeCommand) GetDescription() string { return r.Description }

func (a *EnsureRuntimeAction) getInstallCommands(ctx *ActionContext) ([]RuntimeCommand, error) {
	if a.manager != nil && a.manager.Asdf {
		runtimeManager := a.pkgRegistry.FindRuntimePackageManager(ctx.OS)
		if runtimeManager != nil {
			var commands []RuntimeCommand
			installCommands := runtimeManager.GetInstallCommands(a.runtimeName, a.version)
			for _, cmd := range installCommands {
				commands = append(commands, RuntimeCommand{
					Command:     cmd.Command,
					Args:        cmd.Args,
					Description: cmd.Description,
				})
			}
			return commands, nil
		}
	}

	systemManager := a.pkgRegistry.FindSystemPackageManager(ctx.OS)
	if systemManager != nil {
		if sysManager, ok := systemManager.(interface {
			InstallCommand(pkg, version string) types.Command
		}); ok {
			installCmd := sysManager.InstallCommand(a.runtimeName, a.version)
			return []RuntimeCommand{
				{
					Command:     installCmd.Command,
					Args:        installCmd.Args,
					Description: installCmd.Description,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("no suitable install method found for %s on %s", a.runtimeName, ctx.OS)
}

// IsInstalled checks if the runtime is already installed
func (a *EnsureRuntimeAction) IsInstalled(ctx *ActionContext) (bool, error) {
	_, err := exec.LookPath(a.runtimeName)
	return err == nil, nil
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
