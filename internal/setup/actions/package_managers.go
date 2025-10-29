package actions

import (
	"fmt"
	"os/exec"
)

// Homebrew (macOS)
func (a *EnsurePackageManagerAction) getBrewCommands(os, _ string) ([]PackageManagerCommand, error) {
	if os != OSMacOS && os != OSDarwin {
		return nil, fmt.Errorf("homebrew is only supported on macOS")
	}

	installScript := "/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""

	return []PackageManagerCommand{
		{
			Command:     "which",
			Args:        []string{"brew"},
			Description: "Check if Homebrew is installed",
		},
		{
			Command:     "bash",
			Args:        []string{"-c", installScript},
			Description: "Install Homebrew",
		},
	}, nil
}

func (a *EnsurePackageManagerAction) isBrewInstalled() (bool, error) {
	_, err := exec.LookPath("brew")
	return err == nil, nil
}

// APT (Debian/Ubuntu)
func (a *EnsurePackageManagerAction) getAptCommands(os string) ([]PackageManagerCommand, error) {
	if os != OSLinux {
		return nil, fmt.Errorf("apt is only supported on Linux")
	}

	return []PackageManagerCommand{
		{
			Command:     "which",
			Args:        []string{"apt"},
			Description: "Check if apt is available",
		},
		{
			Command:     "sudo",
			Args:        []string{"apt", "update"},
			Description: "Update package lists",
		},
		{
			Command: "sudo",
			Args: []string{
				"apt", "install", "-y",
				"apt-transport-https", "ca-certificates", "curl", "gnupg", "lsb-release",
			},
			Description: "Install essential apt packages",
		},
	}, nil
}

func (a *EnsurePackageManagerAction) isAptInstalled() (bool, error) {
	_, err := exec.LookPath("apt")
	return err == nil, nil
}

// YUM (RHEL/CentOS)
func (a *EnsurePackageManagerAction) getYumCommands(os string) ([]PackageManagerCommand, error) {
	if os != OSLinux {
		return nil, fmt.Errorf("yum is only supported on Linux")
	}

	return []PackageManagerCommand{
		{
			Command:     "which",
			Args:        []string{"yum"},
			Description: "Check if yum is available",
		},
		{
			Command:     "sudo",
			Args:        []string{"yum", "update", "-y"},
			Description: "Update yum packages",
		},
	}, nil
}

func (a *EnsurePackageManagerAction) isYumInstalled() (bool, error) {
	_, err := exec.LookPath("yum")
	return err == nil, nil
}

// DNF (Fedora)
func (a *EnsurePackageManagerAction) getDnfCommands(os string) ([]PackageManagerCommand, error) {
	if os != OSLinux {
		return nil, fmt.Errorf("dnf is only supported on Linux")
	}

	return []PackageManagerCommand{
		{
			Command:     "which",
			Args:        []string{"dnf"},
			Description: "Check if dnf is available",
		},
		{
			Command:     "sudo",
			Args:        []string{"dnf", "update", "-y"},
			Description: "Update dnf packages",
		},
	}, nil
}

func (a *EnsurePackageManagerAction) isDnfInstalled() (bool, error) {
	_, err := exec.LookPath("dnf")
	return err == nil, nil
}

// Pacman (Arch Linux)
func (a *EnsurePackageManagerAction) getPacmanCommands(os string) ([]PackageManagerCommand, error) {
	if os != "linux" {
		return nil, fmt.Errorf("pacman is only supported on Linux")
	}

	return []PackageManagerCommand{
		{
			Command:     "which",
			Args:        []string{"pacman"},
			Description: "Check if pacman is available",
		},
		{
			Command:     "sudo",
			Args:        []string{"pacman", "-Sy"},
			Description: "Update package database",
		},
	}, nil
}

func (a *EnsurePackageManagerAction) isPacmanInstalled() (bool, error) {
	_, err := exec.LookPath("pacman")
	return err == nil, nil
}

// Chocolatey (Windows)
func (a *EnsurePackageManagerAction) getChocoCommands(os string) ([]PackageManagerCommand, error) {
	if os != OSWindows {
		return nil, fmt.Errorf("chocolatey is only supported on Windows")
	}

	installScript := "Set-ExecutionPolicy Bypass -Scope Process -Force; " +
		"[System.Net.ServicePointManager]::SecurityProtocol = " +
		"[System.Net.ServicePointManager]::SecurityProtocol -bor 3072; " +
		"iex ((New-Object System.Net.WebClient).DownloadString(" +
		"'https://community.chocolatey.org/install.ps1'))"

	return []PackageManagerCommand{
		{
			Command:     "where",
			Args:        []string{"choco"},
			Description: "Check if Chocolatey is installed",
		},
		{
			Command:     "powershell",
			Args:        []string{"-Command", installScript},
			Description: "Install Chocolatey",
		},
	}, nil
}

func (a *EnsurePackageManagerAction) isChocoInstalled() (bool, error) {
	_, err := exec.LookPath("choco")
	return err == nil, nil
}

// Scoop (Windows)
func (a *EnsurePackageManagerAction) getScoopCommands(os string) ([]PackageManagerCommand, error) {
	if os != "windows" {
		return nil, fmt.Errorf("scoop is only supported on Windows")
	}

	installScript := "Set-ExecutionPolicy RemoteSigned -Scope CurrentUser; irm get.scoop.sh | iex"

	return []PackageManagerCommand{
		{
			Command:     "where",
			Args:        []string{"scoop"},
			Description: "Check if Scoop is installed",
		},
		{
			Command:     "powershell",
			Args:        []string{"-Command", installScript},
			Description: "Install Scoop",
		},
	}, nil
}

func (a *EnsurePackageManagerAction) isScoopInstalled() (bool, error) {
	_, err := exec.LookPath("scoop")
	return err == nil, nil
}

// Winget (Windows 10+)
func (a *EnsurePackageManagerAction) getWingetCommands(os string) ([]PackageManagerCommand, error) {
	if os != OSWindows {
		return nil, fmt.Errorf("winget is only supported on Windows")
	}

	return []PackageManagerCommand{
		{
			Command:     "where",
			Args:        []string{"winget"},
			Description: "Check if winget is available",
		},
	}, nil
}

func (a *EnsurePackageManagerAction) isWingetInstalled() (bool, error) {
	_, err := exec.LookPath("winget")
	return err == nil, nil
}
