package python

import (
	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/ast"
	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/treesitter"
)

type normalizer struct{}

func (normalizer) Normalize(root treesitter.Node, _ []byte, _ string) (ast.Node, []ast.ParseError) {
	return treesitter.Normalize(root), nil
}
