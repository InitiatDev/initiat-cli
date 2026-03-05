package setupfixpr

import (
	"path/filepath"
	"strings"
)

const (
	porcelainMinLen   = 4
	renameArrow       = " -> "
	defaultBaseBranch = "main"
)

func EligibleSetupFixPaths(paths []string) []string {
	var out []string
	for _, p := range paths {
		p = filepath.Clean(p)
		switch p {
		case filepath.Join(".initiat", "setup.yml"),
			filepath.Join(".initiat", "setup.yaml"):
			out = append(out, p)
		}
	}
	return out
}

func ParseGitPorcelainPaths(status string) []string {
	var out []string
	for _, line := range strings.Split(status, "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < porcelainMinLen {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if path == "" {
			continue
		}
		if strings.Contains(path, renameArrow) {
			parts := strings.Split(path, renameArrow)
			path = strings.TrimSpace(parts[len(parts)-1])
		}
		out = append(out, path)
	}
	return out
}

func BaseBranchFromOriginHeadRef(originHeadRef string) string {
	s := strings.TrimSpace(originHeadRef)
	const prefix = "refs/remotes/origin/"
	if strings.HasPrefix(s, prefix) {
		return strings.TrimPrefix(s, prefix)
	}
	return defaultBaseBranch
}
