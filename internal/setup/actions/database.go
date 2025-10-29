package actions

import (
	"fmt"
	"os/exec"
	"strings"
)

type EnsureDatabaseAction struct {
	*BaseAction
	engine        string
	version       string
	serviceName   string
	ensureRunning bool
	createDB      []string
}

func NewEnsureDatabaseAction(
	engine, version, serviceName string,
	ensureRunning bool,
	createDB []string,
) *EnsureDatabaseAction {
	return &EnsureDatabaseAction{
		BaseAction:    NewBaseAction(ActionTypeEnsureDatabase),
		engine:        engine,
		version:       version,
		serviceName:   serviceName,
		ensureRunning: ensureRunning,
		createDB:      createDB,
	}
}

func (a *EnsureDatabaseAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.engine) == "" {
		return nil, NewActionError(ActionTypeEnsureDatabase, "database engine cannot be empty", nil)
	}

	commands, err := a.getInstallCommands(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeEnsureDatabase, "failed to generate install commands", err)
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

func (a *EnsureDatabaseAction) Validate() error {
	if strings.TrimSpace(a.engine) == "" {
		return NewActionError(ActionTypeEnsureDatabase, "database engine cannot be empty", nil)
	}

	// Validate database engine
	validEngines := []string{DBPostgres, DBMySQL, DBSQLite}
	if !contains(validEngines, a.engine) {
		return NewActionError(
			ActionTypeEnsureDatabase,
			fmt.Sprintf("invalid database engine '%s', must be one of: %s", a.engine, strings.Join(validEngines, ", ")),
			nil,
		)
	}

	// Set default service name if not provided
	if a.serviceName == "" {
		a.serviceName = a.getDefaultServiceName()
	}

	return nil
}

type DatabaseCommand struct {
	Command     string
	Args        []string
	Description string
}

// getInstallCommands generates database installation and setup commands
func (a *EnsureDatabaseAction) getInstallCommands(ctx *ActionContext) ([]DatabaseCommand, error) {
	os := strings.ToLower(ctx.OS)
	var commands []DatabaseCommand

	// Add installation commands based on OS
	switch os {
	case OSMacOS, OSDarwin:
		commands = append(commands, a.getBrewInstallCommands()...)
	case OSLinux:
		commands = append(commands, a.getAptInstallCommands()...)
	case OSWindows:
		commands = append(commands, a.getChocoInstallCommands()...)
	default:
		return nil, fmt.Errorf("unsupported OS for database installation: %s", os)
	}

	// Add service management commands if ensureRunning is true
	if a.ensureRunning {
		commands = append(commands, a.getServiceCommands(os)...)
	}

	// Add database creation commands
	if len(a.createDB) > 0 {
		commands = append(commands, a.getCreateDatabaseCommands()...)
	}

	return commands, nil
}

// getBrewInstallCommands generates Homebrew installation commands
func (a *EnsureDatabaseAction) getBrewInstallCommands() []DatabaseCommand {
	formula := a.getBrewFormula()

	commands := []DatabaseCommand{
		{
			Command:     "which",
			Args:        []string{a.getExecutableName()},
			Description: fmt.Sprintf("Check if %s is installed", a.engine),
		},
	}

	brewArgs := []string{"install", formula}
	if a.version != "" {
		brewArgs = append(brewArgs, a.version)
	}

	commands = append(commands, DatabaseCommand{
		Command:     "brew",
		Args:        brewArgs,
		Description: fmt.Sprintf("Install %s via Homebrew", a.engine),
	})

	return commands
}

// getAptInstallCommands generates APT installation commands
func (a *EnsureDatabaseAction) getAptInstallCommands() []DatabaseCommand {
	packages := a.getAptPackages()

	commands := []DatabaseCommand{
		{
			Command:     "which",
			Args:        []string{a.getExecutableName()},
			Description: fmt.Sprintf("Check if %s is installed", a.engine),
		},
		{
			Command:     "sudo",
			Args:        []string{"apt", "update"},
			Description: "Update package lists",
		},
	}

	aptArgs := []string{"apt", "install", "-y"}
	aptArgs = append(aptArgs, packages...)

	commands = append(commands, DatabaseCommand{
		Command:     "sudo",
		Args:        aptArgs,
		Description: fmt.Sprintf("Install %s via apt", a.engine),
	})

	return commands
}

// getChocoInstallCommands generates Chocolatey installation commands
func (a *EnsureDatabaseAction) getChocoInstallCommands() []DatabaseCommand {
	packages := a.getChocoPackages()

	commands := []DatabaseCommand{
		{
			Command:     "where",
			Args:        []string{a.getExecutableName()},
			Description: fmt.Sprintf("Check if %s is installed", a.engine),
		},
	}

	chocoArgs := []string{"install", "-y"}
	chocoArgs = append(chocoArgs, packages...)

	commands = append(commands, DatabaseCommand{
		Command:     "choco",
		Args:        chocoArgs,
		Description: fmt.Sprintf("Install %s via Chocolatey", a.engine),
	})

	return commands
}

