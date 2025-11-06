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

func (p *ChocoPackageManager) InstallSelfCommand() types.Command {
	installScript := "Set-ExecutionPolicy Bypass -Scope Process -Force; " +
		"[System.Net.ServicePointManager]::SecurityProtocol = " +
		"[System.Net.ServicePointManager]::SecurityProtocol -bor 3072; " +
		"iex ((New-Object System.Net.WebClient).DownloadString(" +
		"'https://community.chocolatey.org/install.ps1'))"
	return types.Command{
		Command: "powershell",
		Args: []string{
			"-NoProfile",
			"-ExecutionPolicy",
			"Bypass",
			"-Command",
			installScript,
		},
		Description: "Install Chocolatey",
	}
}

func (p *ChocoPackageManager) ExtractToolInstallCommands(config interface{}) ([]types.Command, bool) {
	cfg, ok := config.(*types.ToolInstallConfig)
	if !ok || cfg.Choco == nil || len(cfg.Choco.Packages) == 0 {
		return nil, false
	}

	var commands []types.Command
	for _, pkg := range cfg.Choco.Packages {
		commands = append(commands, p.InstallCommand(pkg, ""))
	}

	return commands, true
}
