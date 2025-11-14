package setup

import (
	"strings"
	"testing"
)

func TestParse_ValidMinimal(t *testing.T) {
	yaml := `version: 1
name: "Test Project"`

	config, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}

	if config.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", config.Name)
	}
}

func TestParse_ValidComplete(t *testing.T) {
	yaml := `version: 1
name: "Complete Setup"
description: "A complete setup example"

matrix:
  os: [macos, linux]
  arch: [x86_64, arm64]

defaults:
  timeout: "10m"
  shell: "bash"
  continue_on_error: true
  cwd: "./app"

env:
  NODE_ENV: development
  secrets: [DATABASE_URL, API_KEY]

bootstrap:
  - name: "Install git (macOS)"
    if: os("macos")
    run: brew install git

  - name: "Install git (Linux)"
    if: os("linux")
    run: sudo apt-get install -y git

provision:
  - name: "Install Node.js"
    run: |
      asdf plugin add nodejs || true
      asdf install nodejs 18.x
      asdf global nodejs 18.x

setup:
  - name: "Install dependencies"
    run: npm install
    timeout: "5m"
    cwd: "./frontend"

  - name: "Print success"
    print: "Setup complete!"

verify:
  - name: "Check Node version"
    run: node --version || exit 1

  - name: "Check API health"
    run: curl -f http://localhost:3000/health || exit 1

post:
  - name: "Final message"
    print: "Ready to develop!"`

	config, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}

	if config.Name != "Complete Setup" {
		t.Errorf("Expected name 'Complete Setup', got '%s'", config.Name)
	}

	if config.Description != "A complete setup example" {
		t.Errorf("Expected description 'A complete setup example', got '%s'", config.Description)
	}

	if config.Matrix == nil {
		t.Fatal("Expected matrix to be set")
	}

	if len(config.Matrix.OS) != 2 {
		t.Errorf("Expected 2 OS values, got %d", len(config.Matrix.OS))
	}

	if config.Matrix.OS[0] != "macos" {
		t.Errorf("Expected first OS 'macos', got '%s'", config.Matrix.OS[0])
	}

	if config.Defaults == nil {
		t.Fatal("Expected defaults to be set")
	}

	if config.Defaults.Timeout != "10m" {
		t.Errorf("Expected timeout '10m', got '%s'", config.Defaults.Timeout)
	}

	if config.Env == nil {
		t.Fatal("Expected env to be set")
	}

	if config.Env.Vars["NODE_ENV"] != "development" {
		t.Errorf("Expected NODE_ENV 'development', got '%s'", config.Env.Vars["NODE_ENV"])
	}

	if len(config.Env.Secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(config.Env.Secrets))
	}

	if len(config.Bootstrap) != 2 {
		t.Errorf("Expected 2 bootstrap steps, got %d", len(config.Bootstrap))
	}

	if config.Bootstrap[0].Name != "Install git (macOS)" {
		t.Errorf("Expected first bootstrap step name 'Install git (macOS)', got '%s'", config.Bootstrap[0].Name)
	}

	if config.Bootstrap[0].Run == "" {
		t.Fatal("Expected first bootstrap step to have run")
	}

	if !strings.Contains(config.Bootstrap[0].Run, "brew install git") {
		t.Errorf("Expected first bootstrap step to contain 'brew install git', got '%s'", config.Bootstrap[0].Run)
	}

	if config.Bootstrap[1].Name != "Install git (Linux)" {
		t.Errorf("Expected second bootstrap step name 'Install git (Linux)', got '%s'", config.Bootstrap[1].Name)
	}

	if config.Bootstrap[1].Run == "" {
		t.Fatal("Expected second bootstrap step to have run")
	}

	if len(config.Provision) != 1 {
		t.Errorf("Expected 1 provision step, got %d", len(config.Provision))
	}

	if config.Provision[0].Run == "" {
		t.Fatal("Expected provision step to have run")
	}

	if !strings.Contains(config.Provision[0].Run, "asdf") {
		t.Errorf("Expected provision step to contain 'asdf', got '%s'", config.Provision[0].Run)
	}

	if len(config.Setup) != 2 {
		t.Errorf("Expected 2 setup steps, got %d", len(config.Setup))
	}

	if config.Setup[0].Run != "npm install" {
		t.Errorf("Expected first setup step run 'npm install', got '%s'", config.Setup[0].Run)
	}

	if config.Setup[0].Timeout != "5m" {
		t.Errorf("Expected first setup step timeout '5m', got '%s'", config.Setup[0].Timeout)
	}

	if config.Setup[0].CWD != "./frontend" {
		t.Errorf("Expected first setup step cwd './frontend', got '%s'", config.Setup[0].CWD)
	}

	if config.Setup[1].Print != "Setup complete!" {
		t.Errorf("Expected second setup step print 'Setup complete!', got '%s'", config.Setup[1].Print)
	}

	if len(config.Verify) != 2 {
		t.Errorf("Expected 2 verify steps, got %d", len(config.Verify))
	}

	if config.Verify[0].Run == "" {
		t.Fatal("Expected first verify step to have run")
	}

	if !strings.Contains(config.Verify[0].Run, "node --version") {
		t.Errorf("Expected first verify step to contain 'node --version', got '%s'", config.Verify[0].Run)
	}

	if config.Verify[1].Run == "" {
		t.Fatal("Expected second verify step to have run")
	}

	if !strings.Contains(config.Verify[1].Run, "curl") {
		t.Errorf("Expected second verify step to contain 'curl', got '%s'", config.Verify[1].Run)
	}

	if len(config.Post) != 1 {
		t.Errorf("Expected 1 post step, got %d", len(config.Post))
	}

	if config.Post[0].Print != "Ready to develop!" {
		t.Errorf("Expected post step print 'Ready to develop!', got '%s'", config.Post[0].Print)
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	yaml := `version: 1
name: "Test"
invalid: [`

	_, err := Parse([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got none")
	}

	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("Expected error to contain 'failed to parse YAML', got: %v", err)
	}
}

