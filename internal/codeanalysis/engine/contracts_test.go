package engine

import (
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/ast"
	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/treesitter"
)

type testDetector struct{}

func (testDetector) Extensions() []string         { return []string{".go"} }
func (testDetector) MatchesPath(path string) bool { return path == "main.go" }
func (testDetector) IsTestPath(path string) bool  { return false }

type testLanguage struct{}

func (testLanguage) ID() string          { return "go" }
func (testLanguage) DisplayName() string { return "Go" }
func (testLanguage) Detector() Detector  { return testDetector{} }
func (testLanguage) ParserFactory() ParserFactory {
	return ParserFactoryFunc(func() (*treesitter.ParserHandle, error) { return nil, nil })
}
func (testLanguage) Normalizer() Normalizer {
	return NormalizerFunc(func(_ treesitter.Node, _ []byte, _ string) (ast.Node, []ast.ParseError) { return ast.Node{}, nil })
}
func (testLanguage) QueryPack() QueryPack { return EmptyQueryPack{} }

func TestContracts_CompileAndBehave(t *testing.T) {
	t.Parallel()

	var _ Language = testLanguage{}
	var _ Detector = testDetector{}

	lang := testLanguage{}
	if lang.ID() != "go" {
		t.Fatalf("unexpected id: %q", lang.ID())
	}
	if !lang.Detector().MatchesPath("main.go") {
		t.Fatalf("expected detector match")
	}
}
