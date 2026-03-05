package cmd

import "github.com/spf13/cobra"

var analyzeCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "Analyze source code",
	Long:  "Analyze source code using Tree-sitter and output generic results for later exploration.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runAnalyzeAST,
}

func init() {
	codeCmd.AddCommand(analyzeCmd)
}
