package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/testutil"
)

func TestRunAnalyzeAST_Go_JSONL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	analyzeASTLang = "go"
	analyzeASTFormat = "jsonl"
	analyzeASTOutput = ""
	analyzeASTRecursive = true
	analyzeASTFailOnErr = true
	analyzeASTMaxBytes = 1024 * 1024

	capture := testutil.CaptureStdout()
	defer capture.Restore()

	if err := runAnalyzeAST(analyzeCmd, []string{file}); err != nil {
		t.Fatalf("run: %v", err)
	}

	out := capture.GetOutput()
	if !strings.Contains(out, `"version":"ast-v1"`) {
		t.Fatalf("expected ast-v1 in output, got: %s", out)
	}
	if !strings.Contains(out, `"language":"go"`) {
		t.Fatalf("expected go language in output, got: %s", out)
	}
}

func TestCLI_CodeAnalyze_CommandExists(t *testing.T) {
	t.Parallel()

	cmd, _, err := rootCmd.Find([]string{"code", "analyze"})
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if cmd == nil || cmd.Name() != "analyze" {
		t.Fatalf("unexpected cmd: %#v", cmd)
	}
}
