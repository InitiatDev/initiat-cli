package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	internalgit "github.com/InitiatDev/initiat-cli/internal/git"
)

type WalkOptions struct {
	Recursive        bool
	RespectGitIgnore bool
	SkipTestDirs     bool
}

func CollectFiles(path string, opts WalkOptions) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		files := []string{path}
		if opts.RespectGitIgnore {
			gitRoot, ok := findGitRootForPath(path)
			if ok {
				return filterGitIgnored(gitRoot, files)
			}
		}
		return files, nil
	}

	gitRoot := ""
	if opts.RespectGitIgnore {
		if root, ok := findGitRootForPath(path); ok {
			gitRoot = root
		}
	}

	var out []string

	walkFn := func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if shouldSkipDir(d.Name(), opts.SkipTestDirs) {
				return fs.SkipDir
			}
			if p != path && !opts.Recursive {
				return fs.SkipDir
			}
			return nil
		}

		out = append(out, p)
		return nil
	}

	if err := filepath.WalkDir(path, walkFn); err != nil {
		return nil, fmt.Errorf("walk %s: %w", path, err)
	}

	if opts.RespectGitIgnore && gitRoot != "" {
		return filterGitIgnored(gitRoot, out)
	}

	return out, nil
}

func shouldSkipDir(name string, skipTests bool) bool {
	switch name {
	case ".git":
		return true
	case "node_modules", "vendor", "deps", "_build":
		return true
	case "dist", "build", "coverage":
		return true
	case ".venv", "venv", ".tox":
		return true
	case ".bundle":
		return true
	case ".idea", ".vscode":
		return true
	case "tmp", "log":
		return true
	default:
		if skipTests {
			switch name {
			case "test", "tests", "spec", "__tests__":
				return true
			}
		}
		return false
	}
}

func findGitRootForPath(path string) (string, bool) {
	h := internalgit.NewHandler()
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, "_")
	}
	return h.FindGitRoot(path)
}

func filterGitIgnored(gitRoot string, paths []string) ([]string, error) {
	if gitRoot == "" || len(paths) == 0 {
		return paths, nil
	}

	var stdin bytes.Buffer
	relToAbs := make(map[string]string, len(paths))
	var rels []string

	for _, abs := range paths {
		rel, err := filepath.Rel(gitRoot, abs)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		rels = append(rels, rel)
		relToAbs[rel] = abs
		stdin.WriteString(rel)
		stdin.WriteByte(0)
	}

	if len(rels) == 0 {
		return paths, nil
	}

	ignored, err := gitCheckIgnoredPaths(gitRoot, stdin.Bytes())
	if err != nil {
		return nil, err
	}

	kept := make([]string, 0, len(paths))
	for _, rel := range rels {
		if _, ok := ignored[rel]; ok {
			continue
		}
		kept = append(kept, relToAbs[rel])
	}
	return kept, nil
}

var gitCheckIgnoredPaths = gitCheckIgnoredPathsViaGit

func gitCheckIgnoredPathsViaGit(gitRoot string, stdinNulSeparated []byte) (map[string]struct{}, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return map[string]struct{}{}, nil
	}

	cmd := exec.Command("git", "-C", gitRoot, "check-ignore", "-z", "--stdin")
	cmd.Stdin = bytes.NewReader(stdinNulSeparated)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				return map[string]struct{}{}, nil
			}
		}
		return nil, fmt.Errorf("git check-ignore: %w", err)
	}

	ignored := make(map[string]struct{})
	for _, part := range bytes.Split(out, []byte{0}) {
		if len(part) == 0 {
			continue
		}
		ignored[string(part)] = struct{}{}
	}
	return ignored, nil
}
