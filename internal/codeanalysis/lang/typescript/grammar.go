package typescript

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
	tsts "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

func grammar() *sitter.Language {
	return sitter.NewLanguage(tsts.LanguageTSX())
}
