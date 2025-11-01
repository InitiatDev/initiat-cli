package setup

import (
	"testing"
	"time"
)

func TestRender_MinimalConfig(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Bootstrap: []Step{
			{
				Run: "echo hello",
			},
		},
	}

	ctx := &RenderContext{
		OS:         "macos",
		Arch:       "x86_64",
		WorkingDir: "/tmp/test",
		Shell:      "/bin/bash",
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(plan.Commands))
	}

	if plan.Commands[0].Command != "echo" {
		t.Errorf("Expected command 'echo', got '%s'", plan.Commands[0].Command)
	}

	if plan.Summary.TotalCommands != 1 {
		t.Errorf("Expected total commands 1, got %d", plan.Summary.TotalCommands)
	}

	if plan.Summary.TotalSteps != 1 {
		t.Errorf("Expected total steps 1, got %d", plan.Summary.TotalSteps)
	}
}

func TestRender_MultiplePhases(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Bootstrap: []Step{
			{Print: "Starting bootstrap"},
		},
		Provision: []Step{
			{Run: "install-tool"},
		},
		Setup: []Step{
			{Run: "setup-project"},
		},
	}

	ctx := &RenderContext{
		OS:         "linux",
		Arch:       "x86_64",
		WorkingDir: "/tmp/test",
		Shell:      "/bin/bash",
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) != 3 {
		t.Fatalf("Expected 3 commands, got %d", len(plan.Commands))
	}

	phases := []string{"bootstrap", "provision", "setup"}
	for i, expectedPhase := range phases {
		if plan.Commands[i].Phase != expectedPhase {
			t.Errorf("Command %d: expected phase '%s', got '%s'", i, expectedPhase, plan.Commands[i].Phase)
		}
	}

	if len(plan.Summary.Phases) != 3 {
		t.Errorf("Expected 3 phases in summary, got %d", len(plan.Summary.Phases))
	}
}

func TestRender_WithConditions(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Bootstrap: []Step{
			{
				Name: "macOS only",
				If:   `os == "macos"`,
				Run:  "macos-command",
			},
			{
				Name: "linux only",
				If:   `os == "linux"`,
				Run:  "linux-command",
			},
			{
				Name: "always",
				Run:  "always-command",
			},
		},
	}

	ctx := &RenderContext{
		OS:         "macos",
		Arch:       "x86_64",
		WorkingDir: "/tmp/test",
		Shell:      "/bin/bash",
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) != 2 {
		t.Fatalf("Expected 2 commands (macOS + always), got %d", len(plan.Commands))
	}

	if plan.Commands[0].Command != "macos-command" && plan.Commands[1].Command != "macos-command" {
		t.Error("Expected to find 'macos-command'")
	}

	if plan.Commands[0].Command != "always-command" && plan.Commands[1].Command != "always-command" {
		t.Error("Expected to find 'always-command'")
	}
}

func TestRender_WithDefaults(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Defaults: &Defaults{
			Timeout:         "30m",
			CWD:             "project",
			ContinueOnError: true,
		},
		Bootstrap: []Step{
			{
				Run: "test-command",
			},
		},
	}

	ctx := &RenderContext{
		OS:         "linux",
		Arch:       "x86_64",
		WorkingDir: "/tmp",
		Shell:      "/bin/bash",
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(plan.Commands))
	}

	expectedTimeout := 30 * time.Minute
	if plan.Commands[0].Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, plan.Commands[0].Timeout)
	}

	expectedWorkingDir := "/tmp/project"
	if plan.Commands[0].WorkingDir != expectedWorkingDir {
		t.Errorf("Expected working dir %s, got %s", expectedWorkingDir, plan.Commands[0].WorkingDir)
	}

	if !plan.Commands[0].ContinueOnError {
		t.Error("Expected ContinueOnError to be true")
	}
}

func TestRender_WithStepOverrides(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Defaults: &Defaults{
			Timeout: "15m",
			CWD:     "project",
		},
		Bootstrap: []Step{
			{
				Name:    "override-timeout",
				Timeout: "5m",
				CWD:     "override",
				Run:     "test-command",
			},
		},
	}

	ctx := &RenderContext{
		OS:         "linux",
		Arch:       "x86_64",
		WorkingDir: "/tmp",
		Shell:      "/bin/bash",
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(plan.Commands))
	}

	expectedTimeout := 5 * time.Minute
	if plan.Commands[0].Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, plan.Commands[0].Timeout)
	}

	expectedWorkingDir := "/tmp/override"
	if plan.Commands[0].WorkingDir != expectedWorkingDir {
		t.Errorf("Expected working dir %s, got %s", expectedWorkingDir, plan.Commands[0].WorkingDir)
	}
}

