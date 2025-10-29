package actions

import (
	"fmt"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/registry"
)

type EnsurePackageManagerAction struct {
	*BaseAction
	packageManagerType string
	pkgRegistry        *registry.PackageManagerRegistry
}

func NewEnsurePackageManagerAction(packageManagerType string) *EnsurePackageManagerAction {
	return &EnsurePackageManagerAction{
		BaseAction:         NewBaseAction(ActionTypeEnsurePackageManager),
		packageManagerType: packageManagerType,
		pkgRegistry:        registry.NewPackageManagerRegistry(),
	}
}

func (a *EnsurePackageManagerAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.packageManagerType) == "" {
		return nil, NewActionError(ActionTypeEnsurePackageManager, "package manager type cannot be empty", nil)
	}

	commands, err := a.getInstallCommands(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsurePackageManager, "failed to generate install commands", err)
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

func (a *EnsurePackageManagerAction) Validate() error {
	if strings.TrimSpace(a.packageManagerType) == "" {
		return NewActionError(ActionTypeEnsurePackageManager, "package manager type cannot be empty", nil)
	}

	validTypes := []string{"brew", "apt", "yum", "dnf", "pacman", "choco", "scoop", "winget", "asdf"}
	if !contains(validTypes, a.packageManagerType) {
		return NewActionError(
			ActionTypeEnsurePackageManager,
			fmt.Sprintf("unsupported package manager type: %s", a.packageManagerType),
			nil,
		)
	}

	return nil
}

type PackageManagerCommand struct {
	Command     string
	Args        []string
	Description string
}

// getInstallCommands dispatches to the appropriate package manager implementation
func (a *EnsurePackageManagerAction) getInstallCommands(ctx *ActionContext) ([]PackageManagerCommand, error) {
	pkgManager := a.pkgRegistry.FindManager(ctx.OS)
	if pkgManager == nil {
		return nil, fmt.Errorf("no suitable package manager found for %s on %s", a.packageManagerType, ctx.OS)
	}

	installCmd := pkgManager.InstallCommand(a.packageManagerType, "")
	commands := []PackageManagerCommand{
		{
			Command:     installCmd.Command,
			Args:        installCmd.Args,
			Description: installCmd.Description,
		},
	}

	return commands, nil
}

// IsInstalled checks if the package manager is already installed
func (a *EnsurePackageManagerAction) IsInstalled(ctx *ActionContext) (bool, error) {
	pkgManager := a.pkgRegistry.FindManager(ctx.OS)
	if pkgManager == nil {
		return false, fmt.Errorf("no suitable package manager found for %s on %s", a.packageManagerType, ctx.OS)
	}

	_ = pkgManager.CheckInstalledCommand(a.packageManagerType)
	return true, nil
}
