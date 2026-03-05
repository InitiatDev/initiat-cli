package golang

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	tsgo "github.com/tree-sitter/tree-sitter-go/bindings/go"
)

func grammar() *sitter.Language {
	return sitter.NewLanguage(tsgo.Language())
}
