package elixir

import (
	"path/filepath"
	"strings"
)

type detector struct{}

func (detector) Extensions() []string { return []string{".ex", ".exs"} }

func (detector) MatchesPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".ex" || ext == ".exs"
}

func (detector) IsTestPath(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), "_test.exs")
}
