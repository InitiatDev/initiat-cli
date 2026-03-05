package javascript

import "testing"

func TestDetector_IsTestPath(t *testing.T) {
	t.Parallel()

	d := detector{}
	for _, p := range []string{
		"a.test.js",
		"b.spec.js",
		"__tests__/c.js",
		"test/d.js",
		"tests/e.js",
	} {
		if !d.IsTestPath(p) {
			t.Fatalf("expected test path: %s", p)
		}
	}
	if d.IsTestPath("src/app.js") {
		t.Fatalf("did not expect src/app.js to be test")
	}
}
