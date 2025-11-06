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

func (p *BrewPackageManager) InstallSelfCommand() types.Command {
	return types.Command{
		Command:     "/bin/bash",
		Args:        []string{"-c", "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"},
		Description: "Install Homebrew",
	}
}

func (p *BrewPackageManager) ExtractToolInstallCommands(config interface{}) ([]types.Command, bool) {
	cfg, ok := config.(*types.ToolInstallConfig)
	if !ok || cfg.Brew == nil || cfg.Brew.Formula == "" {
		return nil, false
	}

	return []types.Command{p.InstallCommand(cfg.Brew.Formula, "")}, true
}
