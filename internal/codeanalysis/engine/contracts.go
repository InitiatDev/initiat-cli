package engine

import (
	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/ast"
	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/treesitter"
)

type Language interface {
	ID() string
	DisplayName() string

	Detector() Detector
	ParserFactory() ParserFactory
	Normalizer() Normalizer
	QueryPack() QueryPack
}

type Detector interface {
	Extensions() []string
	MatchesPath(path string) bool
	IsTestPath(path string) bool
}

type ParserFactory interface {
	New() (*treesitter.ParserHandle, error)
}

type ParserFactoryFunc func() (*treesitter.ParserHandle, error)

func (f ParserFactoryFunc) New() (*treesitter.ParserHandle, error) { return f() }

type Normalizer interface {
	Normalize(root treesitter.Node, source []byte, path string) (ast.Node, []ast.ParseError)
}

type NormalizerFunc func(root treesitter.Node, source []byte, path string) (ast.Node, []ast.ParseError)

func (f NormalizerFunc) Normalize(root treesitter.Node, source []byte, path string) (ast.Node, []ast.ParseError) {
	return f(root, source, path)
}

type Query struct {
	Name string
	Src  string
}

type QueryPack interface {
	Interfaces() []Query
	Controllers() []Query
	IO() []Query
}

type EmptyQueryPack struct{}

func (EmptyQueryPack) Interfaces() []Query  { return nil }
func (EmptyQueryPack) Controllers() []Query { return nil }
func (EmptyQueryPack) IO() []Query          { return nil }
