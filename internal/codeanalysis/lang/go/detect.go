package golang

import (
	"path/filepath"
	"strings"
)

type detector struct{}

func (detector) Extensions() []string { return []string{".go"} }

func (detector) MatchesPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".go")
}

func (detector) IsTestPath(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), "_test.go")
}
