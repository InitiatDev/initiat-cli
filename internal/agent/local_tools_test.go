package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalToolRunner_EditFiles_WritesAndPreventsEscape(t *testing.T) {
	dir := t.TempDir()
	runner, err := NewLocalToolRunner(dir)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if err := runner.EditFiles(context.Background(), ProposedAction{
		Type: ActionEditFiles,
		Edits: []FileEdit{
			{Path: "a/b/c.txt", Content: "hello"},
		},
	}); err != nil {
		t.Fatalf("edit files: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "a/b/c.txt"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("unexpected content: %q", string(got))
	}

	if err := runner.EditFiles(context.Background(), ProposedAction{
		Type: ActionEditFiles,
		Edits: []FileEdit{
			{Path: "../escape.txt", Content: "nope"},
		},
	}); err == nil {
		t.Fatalf("expected escape error")
	}
}

func TestLocalToolRunner_EditFiles_RejectsGitMetadata(t *testing.T) {
	dir := t.TempDir()
	runner, err := NewLocalToolRunner(dir)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if err := runner.EditFiles(context.Background(), ProposedAction{
		Type: ActionEditFiles,
		Edits: []FileEdit{
			{Path: ".git/config", Content: "x"},
		},
	}); err == nil {
		t.Fatalf("expected .git rejection")
	}
}

func TestLocalToolRunner_RunCommand_BadExit(t *testing.T) {
	dir := t.TempDir()
	runner, err := NewLocalToolRunner(dir)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	_, err = runner.RunCommand(context.Background(), ProposedAction{
		Type:    ActionRunCommand,
		Command: "sh -c 'exit 7'",
		CWD:     dir,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}
