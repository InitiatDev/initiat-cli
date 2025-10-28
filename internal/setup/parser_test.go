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
  - name: "Install package manager"
    ensure_package_manager:
      type: auto

  - name: "Install git"
    ensure_tool:
      name: git
      version: ">=2.34"
      install:
        brew: { formula: git }
        apt: { packages: [git], update: true }

provision:
  - name: "Install Node.js"
    ensure_runtime:
      name: node
      version: "18.x"
      manager: { asdf: true }

setup:
  - name: "Install dependencies"
    run: npm install
    timeout: "5m"
    cwd: "./frontend"

  - name: "Print success"
    print: "Setup complete!"

verify:
  - name: "Check Node version"
    assert_command: node --version

  - name: "Check API health"
    assert_http:
      url: "http://localhost:3000/health"
      expect_status: 200

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

	if config.Bootstrap[0].Name != "Install package manager" {
		t.Errorf("Expected first bootstrap step name 'Install package manager', got '%s'", config.Bootstrap[0].Name)
	}

	if config.Bootstrap[0].EnsurePackageManager == nil {
		t.Fatal("Expected first bootstrap step to have ensure_package_manager")
	}

	if config.Bootstrap[0].EnsurePackageManager.Type != "auto" {
		t.Errorf("Expected package manager type 'auto', got '%s'", config.Bootstrap[0].EnsurePackageManager.Type)
	}

	if config.Bootstrap[1].EnsureTool == nil {
		t.Fatal("Expected second bootstrap step to have ensure_tool")
	}

	if config.Bootstrap[1].EnsureTool.Name != "git" {
		t.Errorf("Expected tool name 'git', got '%s'", config.Bootstrap[1].EnsureTool.Name)
	}

	if len(config.Provision) != 1 {
		t.Errorf("Expected 1 provision step, got %d", len(config.Provision))
	}

	if config.Provision[0].EnsureRuntime == nil {
		t.Fatal("Expected provision step to have ensure_runtime")
	}

	if config.Provision[0].EnsureRuntime.Name != "node" {
		t.Errorf("Expected runtime name 'node', got '%s'", config.Provision[0].EnsureRuntime.Name)
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

	if config.Verify[0].AssertCommand != "node --version" {
		t.Errorf("Expected first verify step assert_command 'node --version', got '%s'", config.Verify[0].AssertCommand)
	}

	if config.Verify[1].AssertHTTP == nil {
		t.Fatal("Expected second verify step to have assert_http")
	}

	if config.Verify[1].AssertHTTP.URL != "http://localhost:3000/health" {
		t.Errorf("Expected HTTP URL 'http://localhost:3000/health', got '%s'", config.Verify[1].AssertHTTP.URL)
	}

	if config.Verify[1].AssertHTTP.ExpectStatus != 200 {
		t.Errorf("Expected HTTP status 200, got %d", config.Verify[1].AssertHTTP.ExpectStatus)
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

func TestParse_ComplexToolInstall(t *testing.T) {
	yaml := `version: 1
setup:
  - name: "Install complex tool"
    ensure_tool:
      name: docker
      version: ">=20.0"
      install:
        brew: { formula: docker }
        apt: { packages: [docker.io], update: true }
        choco: { packages: [docker-desktop] }
        fallback_url: "https://docker.com/download"
        checksum: "sha256:abc123..."`

	config, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(config.Setup) != 1 {
		t.Fatalf("Expected 1 setup step, got %d", len(config.Setup))
	}

	tool := config.Setup[0].EnsureTool
	if tool == nil {
		t.Fatal("Expected ensure_tool to be set")
	}

	if tool.Name != "docker" {
		t.Errorf("Expected tool name 'docker', got '%s'", tool.Name)
	}

	if tool.Version != ">=20.0" {
		t.Errorf("Expected version '>=20.0', got '%s'", tool.Version)
	}

	if tool.Install == nil {
		t.Fatal("Expected install config to be set")
	}

	if tool.Install.Brew.Formula != "docker" {
		t.Errorf("Expected brew formula 'docker', got '%s'", tool.Install.Brew.Formula)
	}

	if len(tool.Install.Apt.Packages) != 1 {
		t.Errorf("Expected 1 apt package, got %d", len(tool.Install.Apt.Packages))
	}

	if tool.Install.Apt.Packages[0] != "docker.io" {
		t.Errorf("Expected apt package 'docker.io', got '%s'", tool.Install.Apt.Packages[0])
	}

	if !tool.Install.Apt.Update {
		t.Error("Expected apt update to be true")
	}

	if len(tool.Install.Choco.Packages) != 1 {
		t.Errorf("Expected 1 choco package, got %d", len(tool.Install.Choco.Packages))
	}

	if tool.Install.Choco.Packages[0] != "docker-desktop" {
		t.Errorf("Expected choco package 'docker-desktop', got '%s'", tool.Install.Choco.Packages[0])
	}

	if tool.Install.FallbackURL != "https://docker.com/download" {
		t.Errorf("Expected fallback URL 'https://docker.com/download', got '%s'", tool.Install.FallbackURL)
	}

	if tool.Install.Checksum != "sha256:abc123..." {
		t.Errorf("Expected checksum 'sha256:abc123...', got '%s'", tool.Install.Checksum)
	}
}

func TestParse_ComplexRuntime(t *testing.T) {
	yaml := `version: 1
provision:
  - name: "Install Elixir with fallbacks"
    ensure_runtime:
      name: elixir
      version: "1.16.x"
      manager: { asdf: true }
      fallback_installers:
        - brew: { formula: elixir }
        - apt: { packages: [elixir] }
        - choco: { packages: [elixir] }`

	config, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(config.Provision) != 1 {
		t.Fatalf("Expected 1 provision step, got %d", len(config.Provision))
	}

	runtime := config.Provision[0].EnsureRuntime
	if runtime == nil {
		t.Fatal("Expected ensure_runtime to be set")
	}

	if runtime.Name != "elixir" {
		t.Errorf("Expected runtime name 'elixir', got '%s'", runtime.Name)
	}

	if runtime.Version != "1.16.x" {
		t.Errorf("Expected version '1.16.x', got '%s'", runtime.Version)
	}

	if runtime.Manager == nil {
		t.Fatal("Expected manager to be set")
	}

	if !runtime.Manager.Asdf {
		t.Error("Expected asdf manager to be true")
	}

	if len(runtime.FallbackInstallers) != 3 {
		t.Errorf("Expected 3 fallback installers, got %d", len(runtime.FallbackInstallers))
	}

	if runtime.FallbackInstallers[0].Brew.Formula != "elixir" {
		t.Errorf("Expected first fallback brew formula 'elixir', got '%s'", runtime.FallbackInstallers[0].Brew.Formula)
	}

	if runtime.FallbackInstallers[1].Apt.Packages[0] != "elixir" {
		t.Errorf("Expected second fallback apt package 'elixir', got '%s'", runtime.FallbackInstallers[1].Apt.Packages[0])
	}

	if runtime.FallbackInstallers[2].Choco.Packages[0] != "elixir" {
		t.Errorf("Expected third fallback choco package 'elixir', got '%s'", runtime.FallbackInstallers[2].Choco.Packages[0])
	}
}

func TestParse_DatabaseConfig(t *testing.T) {
	yaml := `version: 1
provision:
  - name: "Setup PostgreSQL"
    ensure_database:
      engine: postgres
      version: "15"
      service_name: "postgres"
      ensure_running: true
      create_db: ["app_dev", "app_test", "app_prod"]`

	config, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(config.Provision) != 1 {
		t.Fatalf("Expected 1 provision step, got %d", len(config.Provision))
	}

	db := config.Provision[0].EnsureDatabase
	if db == nil {
		t.Fatal("Expected ensure_database to be set")
	}

	if db.Engine != "postgres" {
		t.Errorf("Expected engine 'postgres', got '%s'", db.Engine)
	}

	if db.Version != "15" {
		t.Errorf("Expected version '15', got '%s'", db.Version)
	}

	if db.ServiceName != "postgres" {
		t.Errorf("Expected service_name 'postgres', got '%s'", db.ServiceName)
	}

	if !db.EnsureRunning {
		t.Error("Expected ensure_running to be true")
	}

	if len(db.CreateDB) != 3 {
		t.Errorf("Expected 3 databases to create, got %d", len(db.CreateDB))
	}

	expectedDBs := []string{"app_dev", "app_test", "app_prod"}
	for i, expected := range expectedDBs {
		if db.CreateDB[i] != expected {
			t.Errorf("Expected database %d '%s', got '%s'", i, expected, db.CreateDB[i])
		}
	}
}

func TestParse_HTTPAssertion(t *testing.T) {
	yaml := `version: 1
verify:
  - name: "Check API health"
    assert_http:
      url: "http://localhost:8080/api/health"
      expect_status: 200
      retries:
        attempts: 30
        backoff: "2s"`

	config, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(config.Verify) != 1 {
		t.Fatalf("Expected 1 verify step, got %d", len(config.Verify))
	}

	http := config.Verify[0].AssertHTTP
	if http == nil {
		t.Fatal("Expected assert_http to be set")
	}

	if http.URL != "http://localhost:8080/api/health" {
		t.Errorf("Expected URL 'http://localhost:8080/api/health', got '%s'", http.URL)
	}

	if http.ExpectStatus != 200 {
		t.Errorf("Expected status 200, got %d", http.ExpectStatus)
	}

	if http.Retries == nil {
		t.Fatal("Expected retries to be set")
	}

	if http.Retries.Attempts != 30 {
		t.Errorf("Expected 30 retry attempts, got %d", http.Retries.Attempts)
	}

	if http.Retries.Backoff != "2s" {
		t.Errorf("Expected retry backoff '2s', got '%s'", http.Retries.Backoff)
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
