package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/scaffold"
)

type ProjectSnapshot struct {
	BaseDir          string   `json:"base_dir"`
	RootEntries      []string `json:"root_entries"`
	ReadmeCandidates []string `json:"readme_candidates"`
	ImportantFiles   []string `json:"important_files"`
	DetectedKinds    []string `json:"detected_kinds,omitempty"`
}

const (
	maxSnapshotEntries     = 200
	importantFilesCapacity = 16
)

func BuildProjectSnapshot(baseDir string) (string, *ProjectSnapshot, error) {
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		return "", nil, fmt.Errorf("baseDir is required")
	}

	abs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", nil, fmt.Errorf("abs baseDir: %w", err)
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return "", nil, fmt.Errorf("readdir: %w", err)
	}

	root := make([]string, 0, min(len(entries), maxSnapshotEntries))
	for i, e := range entries {
		if i >= maxSnapshotEntries {
			break
		}
		suffix := ""
		if e.IsDir() {
			suffix = string(os.PathSeparator)
		}
		root = append(root, e.Name()+suffix)
	}
	sort.Strings(root)

	candidates := []string{"README.md", "README", "readme.md", "Readme.md"}
	var present []string
	for _, c := range candidates {
		for _, e := range root {
			if strings.TrimSuffix(e, string(os.PathSeparator)) == c {
				present = append(present, c)
				break
			}
		}
	}

	important := selectImportantFiles(root, present)

	kinds, _ := scaffold.Detect(abs)

	snap := &ProjectSnapshot{
		BaseDir:          abs,
		RootEntries:      root,
		ReadmeCandidates: present,
		ImportantFiles:   important,
		DetectedKinds:    kinds,
	}

	b, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("marshal snapshot: %w", err)
	}
	return string(b), snap, nil
}

func selectImportantFiles(rootEntries []string, readmes []string) []string {
	rootSet := make(map[string]struct{}, len(rootEntries))
	for _, e := range rootEntries {
		rootSet[strings.TrimSuffix(e, string(os.PathSeparator))] = struct{}{}
	}
	_, hasInitiatDir := rootSet[".initiat"]

	candidates := []string{
		// initiat files (only if .initiat/ exists)
		".initiat/setup.yml",
		".initiat/config.yml",
		// readme(s)
		"README.md", "README", "readme.md", "Readme.md",
		// package manager / build markers
		"go.mod", "go.sum",
		"package.json", "pnpm-lock.yaml", "yarn.lock", "package-lock.json", "bun.lockb",
		"pyproject.toml", "requirements.txt", "requirements-dev.txt",
		"mix.exs", "mix.lock",
		"Gemfile", "Gemfile.lock",
		"Cargo.toml", "Cargo.lock",
		"composer.json", "composer.lock",
		// container / infra
		"Dockerfile", "docker-compose.yml", "compose.yml",
		// common tooling/version managers
		"Makefile", ".tool-versions", ".node-version", ".python-version", ".ruby-version",
	}

	out := make([]string, 0, importantFilesCapacity)
	add := func(p string) {
		if strings.HasPrefix(p, ".initiat/") && !hasInitiatDir {
			return
		}
		if _, ok := rootSet[p]; !ok {
			return
		}
		for _, existing := range out {
			if existing == p {
				return
			}
		}
		out = append(out, p)
	}

	for _, r := range readmes {
		add(r)
	}
	for _, c := range candidates {
		add(c)
	}
	sort.Strings(out)
	return out
}
