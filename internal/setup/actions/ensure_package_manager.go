package actions

import (
	"fmt"
	"strings"
)

type EnsurePackageManagerAction struct {
	*BaseAction
	packageManagerType string
}

func NewEnsurePackageManagerAction(packageManagerType string) *EnsurePackageManagerAction {
	return &EnsurePackageManagerAction{
		BaseAction:         NewBaseAction(ActionTypeEnsurePackageManager),
		packageManagerType: packageManagerType,
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

	validTypes := []string{"brew", "apt", "yum", "dnf", "pacman", "choco", "scoop", "winget"}
	for _, validType := range validTypes {
		if a.packageManagerType == validType {
			return nil
		}
	}

	return NewActionError(
		ActionTypeEnsurePackageManager,
		fmt.Sprintf("unsupported package manager type: %s", a.packageManagerType),
		nil,
	)
}

type PackageManagerCommand struct {
	Command     string
	Args        []string
	Description string
}

// getInstallCommands dispatches to the appropriate package manager implementation
func (a *EnsurePackageManagerAction) getInstallCommands(ctx *ActionContext) ([]PackageManagerCommand, error) {
	os := strings.ToLower(ctx.OS)
	arch := strings.ToLower(ctx.Arch)

	switch a.packageManagerType {
	case "brew":
		return a.getBrewCommands(os, arch)
	case "apt":
		return a.getAptCommands(os)
	case "yum":
		return a.getYumCommands(os)
	case "dnf":
		return a.getDnfCommands(os)
	case "pacman":
		return a.getPacmanCommands(os)
	case "choco":
		return a.getChocoCommands(os)
	case "scoop":
		return a.getScoopCommands(os)
	case "winget":
		return a.getWingetCommands(os)
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", a.packageManagerType)
	}
}

// IsInstalled checks if the package manager is already installed
func (a *EnsurePackageManagerAction) IsInstalled(ctx *ActionContext) (bool, error) {
	switch a.packageManagerType {
	case "brew":
		return a.isBrewInstalled()
	case "apt":
		return a.isAptInstalled()
	case "yum":
		return a.isYumInstalled()
	case "dnf":
		return a.isDnfInstalled()
	case "pacman":
		return a.isPacmanInstalled()
	case "choco":
		return a.isChocoInstalled()
	case "scoop":
		return a.isScoopInstalled()
	case "winget":
		return a.isWingetInstalled()
	default:
		return false, fmt.Errorf("unsupported package manager: %s", a.packageManagerType)
	}
}
