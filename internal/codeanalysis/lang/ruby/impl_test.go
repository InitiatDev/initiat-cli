package ruby

import "testing"

func TestNew_Basics(t *testing.T) {
	t.Parallel()

	lang := New()
	if lang.ID() != "ruby" {
		t.Fatalf("unexpected id: %q", lang.ID())
	}
	if !lang.Detector().MatchesPath("x.rb") {
		t.Fatalf("expected match")
	}
}
