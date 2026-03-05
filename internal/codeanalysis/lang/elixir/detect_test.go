package elixir

import "testing"

func TestDetector_IsTestPath(t *testing.T) {
	t.Parallel()

	d := detector{}
	if !d.IsTestPath("x_test.exs") {
		t.Fatalf("expected *_test.exs to be test")
	}
	if d.IsTestPath("x.ex") {
		t.Fatalf("did not expect x.ex to be test")
	}
}
