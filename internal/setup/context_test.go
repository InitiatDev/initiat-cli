package setup

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("NewContext() failed: %v", err)
	}

	if ctx.OS == "" {
		t.Error("Expected OS to be set")
	}

	if ctx.Arch == "" {
		t.Error("Expected Arch to be set")
	}

	if ctx.WorkingDir == "" {
		t.Error("Expected WorkingDir to be set")
	}

	if ctx.Secrets == nil {
		t.Error("Expected Secrets map to be initialized")
	}

	currentOS := runtime.GOOS
	expectedOS := normalizeOS(currentOS)
	if ctx.OS != expectedOS {
		t.Errorf("Expected OS %s, got %s", expectedOS, ctx.OS)
	}

	currentArch := runtime.GOARCH
	expectedArch := normalizeArch(currentArch)
	if ctx.Arch != expectedArch {
		t.Errorf("Expected Arch %s, got %s", expectedArch, ctx.Arch)
	}

	wd, _ := os.Getwd()
	if ctx.WorkingDir != wd {
		t.Errorf("Expected WorkingDir %s, got %s", wd, ctx.WorkingDir)
	}
}

func TestContext_WithSecrets(t *testing.T) {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("NewContext() failed: %v", err)
	}

	secrets := map[string]string{
		"API_KEY": "secret123",
		"DB_PASS": "password456",
	}

	newCtx := ctx.WithSecrets(secrets)

	if len(newCtx.Secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(newCtx.Secrets))
	}

	if newCtx.Secrets["API_KEY"] != "secret123" {
		t.Errorf("Expected API_KEY=secret123, got %s", newCtx.Secrets["API_KEY"])
	}

	if newCtx.Secrets["DB_PASS"] != "password456" {
		t.Errorf("Expected DB_PASS=password456, got %s", newCtx.Secrets["DB_PASS"])
	}

	if newCtx.OS != ctx.OS {
		t.Error("OS should remain unchanged")
	}

	if newCtx.Arch != ctx.Arch {
		t.Error("Arch should remain unchanged")
	}

	if newCtx.WorkingDir != ctx.WorkingDir {
		t.Error("WorkingDir should remain unchanged")
	}

	existingSecrets := map[string]string{
		"EXISTING": "value1",
	}
	ctxWithExisting := &Context{
		OS:         "linux",
		Arch:       "x86_64",
		WorkingDir: "/tmp",
		Secrets:    existingSecrets,
	}

	newCtx2 := ctxWithExisting.WithSecrets(secrets)

	if len(newCtx2.Secrets) != 3 {
		t.Errorf("Expected 3 secrets (1 existing + 2 new), got %d", len(newCtx2.Secrets))
	}

	if newCtx2.Secrets["EXISTING"] != "value1" {
		t.Error("Existing secret should be preserved")
	}
}

func TestContext_WithWorkingDir_Absolute(t *testing.T) {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("NewContext() failed: %v", err)
	}

	absDir := "/tmp/test"
	newCtx, err := ctx.WithWorkingDir(absDir)
	if err != nil {
		t.Fatalf("WithWorkingDir() failed: %v", err)
	}

	if newCtx.WorkingDir != absDir {
		t.Errorf("Expected WorkingDir %s, got %s", absDir, newCtx.WorkingDir)
	}

	if newCtx.OS != ctx.OS {
		t.Error("OS should remain unchanged")
	}

	if newCtx.Arch != ctx.Arch {
		t.Error("Arch should remain unchanged")
	}
}

func TestContext_WithWorkingDir_Relative(t *testing.T) {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("NewContext() failed: %v", err)
	}

	relDir := "subdir"
	newCtx, err := ctx.WithWorkingDir(relDir)
	if err != nil {
		t.Fatalf("WithWorkingDir() failed: %v", err)
	}

	expectedPath, _ := filepath.Abs(filepath.Join(ctx.WorkingDir, relDir))
	if newCtx.WorkingDir != expectedPath {
		t.Errorf("Expected WorkingDir %s, got %s", expectedPath, newCtx.WorkingDir)
	}

	if !filepath.IsAbs(newCtx.WorkingDir) {
		t.Error("WorkingDir should be absolute")
	}
}

func TestContext_ShouldExecuteStep_NoCondition(t *testing.T) {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("NewContext() failed: %v", err)
	}

	step := &Step{
		Run: "echo test",
	}

	shouldExecute, err := ctx.ShouldExecuteStep(step)
	if err != nil {
		t.Fatalf("ShouldExecuteStep() failed: %v", err)
	}

	if !shouldExecute {
		t.Error("Expected step to execute when no condition is set")
	}
}

func TestContext_ShouldExecuteStep_WithCondition(t *testing.T) {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("NewContext() failed: %v", err)
	}

	currentOS := normalizeOS(runtime.GOOS)

	step := &Step{
		Run: "echo test",
		If:  "os(\"" + currentOS + "\")",
	}

	shouldExecute, err := ctx.ShouldExecuteStep(step)
	if err != nil {
		t.Fatalf("ShouldExecuteStep() failed: %v", err)
	}

	if !shouldExecute {
		t.Error("Expected step to execute when condition matches current OS")
	}

	step.If = "os(\"nonexistent\")"
	shouldExecute, err = ctx.ShouldExecuteStep(step)
	if err != nil {
		t.Fatalf("ShouldExecuteStep() failed: %v", err)
	}

	if shouldExecute {
		t.Error("Expected step to not execute when condition doesn't match")
	}
}

func TestContext_ShouldExecuteStep_ComplexCondition(t *testing.T) {
	ctx, err := NewContext()
	if err != nil {
		t.Fatalf("NewContext() failed: %v", err)
	}

	currentOS := normalizeOS(runtime.GOOS)
	currentArch := normalizeArch(runtime.GOARCH)

	step := &Step{
		Run: "echo test",
		If:  "os(\"" + currentOS + "\") && arch(\"" + currentArch + "\")",
	}

	shouldExecute, err := ctx.ShouldExecuteStep(step)
	if err != nil {
		t.Fatalf("ShouldExecuteStep() failed: %v", err)
	}

	if !shouldExecute {
		t.Error("Expected step to execute when complex condition matches")
	}
}
