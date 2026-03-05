package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/InitiatDev/initiat-cli/internal/codeanalysis"
	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/ast"
	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/engine"
)

var (
	analyzeASTLang      string
	analyzeASTFormat    string
	analyzeASTOutput    string
	analyzeASTRecursive bool
	analyzeASTFailOnErr bool
	analyzeASTMaxBytes  int64
)

const analyzeASTDefaultMaxBytes int64 = 2 << 20

func init() {
	analyzeCmd.Flags().StringVar(
		&analyzeASTLang,
		"lang",
		"auto",
		"Language (auto, go, javascript, typescript, python, ruby, elixir)",
	)
	analyzeCmd.Flags().StringVar(
		&analyzeASTFormat,
		"format",
		"jsonl",
		"Output format (json or jsonl)",
	)
	analyzeCmd.Flags().StringVarP(
		&analyzeASTOutput,
		"output",
		"o",
		"",
		"Write output to file instead of stdout",
	)
	analyzeCmd.Flags().BoolVar(
		&analyzeASTRecursive,
		"recursive",
		true,
		"Recurse into subdirectories when the path is a directory",
	)
	analyzeCmd.Flags().BoolVar(
		&analyzeASTFailOnErr,
		"fail-on-error",
		false,
		"Exit non-zero if any parse errors are found",
	)
	analyzeCmd.Flags().Int64Var(
		&analyzeASTMaxBytes,
		"max-bytes",
		analyzeASTDefaultMaxBytes,
		"Skip files larger than N bytes (0 disables)",
	)
}

func runAnalyzeAST(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) == 1 && strings.TrimSpace(args[0]) != "" {
		path = args[0]
	}

	format, err := parseASTFormat(analyzeASTFormat)
	if err != nil {
		return err
	}

	reg, err := codeanalysis.DefaultRegistry()
	if err != nil {
		return err
	}

	var w *os.File
	if strings.TrimSpace(analyzeASTOutput) == "" {
		w = os.Stdout
	} else {
		f, err := os.Create(analyzeASTOutput) // #nosec G304 -- user-provided output path is intentional
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	_, err = engine.ExportAST(context.Background(), reg, path, engine.ExportASTOptions{
		Lang:        strings.TrimSpace(analyzeASTLang),
		Recursive:   analyzeASTRecursive,
		Format:      format,
		Output:      analyzeASTOutput,
		MaxBytes:    analyzeASTMaxBytes,
		FailOnError: analyzeASTFailOnErr,
	}, w)
	return err
}

func parseASTFormat(s string) (ast.Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "json":
		return ast.FormatJSON, nil
	case "jsonl":
		return ast.FormatJSONL, nil
	default:
		return "", fmt.Errorf("invalid format %q (expected json or jsonl)", s)
	}
}
