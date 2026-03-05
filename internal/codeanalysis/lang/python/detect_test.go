package python

import "testing"

func TestDetector_IsTestPath(t *testing.T) {
	t.Parallel()

	d := detector{}
	for _, p := range []string{
		"test_a.py",
		"a_test.py",
		"tests/test_b.py",
	} {
		if !d.IsTestPath(p) {
			t.Fatalf("expected test path: %s", p)
		}
	}
	if d.IsTestPath("app/main.py") {
		t.Fatalf("did not expect app/main.py to be test")
	}
}
