package package_managers

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// AsdfPackageManager implements package management via asdf
type AsdfPackageManager struct{}

func (p *AsdfPackageManager) Name() string              { return "asdf" }
func (p *AsdfPackageManager) SupportsOS(os string) bool { return true }

func (p *AsdfPackageManager) InstallCommand(pkg, version string) types.Command {
	args := []string{"install", pkg}
	if version != "" {
		args = append(args, version)
	}
	return types.Command{
		Command:     "asdf",
		Args:        args,
		Description: fmt.Sprintf("Install %s %s via asdf", pkg, version),
	}
}

func (p *AsdfPackageManager) CheckInstalledCommand(pkg string) types.Command {
	return types.Command{
		Command:     "asdf",
		Args:        []string{"list", pkg},
		Description: fmt.Sprintf("Check if %s is installed via asdf", pkg),
	}
}

func (p *AsdfPackageManager) GetInstallCommands(pkg, version string) []types.Command {
	commands := []types.Command{
		p.CheckInstalledCommand(pkg),
		p.InstallCommand(pkg, version),
	}

	if version != "" {
		globalCmd := types.Command{
			Command:     "asdf",
			Args:        []string{"global", pkg, version},
			Description: fmt.Sprintf("Set global %s version to %s", pkg, version),
		}
		commands = append(commands, globalCmd)
	}

	return commands
}
