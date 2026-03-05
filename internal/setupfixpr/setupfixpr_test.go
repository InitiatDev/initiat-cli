package setupfixpr

import "testing"

func TestParseGitPorcelainPaths(t *testing.T) {
	in := " M .initiat/setup.yml\nR  old.yml -> .initiat/setup.yaml\n?? foo.txt\n"
	paths := ParseGitPorcelainPaths(in)
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}
	if paths[0] != ".initiat/setup.yml" {
		t.Fatalf("unexpected first: %q", paths[0])
	}
	if paths[1] != ".initiat/setup.yaml" {
		t.Fatalf("unexpected rename target: %q", paths[1])
	}
	if paths[2] != "foo.txt" {
		t.Fatalf("unexpected third: %q", paths[2])
	}
}

func TestEligibleSetupFixPaths(t *testing.T) {
	in := []string{".initiat/setup.yml", ".initiat/agent-summary.md", "README.md", ".initiat/setup.yaml"}
	got := EligibleSetupFixPaths(in)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[0] != ".initiat/setup.yml" || got[1] != ".initiat/setup.yaml" {
		t.Fatalf("unexpected: %#v", got)
	}
}