// getServiceCommands generates service management commands
func (a *EnsureDatabaseAction) getServiceCommands(os string) []DatabaseCommand {
	var commands []DatabaseCommand

	switch os {
	case OSMacOS, OSDarwin:
		// macOS uses brew services
		commands = append(commands, DatabaseCommand{
			Command:     "brew",
			Args:        []string{"services", "start", a.serviceName},
			Description: fmt.Sprintf("Start %s service", a.engine),
		})
	case OSLinux:
		// Linux uses systemctl
		commands = append(commands, DatabaseCommand{
			Command:     "sudo",
			Args:        []string{"systemctl", "start", a.serviceName},
			Description: fmt.Sprintf("Start %s service", a.engine),
		})
		commands = append(commands, DatabaseCommand{
			Command:     "sudo",
			Args:        []string{"systemctl", "enable", a.serviceName},
			Description: fmt.Sprintf("Enable %s service to start on boot", a.engine),
		})
	case OSWindows:
		// Windows uses net start
		commands = append(commands, DatabaseCommand{
			Command:     "net",
			Args:        []string{"start", a.serviceName},
			Description: fmt.Sprintf("Start %s service", a.engine),
		})
	}

	return commands
}

// getCreateDatabaseCommands generates database creation commands
func (a *EnsureDatabaseAction) getCreateDatabaseCommands() []DatabaseCommand {
	var commands []DatabaseCommand

	for _, dbName := range a.createDB {
		switch a.engine {
		case DBPostgres:
			commands = append(commands, DatabaseCommand{
				Command:     "createdb",
				Args:        []string{dbName},
				Description: fmt.Sprintf("Create PostgreSQL database '%s'", dbName),
			})
		case DBMySQL:
			commands = append(commands, DatabaseCommand{
				Command:     "mysql",
				Args:        []string{"-e", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName)},
				Description: fmt.Sprintf("Create MySQL database '%s'", dbName),
			})
		case DBSQLite:
			commands = append(commands, DatabaseCommand{
				Command:     "sqlite3",
				Args:        []string{dbName + ".db", ".databases"},
				Description: fmt.Sprintf("Create SQLite database '%s.db'", dbName),
			})
		}
	}

	return commands
}

// getDefaultServiceName returns the default service name for the database engine
func (a *EnsureDatabaseAction) getDefaultServiceName() string {
	switch a.engine {
	case DBPostgres:
		return "postgresql"
	case DBMySQL:
		return DBMySQL
	case DBSQLite:
		return "" // SQLite doesn't have a service
	default:
		return a.engine
	}
}

// getExecutableName returns the executable name for the database engine
func (a *EnsureDatabaseAction) getExecutableName() string {
	switch a.engine {
	case DBPostgres:
		return "psql"
	case DBMySQL:
		return "mysql"
	case DBSQLite:
		return "sqlite3"
	default:
		return a.engine
	}
}

// getBrewFormula returns the Homebrew formula name for the database engine
func (a *EnsureDatabaseAction) getBrewFormula() string {
	switch a.engine {
	case DBPostgres:
		return "postgresql"
	case DBMySQL:
		return "mysql"
	case DBSQLite:
		return "sqlite"
	default:
		return a.engine
	}
}

// getAptPackages returns the APT package names for the database engine
func (a *EnsureDatabaseAction) getAptPackages() []string {
	switch a.engine {
	case DBPostgres:
		return []string{"postgresql", "postgresql-contrib"}
	case DBMySQL:
		return []string{"mysql-server", "mysql-client"}
	case DBSQLite:
		return []string{"sqlite3"}
	default:
		return []string{a.engine}
	}
}

// getChocoPackages returns the Chocolatey package names for the database engine
func (a *EnsureDatabaseAction) getChocoPackages() []string {
	switch a.engine {
	case DBPostgres:
		return []string{"postgresql"}
	case DBMySQL:
		return []string{"mysql"}
	case DBSQLite:
		return []string{"sqlite"}
	default:
		return []string{a.engine}
	}
}

// IsInstalled checks if the database is already installed
func (a *EnsureDatabaseAction) IsInstalled(ctx *ActionContext) (bool, error) {
	_, err := exec.LookPath(a.getExecutableName())
	return err == nil, nil
}

// IsRunning checks if the database service is running
func (a *EnsureDatabaseAction) IsRunning(ctx *ActionContext) (bool, error) {
	if a.engine == "sqlite" {
		return true, nil // SQLite doesn't have a service
	}

	os := strings.ToLower(ctx.OS)
	switch os {
	case OSMacOS, OSDarwin:
		// Check brew services
		cmd := exec.Command("brew", "services", "list", "--json")
		output, err := cmd.Output()
		if err != nil {
			return false, err
		}
		return strings.Contains(string(output), fmt.Sprintf(`"name":"%s","status":"started"`, a.serviceName)), nil
	case OSLinux:
		// Check systemctl
		cmd := exec.Command("systemctl", "is-active", a.serviceName) // #nosec G204
		output, err := cmd.Output()
		if err != nil {
			return false, err
		}
		return strings.TrimSpace(string(output)) == "active", nil
	case OSWindows:
		// Check net start
		cmd := exec.Command("net", "start")
		output, err := cmd.Output()
		if err != nil {
			return false, err
		}
		return strings.Contains(string(output), a.serviceName), nil
	default:
		return false, fmt.Errorf("unsupported OS for service check: %s", os)
	}
}
