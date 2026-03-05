package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/InitiatDev/initiat-cli/internal/agent"
	initiatconfig "github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/prompt"
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
	setupConfig, err := setup.ParseFile(setupPath)
	if err != nil {
		return fmt.Errorf("parse %s: %w", setupPath, err)
	}
	if err := setup.Validate(setupConfig); err != nil {
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
	if err := runner.Run(setupConfig); err != nil {
		if errors.Is(err, setup.ErrNoCommandsToExecute) {
			fmt.Println("No commands to execute (all steps skipped by conditions).")
			return nil
		}

		handled, handledErr := maybeRunAgentForSetupFailure(context.Background(), runner, setupConfig, err)
		if handledErr != nil {
			return handledErr
		}
		if handled {
			return nil
		}

		return fmt.Errorf("run: %w", err)
	}
	return nil
}

func maybeRunAgentForSetupFailure(
	ctx context.Context,
	runner *setup.SetupRunner,
	setupConfig *setup.SetupConfig,
	runErr error,
) (bool, error) {
	var execErr *setup.SetupExecutionError
	if !errors.As(runErr, &execErr) || execErr == nil || execErr.Report == nil {
		return false, nil
	}

	cfg := initiatconfig.Get()
	if !cfg.Agent.Enabled {
		return false, nil
	}

	ok, err := prompt.PromptYesNo("Setup failed. Enter agent mode?")
	if err != nil {
		return true, err
	}
	if !ok {
		return false, nil
	}

	return true, runAgentOnce(ctx, runner, setupConfig, execErr)
}

func runAgentOnce(
	ctx context.Context,
	runner *setup.SetupRunner,
	setupConfig *setup.SetupConfig,
	execErr *setup.SetupExecutionError,
) error {
	runtimeCfg, err := agent.LoadRuntimeConfig()
	if err != nil {
		return err
	}

	llm, err := newAgentLLM(runtimeCfg)
	if err != nil {
		return err
	}

	wd := chooseAgentBaseDir(execErr)
	tools, err := agent.NewLocalToolRunner(wd)
	if err != nil {
		return err
	}

	orchestrator, err := agent.NewOrchestrator(llm, runtimeCfg.Model, &agent.PromptApprover{}, tools)
	if err != nil {
		return err
	}

	decision, err := orchestrator.Diagnose(ctx, execErr.Report)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Agent diagnosis:")
	fmt.Println(decision.Explanation)
	fmt.Println()

	if err := orchestrator.Apply(ctx, decision); err != nil {
		return err
	}

	retryErr := retrySetupIfRequested(runner, setupConfig, decision)

	_ = writeAgentSummary(ctx, llm, runtimeCfg.Model, wd, execErr, decision, retryErr)
	return retryErr
}

func newAgentLLM(cfg *agent.RuntimeConfig) (agent.LLM, error) {
	switch cfg.Provider {
	case agent.ProviderOpenAI:
		return agent.NewOpenAIClient(cfg.APIKey), nil
	case agent.ProviderAnthropic:
		return agent.NewAnthropicClient(cfg.APIKey), nil
	default:
		return nil, fmt.Errorf("unsupported agent provider: %s", cfg.Provider)
	}
}

func retrySetupIfRequested(runner *setup.SetupRunner, setupConfig *setup.SetupConfig, decision *agent.Decision) error {
	shouldRetry := false
	for _, a := range decision.Actions {
		if a.Type == agent.ActionRunCommand || a.Type == agent.ActionEditFiles {
			shouldRetry = true
			break
		}
	}
	if !shouldRetry {
		fmt.Println("Agent suggested no executable actions. Re-run setup after addressing prompts.")
		return fmt.Errorf("setup failed; agent provided guidance")
	}

	ok, err := prompt.PromptYesNo("Re-run setup now?")
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Re-run: initiat setup run")
		return fmt.Errorf("setup failed; agent made changes")
	}

	fmt.Println()
	fmt.Println("Re-running setup...")
	if err := runner.Run(setupConfig); err != nil {
		return fmt.Errorf("run after agent: %w", err)
	}

	return nil
}

func chooseAgentBaseDir(execErr *setup.SetupExecutionError) string {
	if execErr != nil && execErr.FailedCommand.WorkingDir != "" {
		return execErr.FailedCommand.WorkingDir
	}
	if execErr != nil && execErr.Report != nil {
		for _, c := range execErr.Report.Commands {
			if c.WorkingDir != "" {
				return c.WorkingDir
			}
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

const (
	agentSummaryDirPerm  os.FileMode = 0o755
	agentSummaryFilePerm os.FileMode = 0o600
)

func writeAgentSummary(
	ctx context.Context,
	llm agent.LLM,
	model string,
	baseDir string,
	execErr *setup.SetupExecutionError,
	decision *agent.Decision,
	retryErr error,
) error {
	if strings.TrimSpace(baseDir) == "" {
		return nil
	}
	if execErr == nil || execErr.Report == nil || decision == nil {
		return nil
	}

	buckets, err := agent.AssessIssues(ctx, llm, model, execErr.Report, decision)
	if err != nil {
		return err
	}

	summaryPath := filepath.Join(baseDir, ".initiat", "agent-summary.md")
	if err := os.MkdirAll(filepath.Dir(summaryPath), agentSummaryDirPerm); err != nil {
		return err
	}

	content := buildAgentSummaryMarkdown(execErr, decision, buckets, retryErr)
	if err := os.WriteFile(summaryPath, []byte(content), agentSummaryFilePerm); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Wrote agent summary to", summaryPath)
	return nil
}

func buildAgentSummaryMarkdown(
	execErr *setup.SetupExecutionError,
	decision *agent.Decision,
	buckets *agent.IssueBuckets,
	retryErr error,
) string {
	outcome := "setup re-run failed"
	if retryErr == nil {
		outcome = "setup re-run succeeded"
	}

	var b strings.Builder
	b.WriteString("# Agent-guided setup summary\n\n")
	b.WriteString("- Generated at: ")
	b.WriteString(time.Now().Format(time.RFC3339))
	b.WriteString("\n- Outcome: ")
	b.WriteString(outcome)
	b.WriteString("\n\n")

	b.WriteString("## Failure context\n\n")
	b.WriteString("- Phase: " + execErr.FailedCommand.Phase + "\n")
	b.WriteString("- Step: " + execErr.FailedCommand.StepName + "\n")
	b.WriteString("- Command: " + execErr.FailedCommand.Command + "\n\n")

	b.WriteString("## Issue buckets\n\n")
	writeBucketList(&b, "### Setup script / app issues", buckets.SetupOrApp)
	writeBucketList(&b, "### Local environment quirks", buckets.LocalEnvironment)

	if strings.TrimSpace(buckets.Notes) != "" {
		b.WriteString("\n## Notes\n\n")
		b.WriteString(buckets.Notes)
		b.WriteString("\n")
	}

	b.WriteString("\n## Agent proposed actions\n\n")
	for _, a := range decision.Actions {
		b.WriteString("- " + string(a.Type))
		if strings.TrimSpace(a.Reason) != "" {
			b.WriteString(": " + a.Reason)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func writeBucketList(b *strings.Builder, title string, items []string) {
	b.WriteString(title)
	b.WriteString("\n\n")
	if len(items) == 0 {
		b.WriteString("- (none identified)\n\n")
		return
	}
	for _, it := range items {
		b.WriteString("- " + it + "\n")
	}
	b.WriteString("\n")
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
