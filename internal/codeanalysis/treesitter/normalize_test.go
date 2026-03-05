package treesitter

import (
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/ast"
	sitter "github.com/tree-sitter/go-tree-sitter"
	tsgo "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

func TestNormalize_BuildsASTNodeTree(t *testing.T) {
	t.Parallel()

	lang := sitter.NewLanguage(tsgo.Language())
	h, err := NewParserHandle(lang)
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	defer h.Close()

	tree, err := h.Parse([]byte("package main\nfunc main() {}\n"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	defer tree.Close()

	root := tree.Root()
	out := Normalize(root)

	if out.Type != "source_file" {
		t.Fatalf("unexpected root type: %q", out.Type)
	}
	if !out.Named {
		t.Fatalf("expected named root")
	}
	if out.Range.End.Byte == 0 {
		t.Fatalf("expected non-zero end byte: %+v", out.Range)
	}
	if len(out.Children) == 0 {
		t.Fatalf("expected children")
	}

	_ = ast.Document{Version: "ast-v1", Language: "go", Path: "x.go", Root: out}
}
