package ruby

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	tsrb "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
)

func grammar() *sitter.Language {
	return sitter.NewLanguage(tsrb.Language())
}
