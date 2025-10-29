package package_managers

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// AptPackageManager implements package management via apt
type AptPackageManager struct{}

func (p *AptPackageManager) Name() string              { return "apt" }
func (p *AptPackageManager) SupportsOS(os string) bool { return os == types.OSLinux }

func (p *AptPackageManager) InstallCommand(pkg, version string) types.Command {
	args := []string{"apt", "install", "-y", pkg}
	return types.Command{
		Command:     "sudo",
		Args:        args,
		Description: fmt.Sprintf("Install %s via apt", pkg),
	}
}

func (p *AptPackageManager) CheckInstalledCommand(pkg string) types.Command {
	return types.Command{
		Command:     "which",
		Args:        []string{pkg},
		Description: fmt.Sprintf("Check if %s is installed", pkg),
	}
}

func (p *AptPackageManager) GetInstallCommands(pkg, version string) []types.Command {
	return []types.Command{
		p.CheckInstalledCommand(pkg),
		p.InstallCommand(pkg, version),
	}
}
