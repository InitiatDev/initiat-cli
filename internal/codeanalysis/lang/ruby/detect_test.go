package ruby

import "testing"

func TestDetector_IsTestPath(t *testing.T) {
	t.Parallel()

	d := detector{}
	for _, p := range []string{
		"test_a.rb",
		"a_test.rb",
		"a_spec.rb",
		"spec/b_spec.rb",
		"test/c_test.rb",
	} {
		if !d.IsTestPath(p) {
			t.Fatalf("expected test path: %s", p)
		}
	}
	if d.IsTestPath("lib/app.rb") {
		t.Fatalf("did not expect lib/app.rb to be test")
	}
}
