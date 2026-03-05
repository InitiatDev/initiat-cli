package engine

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/ast"
)

type ExportASTOptions struct {
	Lang      string
	Recursive bool
	Format    ast.Format
	Output    string
	MaxBytes  int64

	FailOnError bool
}

type ExportASTSummary struct {
	FilesConsidered  int
	FilesParsed      int
	DocumentsEmitted int
	ParseErrors      int
}

func ExportAST(
	ctx context.Context,
	reg *Registry,
	path string,
	opts ExportASTOptions,
	w io.Writer,
) (ExportASTSummary, error) {
	if reg == nil {
		return ExportASTSummary{}, fmt.Errorf("registry cannot be nil")
	}
	if w == nil {
		return ExportASTSummary{}, fmt.Errorf("writer cannot be nil")
	}

	files, err := CollectFiles(path, WalkOptions{
		Recursive:        opts.Recursive,
		RespectGitIgnore: true,
		SkipTestDirs:     true,
	})
	if err != nil {
		return ExportASTSummary{}, err
	}

	var docs []ast.Document
	summary := ExportASTSummary{FilesConsidered: len(files)}

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return summary, err
		}

		doc, parseErrs, ok, err := exportASTFile(reg, file, opts)
		if err != nil {
			return summary, err
		}
		if !ok {
			continue
		}

		docs = append(docs, doc)
		summary.FilesParsed++
		summary.ParseErrors += parseErrs
	}

	if err := ast.Encode(w, opts.Format, docs); err != nil {
		return summary, err
	}
	summary.DocumentsEmitted = len(docs)

	if opts.FailOnError && summary.ParseErrors > 0 {
		return summary, fmt.Errorf("parse errors: %d", summary.ParseErrors)
	}

	return summary, nil
}

func exportASTFile(reg *Registry, file string, opts ExportASTOptions) (ast.Document, int, bool, error) {
	info, err := os.Stat(file)
	if err != nil {
		return ast.Document{}, 0, false, err
	}
	if info.IsDir() {
		return ast.Document{}, 0, false, nil
	}
	if opts.MaxBytes > 0 && info.Size() > opts.MaxBytes {
		return ast.Document{}, 0, false, nil
	}

	lang, ok := pickLanguage(reg, file, opts.Lang)
	if !ok {
		return ast.Document{}, 0, false, nil
	}
	if lang.Detector().IsTestPath(file) {
		return ast.Document{}, 0, false, nil
	}

	src, err := os.ReadFile(file) // #nosec G304 -- file paths come from CollectFiles under the requested root
	if err != nil {
		return ast.Document{}, 0, false, err
	}

	parser, err := lang.ParserFactory().New()
	if err != nil {
		return ast.Document{}, 0, false, fmt.Errorf("%s: new parser: %w", file, err)
	}
	defer parser.Close()

	tree, err := parser.Parse(src)
	if err != nil {
		return ast.Document{}, 0, false, fmt.Errorf("%s: parse: %w", file, err)
	}
	defer tree.Close()

	root := tree.Root()
	node, normErrs := lang.Normalizer().Normalize(root, src, file)

	doc := ast.Document{
		Version:  "ast-v1",
		Language: lang.ID(),
		Path:     file,
		Root:     node,
		Errors:   normErrs,
	}
	return doc, len(normErrs), true, nil
}

func pickLanguage(reg *Registry, path string, forced string) (Language, bool) {
	if forced == "" || forced == "auto" {
		return reg.DetectByPath(path)
	}
	lang, ok := reg.Get(forced)
	return lang, ok
}
