package typescript

import "testing"

func TestNew_Basics(t *testing.T) {
	t.Parallel()

	lang := New()
	if lang.ID() != "typescript" {
		t.Fatalf("unexpected id: %q", lang.ID())
	}
	if !lang.Detector().MatchesPath("x.ts") {
		t.Fatalf("expected match ts")
	}
	if !lang.Detector().MatchesPath("x.tsx") {
		t.Fatalf("expected match tsx")
	}
}
