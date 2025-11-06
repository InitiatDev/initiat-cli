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

func (p *AptPackageManager) InstallSelfCommand() types.Command {
	return types.Command{
		Command:     "which",
		Args:        []string{"apt"},
		Description: "Check if apt is available (usually pre-installed on Linux)",
	}
}

func (p *AptPackageManager) ExtractToolInstallCommands(config interface{}) ([]types.Command, bool) {
	cfg, ok := config.(*types.ToolInstallConfig)
	if !ok || cfg.Apt == nil || len(cfg.Apt.Packages) == 0 {
		return nil, false
	}

	var commands []types.Command
	if cfg.Apt.Update {
		commands = append(commands, types.Command{
			Command:     "sudo",
			Args:        []string{"apt", "update"},
			Description: "Update apt package list",
		})
	}

	for _, pkg := range cfg.Apt.Packages {
		commands = append(commands, p.InstallCommand(pkg, ""))
	}

	return commands, true
}
