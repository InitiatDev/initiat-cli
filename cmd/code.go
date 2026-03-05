package cmd

import "github.com/spf13/cobra"

var codeCmd = &cobra.Command{
	Use:   "code",
	Short: "Code-related tooling",
	Long:  "Code-related tooling such as static analysis outputs for later exploration.",
}

func init() {
	rootCmd.AddCommand(codeCmd)
}
