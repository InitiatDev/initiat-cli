package python

import "testing"

func TestNew_Basics(t *testing.T) {
	t.Parallel()

	lang := New()
	if lang.ID() != "python" {
		t.Fatalf("unexpected id: %q", lang.ID())
	}
	if !lang.Detector().MatchesPath("x.py") {
		t.Fatalf("expected match")
	}
}