func TestRender_WithEnvironmentVariables(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Env: &Environment{
			Vars: map[string]string{
				"GLOBAL_VAR": "global_value",
			},
		},
		Bootstrap: []Step{
			{
				Env: map[string]string{
					"STEP_VAR": "step_value",
				},
				Run: "test-command",
			},
		},
	}

	ctx := &RenderContext{
		OS:         "linux",
		Arch:       "x86_64",
		WorkingDir: "/tmp",
		Shell:      "/bin/bash",
		GlobalEnv: map[string]string{
			"CONTEXT_VAR": "context_value",
		},
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(plan.Commands))
	}

	env := plan.Commands[0].Env
	if env["CONTEXT_VAR"] != "context_value" {
		t.Errorf("Expected CONTEXT_VAR=context_value, got %s", env["CONTEXT_VAR"])
	}

	if env["GLOBAL_VAR"] != "global_value" {
		t.Errorf("Expected GLOBAL_VAR=global_value, got %s", env["GLOBAL_VAR"])
	}

	if env["STEP_VAR"] != "step_value" {
		t.Errorf("Expected STEP_VAR=step_value, got %s", env["STEP_VAR"])
	}
}

func TestRender_WithMatrixMatching(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Matrix: &Matrix{
			OS:   []string{"macos", "linux"},
			Arch: []string{"x86_64"},
		},
		Bootstrap: []Step{
			{Run: "test-command"},
		},
	}

	ctx := &RenderContext{
		OS:         "macos",
		Arch:       "x86_64",
		WorkingDir: "/tmp",
		Shell:      "/bin/bash",
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(plan.Commands))
	}
}

func TestRender_WithMatrixNoMatch(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Matrix: &Matrix{
			OS:   []string{"linux"},
			Arch: []string{"x86_64"},
		},
		Bootstrap: []Step{
			{Run: "test-command"},
		},
	}

	ctx := &RenderContext{
		OS:         "macos",
		Arch:       "x86_64",
		WorkingDir: "/tmp",
		Shell:      "/bin/bash",
	}

	_, err := Render(config, ctx)
	if err == nil {
		t.Fatal("Expected error for non-matching matrix")
	}

	if err.Error() == "" {
		t.Error("Expected error message")
	}
}

func TestRender_WithSecrets(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Bootstrap: []Step{
			{
				EnvFromSecrets: []string{"API_KEY"},
				Run:            "test-command",
			},
		},
	}

	ctx := &RenderContext{
		OS:         "linux",
		Arch:       "x86_64",
		WorkingDir: "/tmp",
		Shell:      "/bin/bash",
		Secrets: map[string]string{
			"API_KEY": "secret-value",
		},
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(plan.Commands))
	}

	if plan.Commands[0].Env["API_KEY"] != "secret-value" {
		t.Errorf("Expected API_KEY=secret-value in env, got %s", plan.Commands[0].Env["API_KEY"])
	}
}

func TestRender_WithRetries(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Bootstrap: []Step{
			{
				Retries: &Retries{
					Attempts: 3,
					Backoff:  "5s",
				},
				Run: "test-command",
			},
		},
	}

	ctx := &RenderContext{
		OS:         "linux",
		Arch:       "x86_64",
		WorkingDir: "/tmp",
		Shell:      "/bin/bash",
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(plan.Commands))
	}

	retries := plan.Commands[0].Retries
	if retries == nil {
		t.Fatal("Expected retries to be set")
	}

	if retries.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", retries.Attempts)
	}

	if retries.Backoff != 5*time.Second {
		t.Errorf("Expected 5s backoff, got %v", retries.Backoff)
	}
}

func TestRender_InvalidConfig(t *testing.T) {
	config := &SetupConfig{
		Version: 2,
		Bootstrap: []Step{
			{Run: "test-command"},
		},
	}

	ctx := &RenderContext{
		OS:         "linux",
		Arch:       "x86_64",
		WorkingDir: "/tmp",
		Shell:      "/bin/bash",
	}

	_, err := Render(config, ctx)
	if err == nil {
		t.Fatal("Expected error for invalid version")
	}
}

func TestRender_AllActionTypes(t *testing.T) {
	config := &SetupConfig{
		Version: 1,
		Bootstrap: []Step{
			{Print: "Starting setup"},
			{Run: "echo test"},
			{AssertCommand: "which git"},
			{
				EnsurePackageManager: &EnsurePackageManager{
					Type: "brew",
				},
			},
			{
				EnsureTool: &EnsureTool{
					Name:    "node",
					Version: "18.0.0",
					Install: &ToolInstallConfig{
						Brew: &BrewInstall{Formula: "node"},
					},
				},
			},
			{
				EnsureRuntime: &EnsureRuntime{
					Name:    "node",
					Version: "18.0.0",
					Manager: &RuntimeManager{Asdf: true},
				},
			},
			{
				EnsureDatabase: &EnsureDatabase{
					Engine: "postgres",
				},
			},
			{
				AssertHTTP: &AssertHTTP{
					URL:          "https://example.com",
					ExpectStatus: 200,
				},
			},
		},
	}

	ctx := &RenderContext{
		OS:             "macos",
		Arch:           "x86_64",
		WorkingDir:     "/tmp/test",
		Shell:          "/bin/bash",
		DefaultTimeout: 30 * time.Minute,
	}

	plan, err := Render(config, ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plan.Commands) == 0 {
		t.Fatal("Expected commands to be generated")
	}

	if plan.Summary.TotalSteps != 8 {
		t.Errorf("Expected 8 steps, got %d", plan.Summary.TotalSteps)
	}
}