func TestParse_EmptyYAML(t *testing.T) {
	config, err := Parse([]byte(""))
	if err != nil {
		t.Fatalf("Expected no error for empty YAML, got: %v", err)
	}

	if config.Version != 0 {
		t.Errorf("Expected version 0 for empty YAML, got %d", config.Version)
	}
}

func TestParseFromReader(t *testing.T) {
	yaml := `version: 1
name: "Reader Test"`

	reader := strings.NewReader(yaml)
	config, err := ParseFromReader(reader)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}

	if config.Name != "Reader Test" {
		t.Errorf("Expected name 'Reader Test', got '%s'", config.Name)
	}
}

func TestParseFile_ValidFile(t *testing.T) {
	config, err := ParseFile("../../test/fixtures/setup_examples/minimal.yml")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}

	if config.Name != "Minimal Setup" {
		t.Errorf("Expected name 'Minimal Setup', got '%s'", config.Name)
	}
}

func TestParseFile_ComplexFile(t *testing.T) {
	config, err := ParseFile("../../test/fixtures/setup_examples/phoenix_basic.yml")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Version != 1 {
		t.Errorf("Expected version 1, got %d", config.Version)
	}

	if config.Name != "Phoenix Local Setup" {
		t.Errorf("Expected name 'Phoenix Local Setup', got '%s'", config.Name)
	}

	if config.Description != "Zero-to-ready Phoenix development environment" {
		t.Errorf("Expected description 'Zero-to-ready Phoenix development environment', got '%s'", config.Description)
	}

	if config.Matrix == nil {
		t.Fatal("Expected matrix to be set")
	}

	if len(config.Matrix.OS) != 3 {
		t.Errorf("Expected 3 OS values, got %d", len(config.Matrix.OS))
	}

	if len(config.Bootstrap) != 2 {
		t.Errorf("Expected 2 bootstrap steps, got %d", len(config.Bootstrap))
	}

	if len(config.Provision) != 3 {
		t.Errorf("Expected 3 provision steps, got %d", len(config.Provision))
	}

	if len(config.Setup) != 5 {
		t.Errorf("Expected 5 setup steps, got %d", len(config.Setup))
	}

	if len(config.Verify) != 2 {
		t.Errorf("Expected 2 verify steps, got %d", len(config.Verify))
	}

	if len(config.Post) != 1 {
		t.Errorf("Expected 1 post step, got %d", len(config.Post))
	}
}

func TestParseFile_InvalidFile(t *testing.T) {
	_, err := ParseFile("../../test/fixtures/setup_examples/invalid_version.yml")
	if err != nil {
		t.Fatalf("Expected no error for invalid version file, got: %v", err)
	}

	_, err = ParseFile("../../test/fixtures/setup_examples/invalid_action.yml")
	if err != nil {
		t.Fatalf("Expected no error for invalid action file, got: %v", err)
	}
}

func TestParseFile_NonExistentFile(t *testing.T) {
	_, err := ParseFile("test/fixtures/setup_examples/nonexistent.yml")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got none")
	}

	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected error to contain 'failed to read file', got: %v", err)
	}
}
