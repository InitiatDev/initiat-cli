package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/InitiatDev/initiat-cli/internal/scaffold"
	"github.com/InitiatDev/initiat-cli/internal/setup"
)

const defaultSetupFile = ".initiat/setup.yml"

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Manage and validate setup scripts",
	Long:  `Validate setup scripts, generate .initiat/setup.yml from templates, and run setup.`,
}

var setupGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate .initiat/setup.yml from detected project",
	Long: `Detect language/framework from the current directory and write .initiat/setup.yml
and .initiat/config.yml. Use --force to overwrite existing setup.yml.`,
	RunE: runSetupGenerate,
}

var setupRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the setup script",
	Long: `Run .initiat/setup.yml (offline; no project context required).
Fails if the script requires secrets.`,
	RunE: runSetupRun,
}

var setupGenerateForce bool

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
	setupCmd.AddCommand(setupGenerateCmd)
	setupCmd.AddCommand(setupRunCmd)
	setupCmd.AddCommand(setupValidateCmd)
	setupCmd.AddCommand(setupSchemaCmd)

	setupGenerateCmd.Flags().BoolVarP(&setupGenerateForce, "force", "f", false, "Overwrite existing .initiat/setup.yml")
	setupSchemaCmd.Flags().StringVarP(&schemaOutput, "output", "o", "", "Save schema to file instead of stdout")
}

func runSetupGenerate(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()
	ids, err := scaffold.Detect(dir)
	if err != nil {
		return fmt.Errorf("detect: %w", err)
	}
	projectName := scaffold.ProjectName(dir)
	config, err := scaffold.Merge(dir, ids, projectName)
	if err != nil {
		return fmt.Errorf("merge: %w", err)
	}
	if err := setup.Validate(config); err != nil {
		return fmt.Errorf("generated config invalid: %w", err)
	}
	wroteSetup, wroteConfig, err := scaffold.Write(config, scaffold.WriteOptions{
		Dir:         dir,
		ProjectName: projectName,
		ForceSetup:  setupGenerateForce,
		ForceConfig: setupGenerateForce,
	})
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if !wroteSetup && !wroteConfig {
		fmt.Println(".initiat/setup.yml already exists. Use --force to overwrite.")
		return nil
	}
	if wroteSetup {
		fmt.Println("Wrote .initiat/setup.yml")
	}
	if wroteConfig {
		fmt.Println("Wrote .initiat/config.yml")
	}
	fmt.Println("Run: initiat setup validate && initiat setup run")
	return nil
}

func runSetupRun(cmd *cobra.Command, args []string) error {
	setupPath := defaultSetupFile
	if len(args) > 0 {
		setupPath = args[0]
	}
	config, err := setup.ParseFile(setupPath)
	if err != nil {
		return fmt.Errorf("parse %s: %w", setupPath, err)
	}
	if err := setup.Validate(config); err != nil {
		if validationErrs, ok := err.(setup.ValidationErrors); ok {
			for _, e := range validationErrs {
				fmt.Printf("  - %s\n", e.Error())
			}
		} else {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("validation failed")
	}
	runner := setup.NewSetupRunner(nil)
	if err := runner.Run(config); err != nil {
		if errors.Is(err, setup.ErrNoCommandsToExecute) {
			fmt.Println("No commands to execute (all steps skipped by conditions).")
			return nil
		}
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

func runSetupValidate(cmd *cobra.Command, args []string) error {
	setupFile := defaultSetupFile
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
