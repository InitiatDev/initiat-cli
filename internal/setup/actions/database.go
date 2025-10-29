package actions

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/registry"
)

type EnsureDatabaseAction struct {
	*BaseAction
	engine        string
	version       string
	serviceName   string
	ensureRunning bool
	createDB      []string
	pkgRegistry   *registry.PackageManagerRegistry
	svcRegistry   *registry.ServiceManagerRegistry
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
		pkgRegistry:   registry.NewPackageManagerRegistry(),
		svcRegistry:   registry.NewServiceManagerRegistry(),
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
	var commands []DatabaseCommand

	pkgManager := a.pkgRegistry.FindManager(ctx.OS)
	if pkgManager == nil {
		return nil, fmt.Errorf("no suitable package manager found for %s on %s", a.engine, ctx.OS)
	}

	installCmd := pkgManager.InstallCommand(a.engine, a.version)
	commands = append(commands, DatabaseCommand{
		Command:     installCmd.Command,
		Args:        installCmd.Args,
		Description: installCmd.Description,
	})

	if a.ensureRunning {
		svcManager := a.svcRegistry.FindManager(ctx.OS)
		if svcManager != nil {
			startCmd := svcManager.StartServiceCommand(a.serviceName)
			commands = append(commands, DatabaseCommand{
				Command:     startCmd.Command,
				Args:        startCmd.Args,
				Description: startCmd.Description,
			})
		}
	}

	if len(a.createDB) > 0 {
		commands = append(commands, a.getCreateDatabaseCommands()...)
	}

	return commands, nil
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

// IsInstalled checks if the database is already installed
func (a *EnsureDatabaseAction) IsInstalled(ctx *ActionContext) (bool, error) {
	_, err := exec.LookPath(a.getExecutableName())
	return err == nil, nil
}

// IsRunning checks if the database service is running
func (a *EnsureDatabaseAction) IsRunning(ctx *ActionContext) (bool, error) {
	if a.engine == "sqlite" {
		return true, nil
	}

	svcManager := a.svcRegistry.FindManager(ctx.OS)
	if svcManager == nil {
		return false, fmt.Errorf("no suitable service manager found for %s on %s", a.serviceName, ctx.OS)
	}

	return true, nil
}
