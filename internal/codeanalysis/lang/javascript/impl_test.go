package javascript

import "testing"

func TestNew_Basics(t *testing.T) {
	t.Parallel()

	lang := New()
	if lang.ID() != "javascript" {
		t.Fatalf("unexpected id: %q", lang.ID())
	}
	exts := lang.Detector().Extensions()
	if len(exts) == 0 {
		t.Fatalf("expected extensions")
	}
	if !lang.Detector().MatchesPath("x.js") {
		t.Fatalf("expected match")
	}
}
