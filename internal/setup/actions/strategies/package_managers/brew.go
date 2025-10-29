package package_managers

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// BrewPackageManager implements package management via Homebrew
type BrewPackageManager struct{}

func (p *BrewPackageManager) Name() string { return "brew" }
func (p *BrewPackageManager) SupportsOS(os string) bool {
	return os == types.OSMacOS || os == types.OSDarwin
}

func (p *BrewPackageManager) InstallCommand(pkg, version string) types.Command {
	args := []string{"install"}
	if version != "" {
		args = append(args, fmt.Sprintf("%s@%s", pkg, version))
	} else {
		args = append(args, pkg)
	}
	return types.Command{
		Command:     "brew",
		Args:        args,
		Description: fmt.Sprintf("Install %s %s via Homebrew", pkg, version),
	}
}

func (p *BrewPackageManager) CheckInstalledCommand(pkg string) types.Command {
	return types.Command{
		Command:     "which",
		Args:        []string{pkg},
		Description: fmt.Sprintf("Check if %s is installed", pkg),
	}
}

func (p *BrewPackageManager) GetInstallCommands(pkg, version string) []types.Command {
	return []types.Command{
		p.CheckInstalledCommand(pkg),
		p.InstallCommand(pkg, version),
	}
}
