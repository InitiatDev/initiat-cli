package javascript

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	tsjs "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
)

func grammar() *sitter.Language {
	return sitter.NewLanguage(tsjs.Language())
}
