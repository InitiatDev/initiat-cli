package elixir

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	tselixir "github.com/tree-sitter/tree-sitter-elixir/bindings/go"
)

func grammar() *sitter.Language {
	return sitter.NewLanguage(tselixir.Language())
}
