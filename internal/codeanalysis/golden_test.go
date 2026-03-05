package codeanalysis

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/ast"
	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/engine"
	elixirlang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/elixir"
	golanglang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/go"
	jslang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/javascript"
	pylang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/python"
	rubylang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/ruby"
	tslang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/typescript"
)

var updateGolden = flag.Bool("update-golden", false, "update AST golden files")

func TestGolden_ASTv1(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fixtureRel string
		goldenRel  string
		newLang    func() engine.Language
	}{
		{
			name:       "go",
			fixtureRel: "testdata/fixtures/sample.go",
			goldenRel:  "testdata/golden/go.json",
			newLang:    golanglang.New,
		},
		{
			name:       "javascript",
			fixtureRel: "testdata/fixtures/sample.js",
			goldenRel:  "testdata/golden/javascript.json",
			newLang:    jslang.New,
		},
		{
			name:       "typescript",
			fixtureRel: "testdata/fixtures/sample.ts",
			goldenRel:  "testdata/golden/typescript.json",
			newLang:    tslang.New,
		},
		{
			name:       "python",
			fixtureRel: "testdata/fixtures/sample.py",
			goldenRel:  "testdata/golden/python.json",
			newLang:    pylang.New,
		},
		{
			name:       "ruby",
			fixtureRel: "testdata/fixtures/sample.rb",
			goldenRel:  "testdata/golden/ruby.json",
			newLang:    rubylang.New,
		},
		{
			name:       "elixir",
			fixtureRel: "testdata/fixtures/sample.ex",
			goldenRel:  "testdata/golden/elixir.json",
			newLang:    elixirlang.New,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src, err := os.ReadFile(filepath.FromSlash(tc.fixtureRel))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			lang := tc.newLang()
			p, err := lang.ParserFactory().New()
			if err != nil {
				t.Fatalf("new parser: %v", err)
			}
			defer p.Close()

			tree, err := p.Parse(src)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			defer tree.Close()

			root := tree.Root()
			out, errs := lang.Normalizer().Normalize(root, src, "fixture")
			if len(errs) != 0 {
				t.Fatalf("unexpected errors: %+v", errs)
			}

			doc := ast.Document{
				Version:  "ast-v1",
				Language: lang.ID(),
				Path:     tc.fixtureRel,
				Root:     out,
			}

			got, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			got = append(got, '\n')

			goldenPath := filepath.FromSlash(tc.goldenRel)
			if *updateGolden {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o700); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(goldenPath, got, 0o600); err != nil {
					t.Fatalf("write golden: %v", err)
				}
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden: %v (run with -update-golden)", err)
			}
			if string(want) != string(got) {
				t.Fatalf("golden mismatch for %s\n--- want\n%s\n--- got\n%s", tc.name, string(want), string(got))
			}
		})
	}
}
