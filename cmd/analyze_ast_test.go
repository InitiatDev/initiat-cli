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

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	if err := runAnalyzeAST(analyzeCmd, []string{file}); err != nil {
		t.Fatalf("run: %v", err)
	}

	if gotStdout := capture.GetOutput(); strings.TrimSpace(gotStdout) != "" {
		t.Fatalf("expected no stdout output, got: %s", gotStdout)
	}

	outPath := filepath.Join(dir, ".initiat", "code", "ast-v1.jsonl")
	out, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(out), `"version":"ast-v1"`) {
		t.Fatalf("expected ast-v1 in file output, got: %s", string(out))
	}
	if !strings.Contains(string(out), `"language":"go"`) {
		t.Fatalf("expected go language in file output, got: %s", string(out))
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
