package actions

import (
	"strings"
	"testing"
	"time"
)

func TestEnsureDatabaseAction_Validate(t *testing.T) {
	tests := []struct {
		name          string
		engine        string
		version       string
		serviceName   string
		ensureRunning bool
		createDB      []string
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "valid postgres",
			engine:        "postgres",
			version:       "15",
			serviceName:   "postgresql",
			ensureRunning: true,
			createDB:      []string{"myapp", "test"},
			expectError:   false,
		},
		{
			name:          "valid mysql",
			engine:        "mysql",
			version:       "8.0",
			ensureRunning: true,
			expectError:   false,
		},
		{
			name:        "valid sqlite",
			engine:      "sqlite",
			createDB:    []string{"myapp"},
			expectError: false,
		},
		{
			name:        "empty engine",
			engine:      "",
			expectError: true,
			errorMsg:    "database engine cannot be empty",
		},
		{
			name:        "invalid engine",
			engine:      "invalid-db",
			expectError: true,
			errorMsg:    "invalid database engine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewEnsureDatabaseAction(tt.engine, tt.version, tt.serviceName, tt.ensureRunning, tt.createDB)
			err := action.Validate()

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestEnsureDatabaseAction_Render(t *testing.T) {
	tests := []struct {
		name          string
		engine        string
		version       string
		serviceName   string
		ensureRunning bool
		createDB      []string
		os            string
		expectError   bool
	}{
		{
			name:          "postgres on macOS",
			engine:        "postgres",
			version:       "15",
			ensureRunning: true,
			createDB:      []string{"myapp"},
			os:            OSMacOS,
			expectError:   false,
		},
		{
			name:          "mysql on Linux",
			engine:        "mysql",
			version:       "8.0",
			ensureRunning: true,
			os:            OSLinux,
			expectError:   false,
		},
		{
			name:        "sqlite on Windows",
			engine:      "sqlite",
			createDB:    []string{"myapp"},
			os:          OSWindows,
			expectError: false,
		},
		{
			name:        "empty engine",
			engine:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewEnsureDatabaseAction(tt.engine, tt.version, tt.serviceName, tt.ensureRunning, tt.createDB)
			ctx := &ActionContext{
				OS:         tt.os,
				Arch:       "x86_64",
				Env:        map[string]string{},
				WorkingDir: "/tmp",
				Timeout:    30 * time.Second,
			}

			commands, err := action.Render(ctx)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(commands) == 0 {
					t.Error("Expected commands but got none")
				}
			}
		})
	}
}

func TestEnsureDatabaseAction_GetDefaultServiceName(t *testing.T) {
	tests := []struct {
		engine   string
		expected string
	}{
		{"postgres", "postgresql"},
		{"mysql", "mysql"},
		{"sqlite", ""},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			action := &EnsureDatabaseAction{engine: tt.engine}
			result := action.getDefaultServiceName()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEnsureDatabaseAction_GetExecutableName(t *testing.T) {
	tests := []struct {
		engine   string
		expected string
	}{
		{"postgres", "psql"},
		{"mysql", "mysql"},
		{"sqlite", "sqlite3"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			action := &EnsureDatabaseAction{engine: tt.engine}
			result := action.getExecutableName()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestEnsureDatabaseAction_GetBrewFormula removed - logic now handled by strategy pattern

// TestEnsureDatabaseAction_GetAptPackages removed - logic now handled by strategy pattern

// TestEnsureDatabaseAction_GetChocoPackages removed - logic now handled by strategy pattern

// TestEnsureDatabaseAction_BrewInstallCommands removed - functionality now handled by strategy pattern

// TestEnsureDatabaseAction_AptInstallCommands removed - functionality now handled by strategy pattern

// TestEnsureDatabaseAction_ChocoInstallCommands removed - functionality now handled by strategy pattern

// TestEnsureDatabaseAction_ServiceCommands removed - functionality now handled by strategy pattern

func TestEnsureDatabaseAction_CreateDatabaseCommands(t *testing.T) {
	tests := []struct {
		name     string
		engine   string
		createDB []string
		expected []string
	}{
		{
			name:     "postgres databases",
			engine:   "postgres",
			createDB: []string{"myapp", "test"},
			expected: []string{"createdb myapp", "createdb test"},
		},
		{
			name:     "mysql databases",
			engine:   "mysql",
			createDB: []string{"myapp", "test"},
			expected: []string{"mysql -e CREATE DATABASE", "mysql -e CREATE DATABASE"},
		},
		{
			name:     "sqlite databases",
			engine:   "sqlite",
			createDB: []string{"myapp", "test"},
			expected: []string{"sqlite3 myapp.db", "sqlite3 test.db"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &EnsureDatabaseAction{
				engine:   tt.engine,
				createDB: tt.createDB,
			}

			commands := action.getCreateDatabaseCommands()
			if len(commands) == 0 {
				t.Fatal("Expected commands but got none")
			}

			for _, expected := range tt.expected {
				found := false
				for _, cmd := range commands {
					if strings.Contains(cmd.Command+" "+strings.Join(cmd.Args, " "), expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected command containing '%s' not found", expected)
				}
			}
		})
	}
}

func TestEnsureDatabaseAction_StrategyBasedCommands(t *testing.T) {
	action := NewEnsureDatabaseAction("postgres", "15", "postgresql", true, []string{"testdb"})
	ctx := &ActionContext{
		OS:         OSMacOS,
		Arch:       "x86_64",
		Env:        map[string]string{},
		WorkingDir: "/tmp",
		Timeout:    30 * time.Second,
	}

	commands, err := action.getInstallCommands(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Should have install, service start, and database creation commands
	if len(commands) < 3 {
		t.Errorf("Expected at least 3 commands, got %d", len(commands))
	}

	// Log the actual commands for debugging
	t.Logf("Generated %d commands:", len(commands))
	for i, cmd := range commands {
		t.Logf("  %d: %s %v", i, cmd.Command, cmd.Args)
	}
}
