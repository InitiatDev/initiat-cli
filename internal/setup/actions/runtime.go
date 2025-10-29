package actions

import (
	"fmt"
	"os/exec"
	"strings"
)

type EnsureRuntimeAction struct {
	*BaseAction
	runtimeName        string
	version            string
	manager            *RuntimeManager
	fallbackInstallers []RuntimeFallbackInstaller
}

type RuntimeManager struct {
	Asdf bool `yaml:"asdf,omitempty" json:"asdf,omitempty"`
}

type RuntimeFallbackInstaller struct {
	Brew  *BrewInstall  `yaml:"brew,omitempty" json:"brew,omitempty"`
	Apt   *AptInstall   `yaml:"apt,omitempty" json:"apt,omitempty"`
	Choco *ChocoInstall `yaml:"choco,omitempty" json:"choco,omitempty"`
}

func NewEnsureRuntimeAction(
	runtimeName, version string,
	manager *RuntimeManager,
	fallbackInstallers []RuntimeFallbackInstaller,
) *EnsureRuntimeAction {
	return &EnsureRuntimeAction{
		BaseAction:         NewBaseAction(ActionTypeEnsureRuntime),
		runtimeName:        runtimeName,
		version:            version,
		manager:            manager,
		fallbackInstallers: fallbackInstallers,
	}
}

func (a *EnsureRuntimeAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.runtimeName) == "" {
		return nil, NewActionError(ActionTypeEnsureRuntime, "runtime name cannot be empty", nil)
	}

	commands, err := a.getInstallCommands(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsureRuntime, "failed to generate install commands", err)
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

func (a *EnsureRuntimeAction) Validate() error {
	if strings.TrimSpace(a.runtimeName) == "" {
		return NewActionError(ActionTypeEnsureRuntime, "runtime name cannot be empty", nil)
	}

	// Validate runtime name
	validRuntimes := []string{
		RuntimeNode, RuntimePython, RuntimeGo, RuntimeElixir,
		RuntimeErlang, RuntimeJava, RuntimeRust, RuntimeDotnet,
	}
	if !contains(validRuntimes, a.runtimeName) {
		return NewActionError(
			ActionTypeEnsureRuntime,
			fmt.Sprintf("invalid runtime name '%s', must be one of: %s", a.runtimeName, strings.Join(validRuntimes, ", ")),
			nil,
		)
	}

	// Validate that at least one install method is available
	hasInstallMethod := a.manager != nil && a.manager.Asdf
	for _, installer := range a.fallbackInstallers {
		if installer.Brew != nil || installer.Apt != nil || installer.Choco != nil {
			hasInstallMethod = true
			break
		}
	}

	if !hasInstallMethod {
		return NewActionError(ActionTypeEnsureRuntime, "at least one install method must be specified", nil)
	}

	return nil
}

type RuntimeCommand struct {
	Command     string
	Args        []string
	Description string
}

// getInstallCommands determines the best install method for the current platform
func (a *EnsureRuntimeAction) getInstallCommands(ctx *ActionContext) ([]RuntimeCommand, error) {
	os := strings.ToLower(ctx.OS)

	// Try asdf first if configured
	if a.manager != nil && a.manager.Asdf {
		if a.isAsdfAvailable() {
			return a.getAsdfCommands(ctx)
		}
	}

	// Try platform-specific package managers
	for _, installer := range a.fallbackInstallers {
		switch os {
		case OSMacOS, OSDarwin:
			if installer.Brew != nil {
				return a.getBrewCommands(installer.Brew, ctx)
			}
		case OSLinux:
			if installer.Apt != nil {
				return a.getAptCommands(installer.Apt, ctx)
			}
		case OSWindows:
			if installer.Choco != nil {
				return a.getChocoCommands(installer.Choco, ctx)
			}
		}
	}

	return nil, fmt.Errorf("no suitable install method found for %s on %s", a.runtimeName, os)
}

// isAsdfAvailable checks if asdf is installed
func (a *EnsureRuntimeAction) isAsdfAvailable() bool {
	_, err := exec.LookPath("asdf")
	return err == nil
}

// getAsdfCommands generates asdf-based install commands
func (a *EnsureRuntimeAction) getAsdfCommands(_ *ActionContext) ([]RuntimeCommand, error) {
	commands := []RuntimeCommand{
		{
			Command:     "asdf",
			Args:        []string{"list", a.runtimeName},
			Description: fmt.Sprintf("Check if %s is installed via asdf", a.runtimeName),
		},
	}

	// Add asdf plugin if not already added
	commands = append(commands, RuntimeCommand{
		Command:     "asdf",
		Args:        []string{"plugin", "add", a.runtimeName},
		Description: fmt.Sprintf("Add %s plugin to asdf", a.runtimeName),
	})

	// Install specific version if specified
	installArgs := []string{"install", a.runtimeName}
	if a.version != "" {
		installArgs = append(installArgs, a.version)
	} else {
		installArgs = append(installArgs, "latest")
	}

	commands = append(commands, RuntimeCommand{
		Command:     "asdf",
		Args:        installArgs,
		Description: fmt.Sprintf("Install %s via asdf", a.runtimeName),
	})

	// Set global version if specified
	if a.version != "" {
		commands = append(commands, RuntimeCommand{
			Command:     "asdf",
			Args:        []string{"global", a.runtimeName, a.version},
			Description: fmt.Sprintf("Set global %s version to %s", a.runtimeName, a.version),
		})
	}

	return commands, nil
}

