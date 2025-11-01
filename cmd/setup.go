package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Manage and validate setup scripts",
	Long:  `Validate setup scripts and generate JSON schemas for .initiat/setup.yml files.`,
}

var setupValidateCmd = &cobra.Command{
	Use:   "validate [setup-file]",
	Short: "Validate a setup script",
	Long: `Validate a setup script YAML file against the schema.

If no file is provided, defaults to .initiat/setup.yml.

Examples:
  initiat setup validate
  initiat setup validate .initiat/setup.yml
  initiat setup validate custom-setup.yml`,
	RunE: runSetupValidate,
}

var setupSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Output JSON Schema for setup scripts",
	Long: `Output the JSON Schema for .initiat/setup.yml files.

Examples:
  initiat setup schema
  initiat setup schema --output schemas/setup-v1.json`,
	RunE: runSetupSchema,
}

var schemaOutput string

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.AddCommand(setupValidateCmd)
	setupCmd.AddCommand(setupSchemaCmd)

	setupSchemaCmd.Flags().StringVarP(&schemaOutput, "output", "o", "", "Save schema to file instead of stdout")
}

func runSetupValidate(cmd *cobra.Command, args []string) error {
	setupFile := ".initiat/setup.yml"
	if len(args) > 0 {
		setupFile = args[0]
	}

	fmt.Printf("Validating %s...\n", setupFile)

	config, err := setup.ParseFile(setupFile)
	if err != nil {
		return fmt.Errorf("❌ Failed to parse setup file: %w", err)
	}

	if err := setup.Validate(config); err != nil {
		fmt.Println("❌ Validation failed:")
		if validationErrs, ok := err.(setup.ValidationErrors); ok {
			for _, validationErr := range validationErrs {
				fmt.Printf("  - %s\n", validationErr.Error())
			}
		} else {
			fmt.Printf("  - %s\n", err.Error())
		}
		return fmt.Errorf("validation failed")
	}

	fmt.Println("✅ Setup script is valid!")
	return nil
}

func runSetupSchema(cmd *cobra.Command, args []string) error {
	if schemaOutput != "" {
		if err := setup.SaveSchemaToFile(schemaOutput); err != nil {
			return fmt.Errorf("❌ Failed to save schema: %w", err)
		}
		fmt.Printf("✅ Schema saved to %s\n", schemaOutput)
		return nil
	}

	schema, err := setup.GenerateJSONSchema()
	if err != nil {
		return fmt.Errorf("❌ Failed to generate schema: %w", err)
	}

	fmt.Fprint(os.Stdout, string(schema))
	return nil
}
