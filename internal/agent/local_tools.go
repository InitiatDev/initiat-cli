package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

type LocalToolRunner struct {
	baseDir  string
	executor setup.CommandExecutor
}

const (
	defaultCommandTimeout             = 30 * time.Minute
	defaultDirPerm        os.FileMode = 0o755
	defaultFilePerm       os.FileMode = 0o644
)

func NewLocalToolRunner(baseDir string) (*LocalToolRunner, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("baseDir is required")
	}
	abs, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("abs baseDir: %w", err)
	}
	return &LocalToolRunner{
		baseDir:  abs,
		executor: setup.NewRealCommandExecutor(),
	}, nil
}

func (t *LocalToolRunner) RunCommand(ctx context.Context, action ProposedAction) error {
	command := strings.TrimSpace(action.Command)
	if command == "" {
		return fmt.Errorf("command is required")
	}

	wd, err := t.resolvePathWithinBase(action.CWD, true)
	if err != nil {
		return err
	}

	req := &setup.CommandRequest{
		Command:    command,
		Env:        action.Env,
		WorkingDir: wd,
		Timeout:    defaultCommandTimeout,
	}

	// #nosec G204 -- command execution is explicitly user-approved via Approver before reaching ToolRunner.
	res, err := t.executor.Execute(ctx, req)
	if err != nil {
		exitCode := -1
		timedOut := false
		if res != nil {
			exitCode = res.ExitCode
			timedOut = res.TimedOut
		}
		return fmt.Errorf("command failed (exit=%d timed_out=%t): %w", exitCode, timedOut, err)
	}
	return nil
}

func (t *LocalToolRunner) EditFiles(ctx context.Context, action ProposedAction) error {
	_ = ctx
	if len(action.Edits) == 0 {
		return fmt.Errorf("edits are required")
	}

	for i := range action.Edits {
		edit := action.Edits[i]
		if strings.TrimSpace(edit.Path) == "" {
			return fmt.Errorf("edit[%d] path is required", i)
		}
		if t.isGitMetadataPath(edit.Path) {
			return fmt.Errorf("edit[%d] path %q is not allowed", i, edit.Path)
		}

		target, err := t.resolvePathWithinBase(edit.Path, false)
		if err != nil {
			return fmt.Errorf("edit[%d]: %w", i, err)
		}

		dir := filepath.Dir(target)
		if err := os.MkdirAll(dir, defaultDirPerm); err != nil {
			return fmt.Errorf("edit[%d] mkdir: %w", i, err)
		}

		perm := defaultFilePerm
		if existing, err := os.Stat(target); err == nil {
			perm = existing.Mode().Perm()
		}

		// #nosec G306 -- writes are explicitly user-approved and target workspace code, not secrets.
		if err := os.WriteFile(target, []byte(edit.Content), perm); err != nil {
			return fmt.Errorf("edit[%d] write: %w", i, err)
		}
	}

	return nil
}

func (t *LocalToolRunner) resolvePathWithinBase(path string, allowEmpty bool) (string, error) {
	if strings.TrimSpace(path) == "" {
		if allowEmpty {
			return t.baseDir, nil
		}
		return "", fmt.Errorf("path is required")
	}

	var abs string
	if filepath.IsAbs(path) {
		abs = filepath.Clean(path)
	} else {
		abs = filepath.Clean(filepath.Join(t.baseDir, path))
	}

	baseClean := filepath.Clean(t.baseDir)
	if abs == baseClean {
		return abs, nil
	}

	prefix := baseClean + string(os.PathSeparator)
	if !strings.HasPrefix(abs, prefix) {
		return "", fmt.Errorf("path %q escapes baseDir", path)
	}

	if err := disallowSymlinkPath(baseClean, abs); err != nil {
		return "", err
	}

	return abs, nil
}

func disallowSymlinkPath(baseDir, targetAbs string) error {
	rel, err := filepath.Rel(baseDir, targetAbs)
	if err != nil {
		return fmt.Errorf("rel: %w", err)
	}
	if rel == "." {
		return nil
	}
	if strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." {
		return fmt.Errorf("path escapes baseDir")
	}

	cur := baseDir
	parts := strings.Split(rel, string(os.PathSeparator))
	for _, p := range parts {
		if p == "" || p == "." {
			continue
		}
		cur = filepath.Join(cur, p)
		fi, err := os.Lstat(cur)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("lstat %s: %w", cur, err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path contains symlink: %s", cur)
		}
	}
	return nil
}

func (t *LocalToolRunner) isGitMetadataPath(p string) bool {
	clean := filepath.Clean(p)
	if clean == ".git" {
		return true
	}
	return strings.HasPrefix(clean, ".git"+string(os.PathSeparator)) ||
		strings.Contains(clean, string(os.PathSeparator)+".git"+string(os.PathSeparator))
}
