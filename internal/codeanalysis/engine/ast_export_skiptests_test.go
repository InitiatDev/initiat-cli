package engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/ast"
	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/engine"
	golang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/go"
)

func TestExportAST_SkipsTestFilesPerLanguageDetector(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main_test.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("write main_test.go: %v", err)
	}

	reg := engine.NewRegistry()
	if err := reg.Register(golang.New()); err != nil {
		t.Fatalf("register go: %v", err)
	}

	var buf bytes.Buffer
	_, err := engine.ExportAST(context.Background(), reg, dir, engine.ExportASTOptions{
		Lang:      "auto",
		Recursive: true,
		Format:    ast.FormatJSONL,
		MaxBytes:  1024 * 1024,
	}, &buf)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 doc, got %d: %q", len(lines), buf.String())
	}

	var doc ast.Document
	if err := json.Unmarshal([]byte(lines[0]), &doc); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, lines[0])
	}
	if filepath.Base(doc.Path) != "main.go" {
		t.Fatalf("unexpected path: %s", doc.Path)
	}
}
