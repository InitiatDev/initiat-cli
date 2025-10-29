package actions

import (
	"fmt"
	"os/exec"
	"strings"
)

type EnsureToolAction struct {
	*BaseAction
	toolName      string
	version       string
	installConfig *ToolInstallConfig
}

type ToolInstallConfig struct {
	Brew        *BrewInstall  `yaml:"brew,omitempty" json:"brew,omitempty"`
	Apt         *AptInstall   `yaml:"apt,omitempty" json:"apt,omitempty"`
	Choco       *ChocoInstall `yaml:"choco,omitempty" json:"choco,omitempty"`
	FallbackURL string        `yaml:"fallback_url,omitempty" json:"fallback_url,omitempty"`
	Checksum    string        `yaml:"checksum,omitempty" json:"checksum,omitempty"`
}

type BrewInstall struct {
	Formula string `yaml:"formula" json:"formula"`
}

type AptInstall struct {
	Packages []string `yaml:"packages" json:"packages"`
	Update   bool     `yaml:"update,omitempty" json:"update,omitempty"`
}

type ChocoInstall struct {
	Packages []string `yaml:"packages" json:"packages"`
}

func NewEnsureToolAction(toolName, version string, installConfig *ToolInstallConfig) *EnsureToolAction {
	return &EnsureToolAction{
		BaseAction:    NewBaseAction(ActionTypeEnsureTool),
		toolName:      toolName,
		version:       version,
		installConfig: installConfig,
	}
}

func (a *EnsureToolAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.toolName) == "" {
		return nil, NewActionError(ActionTypeEnsureTool, "tool name cannot be empty", nil)
	}

	commands, err := a.getInstallCommands(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsureTool, "failed to generate install commands", err)
	}

	var result []Command
	for _, cmd := range commands {
		result = append(result, Command{
			Command:     cmd.Command,
			Args:        cmd.Args,
			Env:         ctx.Env,
			WorkingDir:  ctx.WorkingDir,
			Timeout:     ctx.Timeout,
			Description: cmd.Description,
		})
	}

	return result, nil
}

func (a *EnsureToolAction) Validate() error {
	if strings.TrimSpace(a.toolName) == "" {
		return NewActionError(ActionTypeEnsureTool, "tool name cannot be empty", nil)
	}

	if a.installConfig == nil {
		return NewActionError(ActionTypeEnsureTool, "install configuration is required", nil)
	}

	// Validate that at least one install method is specified
	hasInstallMethod := a.installConfig.Brew != nil ||
		a.installConfig.Apt != nil ||
		a.installConfig.Choco != nil ||
		a.installConfig.FallbackURL != ""

	if !hasInstallMethod {
		return NewActionError(ActionTypeEnsureTool, "at least one install method must be specified", nil)
	}

	return nil
}

type ToolCommand struct {
	Command     string
	Args        []string
	Description string
}

// getInstallCommands determines the best install method for the current platform
func (a *EnsureToolAction) getInstallCommands(ctx *ActionContext) ([]ToolCommand, error) {
	os := strings.ToLower(ctx.OS)

	// Try platform-specific package managers first
	switch os {
	case OSMacOS, OSDarwin:
		if a.installConfig.Brew != nil {
			return a.getBrewCommands()
		}
	case OSLinux:
		if a.installConfig.Apt != nil {
			return a.getAptCommands()
		}
	case OSWindows:
		if a.installConfig.Choco != nil {
			return a.getChocoCommands()
		}
	}

	// Fall back to direct download if available
	if a.installConfig.FallbackURL != "" {
		return a.getFallbackCommands(ctx)
	}

	return nil, fmt.Errorf("no suitable install method found for %s on %s", a.toolName, os)
}

// IsInstalled checks if the tool is already installed
func (a *EnsureToolAction) IsInstalled(ctx *ActionContext) (bool, error) {
	_, err := exec.LookPath(a.toolName)
	return err == nil, nil
}
