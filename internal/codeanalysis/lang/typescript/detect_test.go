package typescript

import "testing"

func TestDetector_IsTestPath(t *testing.T) {
	t.Parallel()

	d := detector{}
	for _, p := range []string{
		"a.test.ts",
		"b.spec.ts",
		"a.test.tsx",
		"b.spec.tsx",
		"__tests__/c.ts",
		"test/d.ts",
		"tests/e.tsx",
	} {
		if !d.IsTestPath(p) {
			t.Fatalf("expected test path: %s", p)
		}
	}
	if d.IsTestPath("src/app.ts") {
		t.Fatalf("did not expect src/app.ts to be test")
	}
}
