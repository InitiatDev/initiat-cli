package javascript

import (
	"path/filepath"
	"strings"
)

type detector struct{}

func (detector) Extensions() []string { return []string{".js", ".mjs", ".cjs"} }

func (detector) MatchesPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}

func (detector) IsTestPath(path string) bool {
	p := strings.ToLower(filepath.ToSlash(path))
	base := strings.ToLower(filepath.Base(p))

	if strings.HasPrefix(base, "test_") {
		return true
	}
	if strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") {
		return true
	}
	if strings.HasSuffix(strings.TrimSuffix(base, filepath.Ext(base)), "_test") {
		return true
	}
	if strings.HasSuffix(strings.TrimSuffix(base, filepath.Ext(base)), "_spec") {
		return true
	}

	if strings.Contains(p, "/__tests__/") || strings.HasPrefix(p, "__tests__/") {
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
