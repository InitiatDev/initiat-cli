package golang

import (
	"testing"
)

func TestNew_Basics(t *testing.T) {
	t.Parallel()

	lang := New()
	if lang.ID() != "go" {
		t.Fatalf("unexpected id: %q", lang.ID())
	}
	if lang.DisplayName() == "" {
		t.Fatalf("expected display name")
	}
	exts := lang.Detector().Extensions()
	if len(exts) != 1 || exts[0] != ".go" {
		t.Fatalf("unexpected extensions: %v", exts)
	}
	if !lang.Detector().MatchesPath("main.go") {
		t.Fatalf("expected match")
	}
}
