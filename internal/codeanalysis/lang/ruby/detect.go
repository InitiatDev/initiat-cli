package ruby

import (
	"path/filepath"
	"strings"
)

type detector struct{}

func (detector) Extensions() []string { return []string{".rb"} }

func (detector) MatchesPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".rb")
}

func (detector) IsTestPath(path string) bool {
	p := strings.ToLower(filepath.ToSlash(path))
	base := strings.ToLower(filepath.Base(p))

	if strings.HasPrefix(base, "test_") {
		return true
	}

	stem := strings.TrimSuffix(base, filepath.Ext(base))
	if strings.HasSuffix(stem, "_test") || strings.HasSuffix(stem, "_spec") {
		return true
	}

	if strings.Contains(p, "/test/") || strings.HasPrefix(p, "test/") {
		return true
	}
	if strings.Contains(p, "/tests/") || strings.HasPrefix(p, "tests/") {
		return true
	}
	if strings.Contains(p, "/spec/") || strings.HasPrefix(p, "spec/") {
		return true
	}
	return false
}
