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

func TestEnsureDatabaseAction_GetBrewFormula(t *testing.T) {
	tests := []struct {
		engine   string
		expected string
	}{
		{"postgres", "postgresql"},
		{"mysql", "mysql"},
		{"sqlite", "sqlite"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			action := &EnsureDatabaseAction{engine: tt.engine}
			result := action.getBrewFormula()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEnsureDatabaseAction_GetAptPackages(t *testing.T) {
	tests := []struct {
		engine   string
		expected []string
	}{
		{"postgres", []string{"postgresql", "postgresql-contrib"}},
		{"mysql", []string{"mysql-server", "mysql-client"}},
		{"sqlite", []string{"sqlite3"}},
		{"unknown", []string{"unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			action := &EnsureDatabaseAction{engine: tt.engine}
			result := action.getAptPackages()
			if !stringSlicesEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEnsureDatabaseAction_GetChocoPackages(t *testing.T) {
	tests := []struct {
		engine   string
		expected []string
	}{
		{"postgres", []string{"postgresql"}},
		{"mysql", []string{"mysql"}},
		{"sqlite", []string{"sqlite"}},
		{"unknown", []string{"unknown"}},
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			action := &EnsureDatabaseAction{engine: tt.engine}
			result := action.getChocoPackages()
			if !stringSlicesEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEnsureDatabaseAction_BrewInstallCommands(t *testing.T) {
	action := &EnsureDatabaseAction{
		engine:  "postgres",
		version: "15",
	}

	commands := action.getBrewInstallCommands()
	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Check that we have the expected commands
	expectedCommands := []string{"which", "brew install"}
	for _, expected := range expectedCommands {
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
}

func TestEnsureDatabaseAction_AptInstallCommands(t *testing.T) {
	action := &EnsureDatabaseAction{
		engine:  "postgres",
		version: "15",
	}

	commands := action.getAptInstallCommands()
	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Check that we have the expected commands
	expectedCommands := []string{"which", "apt update", "apt install"}
	for _, expected := range expectedCommands {
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
}

func TestEnsureDatabaseAction_ChocoInstallCommands(t *testing.T) {
	action := &EnsureDatabaseAction{
		engine:  "postgres",
		version: "15",
	}

	commands := action.getChocoInstallCommands()
	if len(commands) == 0 {
		t.Fatal("Expected commands but got none")
	}

	// Check that we have the expected commands
	expectedCommands := []string{"where", "choco install"}
	for _, expected := range expectedCommands {
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
}

func TestEnsureDatabaseAction_ServiceCommands(t *testing.T) {
	tests := []struct {
		name     string
		engine   string
		os       string
		expected []string
	}{
		{
			name:     "postgres on macOS",
			engine:   "postgres",
			os:       OSMacOS,
			expected: []string{"brew services start"},
		},
		{
			name:     "mysql on Linux",
			engine:   "mysql",
			os:       OSLinux,
			expected: []string{"systemctl start", "systemctl enable"},
		},
		{
			name:     "postgres on Windows",
			engine:   "postgres",
			os:       OSWindows,
			expected: []string{"net start"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &EnsureDatabaseAction{
				engine:        tt.engine,
				serviceName:   tt.engine,
				ensureRunning: true,
			}

			commands := action.getServiceCommands(tt.os)
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

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
