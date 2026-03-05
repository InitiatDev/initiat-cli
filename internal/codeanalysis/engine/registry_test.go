package engine

import "testing"

func TestRegistry_RegisterAndLookup(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	if err := r.Register(testLanguage{}); err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, ok := r.Get("go"); !ok {
		t.Fatalf("expected get ok")
	}

	lang, ok := r.DetectByPath("main.go")
	if !ok {
		t.Fatalf("expected detect ok")
	}
	if lang.ID() != "go" {
		t.Fatalf("unexpected detected id: %q", lang.ID())
	}
}

func TestRegistry_DuplicateID(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	if err := r.Register(testLanguage{}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := r.Register(testLanguage{}); err == nil {
		t.Fatalf("expected duplicate error")
	}
}
