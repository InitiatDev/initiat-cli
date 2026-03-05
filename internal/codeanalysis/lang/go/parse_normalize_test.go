package golang

import (
	"testing"
)

func TestLanguage_ParseAndNormalize(t *testing.T) {
	t.Parallel()

	lang := New()
	p, err := lang.ParserFactory().New()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	defer p.Close()

	tree, err := p.Parse([]byte("package main\nfunc main() {}\n"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	defer tree.Close()

	root := tree.Root()
	out, errs := lang.Normalizer().Normalize(root, []byte("package main\nfunc main() {}\n"), "main.go")
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %+v", errs)
	}
	if out.Type != "source_file" {
		t.Fatalf("unexpected root type: %q", out.Type)
	}
}
