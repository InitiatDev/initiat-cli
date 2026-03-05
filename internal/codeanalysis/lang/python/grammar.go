package python

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	tspy "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

func grammar() *sitter.Language {
	return sitter.NewLanguage(tspy.Language())
}