// getBrewCommands generates Homebrew-based install commands
func (a *EnsureRuntimeAction) getBrewCommands(brew *BrewInstall, _ *ActionContext) ([]RuntimeCommand, error) {
	formula := brew.Formula
	if formula == "" {
		formula = a.getBrewFormula()
	}

	commands := []RuntimeCommand{
		{
			Command:     "which",
			Args:        []string{a.runtimeName},
			Description: fmt.Sprintf("Check if %s is installed", a.runtimeName),
		},
	}

	brewArgs := []string{"install", formula}
	if a.version != "" {
		brewArgs = append(brewArgs, a.version)
	}

	commands = append(commands, RuntimeCommand{
		Command:     "brew",
		Args:        brewArgs,
		Description: fmt.Sprintf("Install %s via Homebrew", a.runtimeName),
	})

	return commands, nil
}

// getAptCommands generates APT-based install commands
func (a *EnsureRuntimeAction) getAptCommands(apt *AptInstall, _ *ActionContext) ([]RuntimeCommand, error) {
	packages := apt.Packages
	if len(packages) == 0 {
		packages = []string{a.getAptPackage()}
	}

	commands := []RuntimeCommand{
		{
			Command:     "which",
			Args:        []string{a.runtimeName},
			Description: fmt.Sprintf("Check if %s is installed", a.runtimeName),
		},
	}

	if apt.Update {
		commands = append(commands, RuntimeCommand{
			Command:     "sudo",
			Args:        []string{"apt", "update"},
			Description: "Update package lists",
		})
	}

	aptArgs := []string{"apt", "install", "-y"}
	aptArgs = append(aptArgs, packages...)

	commands = append(commands, RuntimeCommand{
		Command:     "sudo",
		Args:        aptArgs,
		Description: fmt.Sprintf("Install %s via apt", a.runtimeName),
	})

	return commands, nil
}

// getChocoCommands generates Chocolatey-based install commands
func (a *EnsureRuntimeAction) getChocoCommands(choco *ChocoInstall, _ *ActionContext) ([]RuntimeCommand, error) {
	packages := choco.Packages
	if len(packages) == 0 {
		packages = []string{a.getChocoPackage()}
	}

	commands := []RuntimeCommand{
		{
			Command:     "where",
			Args:        []string{a.runtimeName},
			Description: fmt.Sprintf("Check if %s is installed", a.runtimeName),
		},
	}

	chocoArgs := []string{"install", "-y"}
	chocoArgs = append(chocoArgs, packages...)

	commands = append(commands, RuntimeCommand{
		Command:     "choco",
		Args:        chocoArgs,
		Description: fmt.Sprintf("Install %s via Chocolatey", a.runtimeName),
	})

	return commands, nil
}

// getBrewFormula returns the Homebrew formula name for the runtime
func (a *EnsureRuntimeAction) getBrewFormula() string {
	switch a.runtimeName {
	case RuntimeNode:
		return "node"
	case RuntimePython:
		return "python"
	case RuntimeGo:
		return "go"
	case RuntimeElixir:
		return RuntimeElixir
	case RuntimeErlang:
		return RuntimeErlang
	case RuntimeJava:
		return "openjdk"
	case RuntimeRust:
		return "rust"
	case RuntimeDotnet:
		return "dotnet"
	default:
		return a.runtimeName
	}
}

// getAptPackage returns the APT package name for the runtime
func (a *EnsureRuntimeAction) getAptPackage() string {
	switch a.runtimeName {
	case RuntimeNode:
		return "nodejs"
	case RuntimePython:
		return "python3"
	case RuntimeGo:
		return "golang-go"
	case RuntimeElixir:
		return "elixir"
	case RuntimeErlang:
		return "erlang"
	case RuntimeJava:
		return "openjdk-11-jdk"
	case RuntimeRust:
		return "rustc"
	case RuntimeDotnet:
		return "dotnet-sdk-6.0"
	default:
		return a.runtimeName
	}
}

// getChocoPackage returns the Chocolatey package name for the runtime
func (a *EnsureRuntimeAction) getChocoPackage() string {
	switch a.runtimeName {
	case RuntimeNode:
		return "nodejs"
	case RuntimePython:
		return "python"
	case RuntimeGo:
		return "golang"
	case RuntimeElixir:
		return "elixir"
	case RuntimeErlang:
		return "erlang"
	case RuntimeJava:
		return "openjdk"
	case RuntimeRust:
		return "rust"
	case RuntimeDotnet:
		return "dotnet"
	default:
		return a.runtimeName
	}
}

// IsInstalled checks if the runtime is already installed
func (a *EnsureRuntimeAction) IsInstalled(ctx *ActionContext) (bool, error) {
	_, err := exec.LookPath(a.runtimeName)
	return err == nil, nil
}

// GetVersion checks the installed version of the runtime
func (a *EnsureRuntimeAction) GetVersion(ctx *ActionContext) (string, error) {
	cmd := exec.Command(a.runtimeName, "--version") // #nosec G204
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
