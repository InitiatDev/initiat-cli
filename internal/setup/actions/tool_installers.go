package actions

import (
	"fmt"
	"os"
	"path/filepath"
)

// ============================================================================
// TOOL INSTALLER IMPLEMENTATIONS
// ============================================================================
// Each installer handles a specific package manager or installation method
// ============================================================================

// Homebrew installer
func (a *EnsureToolAction) getBrewCommands() ([]ToolCommand, error) {
	if a.installConfig.Brew == nil {
		return nil, fmt.Errorf("brew configuration not provided")
	}

	formula := a.installConfig.Brew.Formula
	if formula == "" {
		formula = a.toolName
	}

	commands := []ToolCommand{
		{
			Command:     "which",
			Args:        []string{a.toolName},
			Description: fmt.Sprintf("Check if %s is installed", a.toolName),
		},
	}

	brewArgs := []string{"install", formula}
	if a.version != "" {
		brewArgs = append(brewArgs, a.version)
	}

	commands = append(commands, ToolCommand{
		Command:     "brew",
		Args:        brewArgs,
		Description: fmt.Sprintf("Install %s via Homebrew", a.toolName),
	})

	return commands, nil
}

// APT installer
func (a *EnsureToolAction) getAptCommands() ([]ToolCommand, error) {
	if a.installConfig.Apt == nil {
		return nil, fmt.Errorf("apt configuration not provided")
	}

	packages := a.installConfig.Apt.Packages
	if len(packages) == 0 {
		packages = []string{a.toolName}
	}

	commands := []ToolCommand{
		{
			Command:     "which",
			Args:        []string{a.toolName},
			Description: fmt.Sprintf("Check if %s is installed", a.toolName),
		},
	}

	if a.installConfig.Apt.Update {
		commands = append(commands, ToolCommand{
			Command:     "sudo",
			Args:        []string{"apt", "update"},
			Description: "Update package lists",
		})
	}

	aptArgs := []string{"apt", "install", "-y"}
	aptArgs = append(aptArgs, packages...)

	commands = append(commands, ToolCommand{
		Command:     "sudo",
		Args:        aptArgs,
		Description: fmt.Sprintf("Install %s via apt", a.toolName),
	})

	return commands, nil
}

// Chocolatey installer
func (a *EnsureToolAction) getChocoCommands() ([]ToolCommand, error) {
	if a.installConfig.Choco == nil {
		return nil, fmt.Errorf("choco configuration not provided")
	}

	packages := a.installConfig.Choco.Packages
	if len(packages) == 0 {
		packages = []string{a.toolName}
	}

	commands := []ToolCommand{
		{
			Command:     "where",
			Args:        []string{a.toolName},
			Description: fmt.Sprintf("Check if %s is installed", a.toolName),
		},
	}

	chocoArgs := []string{"install", "-y"}
	chocoArgs = append(chocoArgs, packages...)

	commands = append(commands, ToolCommand{
		Command:     "choco",
		Args:        chocoArgs,
		Description: fmt.Sprintf("Install %s via Chocolatey", a.toolName),
	})

	return commands, nil
}

// Fallback installer (direct download)
func (a *EnsureToolAction) getFallbackCommands(ctx *ActionContext) ([]ToolCommand, error) {
	if a.installConfig.FallbackURL == "" {
		return nil, fmt.Errorf("fallback URL not provided")
	}

	// Determine the appropriate binary name and install location
	binaryName := a.toolName
	if ctx.OS == OSWindows {
		binaryName += ".exe"
	}

	installDir := "/usr/local/bin"
	if ctx.OS == OSWindows {
		installDir = filepath.Join(os.Getenv("PROGRAMFILES"), "bin")
	}

	commands := []ToolCommand{
		{
			Command:     "which",
			Args:        []string{a.toolName},
			Description: fmt.Sprintf("Check if %s is installed", a.toolName),
		},
	}

	// Create install directory
	commands = append(commands, ToolCommand{
		Command:     "mkdir",
		Args:        []string{"-p", installDir},
		Description: "Create install directory",
	})

	// Download the binary
	downloadCmd := fmt.Sprintf("curl -L -o %s %s",
		filepath.Join(installDir, binaryName),
		a.installConfig.FallbackURL)

	if ctx.OS == OSWindows {
		downloadCmd = fmt.Sprintf("powershell -Command \"Invoke-WebRequest -Uri %s -OutFile %s\"",
			a.installConfig.FallbackURL,
			filepath.Join(installDir, binaryName))
	}

	commands = append(commands, ToolCommand{
		Command:     "bash",
		Args:        []string{"-c", downloadCmd},
		Description: fmt.Sprintf("Download %s", a.toolName),
	})

	// Make executable (Unix systems)
	if ctx.OS != OSWindows {
		commands = append(commands, ToolCommand{
			Command:     "chmod",
			Args:        []string{"+x", filepath.Join(installDir, binaryName)},
			Description: "Make binary executable",
		})
	}

	// Verify checksum if provided
	if a.installConfig.Checksum != "" {
		checksumCmd := fmt.Sprintf("echo '%s %s' | sha256sum -c",
			a.installConfig.Checksum,
			filepath.Join(installDir, binaryName))

		if ctx.OS == OSWindows {
			checksumCmd = fmt.Sprintf("powershell -Command \"(Get-FileHash %s -Algorithm SHA256).Hash -eq '%s'\"",
				filepath.Join(installDir, binaryName),
				a.installConfig.Checksum)
		}

		commands = append(commands, ToolCommand{
			Command:     "bash",
			Args:        []string{"-c", checksumCmd},
			Description: "Verify checksum",
		})
	}

	return commands, nil
}
