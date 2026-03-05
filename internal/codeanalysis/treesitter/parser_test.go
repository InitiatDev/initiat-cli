package treesitter

import (
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tsgo "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

func TestParserHandle_ParseAndClose(t *testing.T) {
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
	if root.Type() == "" {
		t.Fatalf("expected root type")
	}
	if !root.IsNamed() {
		t.Fatalf("expected named root")
	}
}
