package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCollectFiles_SkipsDotGit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a"), 0o600); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o700); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "ignored.go"), []byte("package ignored"), 0o600); err != nil {
		t.Fatalf("write ignored.go: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "b.py"), []byte("print(1)"), 0o600); err != nil {
		t.Fatalf("write b.py: %v", err)
	}

	files, err := CollectFiles(dir, WalkOptions{Recursive: true})
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	wantA := filepath.Join(dir, "a.go")
	wantB := filepath.Join(dir, "sub", "b.py")

	hasA := false
	hasB := false
	for _, f := range files {
		if samePath(f, wantA) {
			hasA = true
		}
		if samePath(f, wantB) {
			hasB = true
		}
		if filepath.Base(filepath.Dir(f)) == ".git" {
			t.Fatalf("should not include .git file: %s", f)
		}
	}
	if !hasA || !hasB {
		t.Fatalf("missing expected files. hasA=%v hasB=%v files=%v", hasA, hasB, files)
	}
}

func TestCollectFiles_RespectsGitIgnore(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o700); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	gitignore := "ignored.py\nsub/*.txt\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0o600); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "keep.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("write keep.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored.py"), []byte("print(1)\n"), 0o600); err != nil {
		t.Fatalf("write ignored.py: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o700); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "keep.rb"), []byte("puts 1\n"), 0o600); err != nil {
		t.Fatalf("write keep.rb: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "ignored.txt"), []byte("nope\n"), 0o600); err != nil {
		t.Fatalf("write ignored.txt: %v", err)
	}

	old := gitCheckIgnoredPaths
	called := false
	gitCheckIgnoredPaths = func(_ string, stdin []byte) (map[string]struct{}, error) {
		called = true
		ignored := make(map[string]struct{})
		for _, part := range bytes.Split(stdin, []byte{0}) {
			if len(part) == 0 {
				continue
			}
			s := string(part)
			if strings.HasSuffix(s, "ignored.py") || strings.HasSuffix(s, "sub/ignored.txt") {
				ignored[s] = struct{}{}
			}
		}
		return ignored, nil
	}
	t.Cleanup(func() { gitCheckIgnoredPaths = old })

	files, err := CollectFiles(dir, WalkOptions{Recursive: true, RespectGitIgnore: true})
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if !called {
		t.Fatalf("expected git ignore checker to be invoked")
	}

	mustHave(t, files, filepath.Join(dir, "keep.go"))
	mustHave(t, files, filepath.Join(dir, "sub", "keep.rb"))
	mustNotHave(t, files, filepath.Join(dir, "ignored.py"))
	mustNotHave(t, files, filepath.Join(dir, "sub", "ignored.txt"))
}

func TestCollectFiles_SkipsTestDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("write keep.go: %v", err)
	}

	for _, d := range []string{"test", "tests", "spec", "__tests__"} {
		if err := os.Mkdir(filepath.Join(dir, d), 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
		if err := os.WriteFile(filepath.Join(dir, d, "ignored.go"), []byte("package ignored\n"), 0o600); err != nil {
			t.Fatalf("write %s/ignored.go: %v", d, err)
		}
	}

	files, err := CollectFiles(dir, WalkOptions{Recursive: true, SkipTestDirs: true})
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	mustHave(t, files, filepath.Join(dir, "keep.go"))
	for _, d := range []string{"test", "tests", "spec", "__tests__"} {
		mustNotHave(t, files, filepath.Join(dir, d, "ignored.go"))
	}
}

func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return a == b
}

func mustHave(t *testing.T, files []string, want string) {
	t.Helper()
	for _, f := range files {
		if samePath(f, want) {
			return
		}
	}
	t.Fatalf("expected to contain %s\nfiles=%v", want, files)
}

func mustNotHave(t *testing.T, files []string, bad string) {
	t.Helper()
	for _, f := range files {
		if samePath(f, bad) {
			t.Fatalf("expected NOT to contain %s\nfiles=%v", bad, files)
		}
	}
}
