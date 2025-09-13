package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "initflow",
	Short: "InitFlow CLI",
	Long:  `InitFlow CLI — secure secrets, onboarding, and policy tooling.`,
}

// Execute is the entry point called by main.go
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
