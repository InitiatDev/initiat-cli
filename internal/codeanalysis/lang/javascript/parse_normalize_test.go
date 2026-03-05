package javascript

import "testing"

func TestLanguage_ParseAndNormalize(t *testing.T) {
	t.Parallel()

	lang := New()
	p, err := lang.ParserFactory().New()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	defer p.Close()

	src := []byte("const x = 1;\n")
	tree, err := p.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	defer tree.Close()

	out, errs := lang.Normalizer().Normalize(tree.Root(), src, "x.js")
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %+v", errs)
	}
	if out.Type == "" {
		t.Fatalf("expected root type")
	}
	if len(out.Children) == 0 {
		t.Fatalf("expected children")
	}
}
