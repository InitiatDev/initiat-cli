package ruby

import "github.com/InitiatDev/initiat-cli/internal/codeanalysis/treesitter"

type parserFactory struct{}

func (parserFactory) New() (*treesitter.ParserHandle, error) {
	return treesitter.NewParserHandle(grammar())
}
