package golang

import "testing"

func TestDetector_IsTestPath(t *testing.T) {
	t.Parallel()

	d := detector{}
	if !d.IsTestPath("x_test.go") {
		t.Fatalf("expected *_test.go to be test")
	}
	if d.IsTestPath("x.go") {
		t.Fatalf("did not expect x.go to be test")
	}
}
