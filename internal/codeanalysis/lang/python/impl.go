package python

import "github.com/InitiatDev/initiat-cli/internal/codeanalysis/engine"

func New() engine.Language {
	return language{}
}

type language struct{}

func (language) ID() string          { return "python" }
func (language) DisplayName() string { return "Python" }
func (language) Detector() engine.Detector {
	return detector{}
}
func (language) ParserFactory() engine.ParserFactory {
	return parserFactory{}
}
func (language) Normalizer() engine.Normalizer {
	return normalizer{}
}
func (language) QueryPack() engine.QueryPack { return engine.EmptyQueryPack{} }
