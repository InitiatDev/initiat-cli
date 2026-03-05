package codeanalysis

import "testing"

func TestDefaultRegistry_RegistersLanguages(t *testing.T) {
	t.Parallel()

	r, err := DefaultRegistry()
	if err != nil {
		t.Fatalf("default registry: %v", err)
	}

	for _, id := range []string{"go", "javascript", "typescript", "python", "ruby", "elixir"} {
		if _, ok := r.Get(id); !ok {
			t.Fatalf("expected language registered: %s", id)
		}
	}
}
