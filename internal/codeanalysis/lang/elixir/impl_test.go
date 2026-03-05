package elixir

import "testing"

func TestNew_Basics(t *testing.T) {
	t.Parallel()

	lang := New()
	if lang.ID() != "elixir" {
		t.Fatalf("unexpected id: %q", lang.ID())
	}
	if !lang.Detector().MatchesPath("x.ex") {
		t.Fatalf("expected match ex")
	}
	if !lang.Detector().MatchesPath("x.exs") {
		t.Fatalf("expected match exs")
	}
}
