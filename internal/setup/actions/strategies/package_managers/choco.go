package package_managers

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// ChocoPackageManager implements package management via Chocolatey
type ChocoPackageManager struct{}

func (p *ChocoPackageManager) Name() string              { return "choco" }
func (p *ChocoPackageManager) SupportsOS(os string) bool { return os == types.OSWindows }

func (p *ChocoPackageManager) InstallCommand(pkg, version string) types.Command {
	args := []string{"install", "-y", pkg}
	if version != "" {
		args = append(args, "--version", version)
	}
	return types.Command{
		Command:     "choco",
		Args:        args,
		Description: fmt.Sprintf("Install %s %s via Chocolatey", pkg, version),
	}
}

func (p *ChocoPackageManager) CheckInstalledCommand(pkg string) types.Command {
	return types.Command{
		Command:     "where",
		Args:        []string{pkg},
		Description: fmt.Sprintf("Check if %s is installed", pkg),
	}
}

func (p *ChocoPackageManager) GetInstallCommands(pkg, version string) []types.Command {
	return []types.Command{
		p.CheckInstalledCommand(pkg),
		p.InstallCommand(pkg, version),
	}
}
