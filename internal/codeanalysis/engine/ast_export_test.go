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
	js "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/javascript"
)

func TestExportAST_JSONL_AutoDetect(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	jsFile := filepath.Join(dir, "a.js")

	if err := os.WriteFile(goFile, []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatalf("write go: %v", err)
	}
	if err := os.WriteFile(jsFile, []byte("const x = 1;\n"), 0o600); err != nil {
		t.Fatalf("write js: %v", err)
	}

	reg := engine.NewRegistry()
	if err := reg.Register(golang.New()); err != nil {
		t.Fatalf("register go: %v", err)
	}
	if err := reg.Register(js.New()); err != nil {
		t.Fatalf("register js: %v", err)
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
	if len(lines) != 2 {
		t.Fatalf("expected 2 jsonl lines, got %d: %q", len(lines), buf.String())
	}

	var docs []ast.Document
	for _, line := range lines {
		var d ast.Document
		if err := json.Unmarshal([]byte(line), &d); err != nil {
			t.Fatalf("unmarshal: %v\n%s", err, line)
		}
		docs = append(docs, d)
	}

	gotLangs := map[string]bool{}
	gotPaths := map[string]bool{}
	for _, d := range docs {
		gotLangs[d.Language] = true
		gotPaths[filepath.Base(d.Path)] = true
		if d.Version != "ast-v1" {
			t.Fatalf("unexpected version: %q", d.Version)
		}
		if d.Root.Type == "" {
			t.Fatalf("missing root type for %s", d.Path)
		}
	}

	if !gotLangs["go"] || !gotLangs["javascript"] {
		t.Fatalf("unexpected langs: %+v", gotLangs)
	}
	if !gotPaths["main.go"] || !gotPaths["a.js"] {
		t.Fatalf("unexpected paths: %+v", gotPaths)
	}
}
