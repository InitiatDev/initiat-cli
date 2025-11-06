package actions

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/registry"
)

type EnsurePackageManagerAction struct {
	*BaseAction
	packageManagerType string
	pkgRegistry        *registry.PackageManagerRegistry
	resolvedType       string
	resolvedManager    registry.SystemPackageManager
	resolvedOS         string
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

	managerType, err := a.resolveManagerType(ctx.OS)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsurePackageManager, "failed to resolve package manager type", err)
	}

	installed, err := a.isManagerInstalled(managerType)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsurePackageManager, "failed to check if package manager is installed", err)
	}

	if installed {
		if a.resolvedManager == nil {
			return nil, NewActionError(ActionTypeEnsurePackageManager, "package manager not resolved", nil)
		}

		checkCmd := a.resolvedManager.CheckInstalledCommand(managerType)
		return []Command{
			{
				Command:     checkCmd.Command,
				Args:        checkCmd.Args,
				Env:         ctx.Env,
				WorkingDir:  ctx.WorkingDir,
				Timeout:     ctx.Timeout,
				Description: checkCmd.Description,
			},
		}, nil
	}

	if a.resolvedManager == nil {
		return nil, NewActionError(ActionTypeEnsurePackageManager, "package manager not resolved", nil)
	}

	installCmd := a.resolvedManager.InstallSelfCommand()
	return []Command{
		{
			Command:     installCmd.Command,
			Args:        installCmd.Args,
			Env:         ctx.Env,
			WorkingDir:  ctx.WorkingDir,
			Timeout:     ctx.Timeout,
			Description: installCmd.Description,
		},
	}, nil
}

func (a *EnsurePackageManagerAction) resolveManagerType(os string) (string, error) {
	if a.resolvedType != "" && a.resolvedOS == os {
		return a.resolvedType, nil
	}

	if a.packageManagerType != "auto" {
		pkgManager := a.pkgRegistry.FindByName(a.packageManagerType)
		if pkgManager != nil {
			if sysMgr, ok := pkgManager.(registry.SystemPackageManager); ok {
				a.resolvedType = sysMgr.Name()
				a.resolvedManager = sysMgr
				a.resolvedOS = os
				return a.resolvedType, nil
			}
			return "", fmt.Errorf("package manager %s is not a system package manager", a.packageManagerType)
		}

		a.resolvedType = a.packageManagerType
		a.resolvedOS = os
		return a.resolvedType, nil
	}

	pkgManager := a.findSystemPackageManager(os)
	if pkgManager == nil {
		return "", fmt.Errorf("no suitable package manager found for OS: %s", os)
	}

	a.resolvedType = pkgManager.Name()
	a.resolvedManager = pkgManager
	a.resolvedOS = os
	return a.resolvedType, nil
}

func (a *EnsurePackageManagerAction) Validate() error {
	if strings.TrimSpace(a.packageManagerType) == "" {
		return NewActionError(ActionTypeEnsurePackageManager, "package manager type cannot be empty", nil)
	}

	validTypes := []string{"auto", "brew", "apt", "yum", "dnf", "pacman", "choco", "scoop", "winget", "asdf"}
	if !contains(validTypes, a.packageManagerType) {
		return NewActionError(
			ActionTypeEnsurePackageManager,
			fmt.Sprintf("unsupported package manager type: %s", a.packageManagerType),
			nil,
		)
	}

	return nil
}

func (a *EnsurePackageManagerAction) findSystemPackageManager(os string) registry.SystemPackageManager {
	return a.pkgRegistry.FindSystemPackageManager(os)
}

func (a *EnsurePackageManagerAction) IsInstalled(ctx *ActionContext) (bool, error) {
	managerType, err := a.resolveManagerType(ctx.OS)
	if err != nil {
		return false, err
	}

	return a.isManagerInstalled(managerType)
}

func (a *EnsurePackageManagerAction) isManagerInstalled(managerType string) (bool, error) {
	_, err := exec.LookPath(managerType)
	return err == nil, nil
}
