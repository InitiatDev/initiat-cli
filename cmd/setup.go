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
	Long: `Run .initiat/setup.yml. Uses project context when available (-p/-P, or .initiat/config.yml)
so steps that need secrets can run; otherwise runs offline. Agent recovery may run on failure if enabled.`,
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
	projectCtx, _ := initiatconfig.ResolveProjectContext(projectPath, org, projectName)
	runner := setup.NewSetupRunner(projectCtx)
	if err := runner.Run(setupConfig); err != nil {
		if errors.Is(err, setup.ErrNoCommandsToExecute) {
			fmt.Println("No commands to execute (all steps skipped by conditions).")
			return nil
		}

		handled, handledErr := maybeRunAgentForSetupFailure(context.Background(), runner, setupPath, setupConfig, err)
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
	setupPath string,
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

	return true, runAgentIterative(ctx, runner, setupPath, setupConfig, execErr)
}

func runAgentIterative(
	ctx context.Context,
	runner *setup.SetupRunner,
	setupPath string,
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
	orchestrator.SetDebugWriter(os.Stdout)

	var lastDecision *agent.Decision
	var lastApply *agent.ApplyResult

	for round := 1; round <= 10; round++ {
		next, done, err := runAgentRound(
			ctx,
			runner,
			orchestrator,
			llm,
			runtimeCfg,
			wd,
			setupPath,
			setupConfig,
			execErr,
			lastApply,
			round,
		)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		lastDecision = next.Decision
		lastApply = next.Apply
		setupConfig = next.SetupConfig
		_ = next.RetryErr

		if next.NextExecErr == nil {
			continue
		}
		execErr = next.NextExecErr
	}

	_ = lastDecision
	return fmt.Errorf("setup still failing after multiple agent rounds")
}

type agentRoundResult struct {
	Decision    *agent.Decision
	Apply       *agent.ApplyResult
	SetupConfig *setup.SetupConfig
	RetryErr    error
	NextExecErr *setup.SetupExecutionError
}

func runAgentRound(
	ctx context.Context,
	runner *setup.SetupRunner,
	orchestrator *agent.Orchestrator,
	llm agent.LLM,
	runtimeCfg *agent.RuntimeConfig,
	wd string,
	setupPath string,
	setupConfig *setup.SetupConfig,
	execErr *setup.SetupExecutionError,
	lastApply *agent.ApplyResult,
	round int,
) (*agentRoundResult, bool, error) {
	decision, applyRes, updatedSetup, changedSomething, err := diagnoseApplyReload(
		ctx,
		orchestrator,
		wd,
		setupPath,
		setupConfig,
		execErr,
		lastApply,
		round,
	)
	if err != nil {
		return nil, false, err
	}

	if !changedSomething {
		return &agentRoundResult{
			Decision:    decision,
			Apply:       applyRes,
			SetupConfig: updatedSetup,
		}, false, nil
	}

	return rerunSetupAndMaybeContinue(
		ctx,
		runner,
		llm,
		runtimeCfg,
		wd,
		updatedSetup,
		execErr,
		decision,
		applyRes,
	)
}

func diagnoseApplyReload(
	ctx context.Context,
	orchestrator *agent.Orchestrator,
	wd string,
	setupPath string,
	setupConfig *setup.SetupConfig,
	execErr *setup.SetupExecutionError,
	lastApply *agent.ApplyResult,
	round int,
) (*agent.Decision, *agent.ApplyResult, *setup.SetupConfig, bool, error) {
	snapshotJSON, snapshot, snapErr := agent.BuildProjectSnapshot(wd)
	if snapErr != nil {
		return nil, nil, nil, false, snapErr
	}

	contextMsg := buildAgentContextMsg(snapshotJSON, snapshot, lastApply)
	decision, err := orchestrator.DiagnoseWithContext(ctx, execErr.Report, contextMsg)
	if err != nil {
		return nil, nil, nil, false, err
	}

	fmt.Println()
	fmt.Printf("Agent round %d diagnosis:\n", round)
	fmt.Println(decision.Explanation)
	fmt.Println()

	beforeSetupFP, _ := fileFingerprint(setupPath)
	applyRes, err := orchestrator.ApplyWithResults(ctx, decision)
	if err != nil {
		return nil, nil, nil, false, err
	}

	afterSetupFP, _ := fileFingerprint(setupPath)
	setupConfig, err = reloadSetupIfChanged(setupPath, setupConfig, beforeSetupFP, afterSetupFP)
	if err != nil {
		return nil, nil, nil, false, err
	}

	if !hasExecutableActions(decision) {
		fmt.Println("Agent suggested no executable actions. Re-run setup after addressing prompts.")
		return nil, nil, nil, false, fmt.Errorf("setup failed; agent provided guidance")
	}

	if !decisionLikelyChangedSomething(decision, applyRes) {
		fmt.Println()
		fmt.Println("No changes applied that warrant re-running setup. Continuing agent mode...")
		return decision, applyRes, setupConfig, false, nil
	}

	return decision, applyRes, setupConfig, true, nil
}

func rerunSetupAndMaybeContinue(
	ctx context.Context,
	runner *setup.SetupRunner,
	llm agent.LLM,
	runtimeCfg *agent.RuntimeConfig,
	wd string,
	setupConfig *setup.SetupConfig,
	execErr *setup.SetupExecutionError,
	decision *agent.Decision,
	applyRes *agent.ApplyResult,
) (*agentRoundResult, bool, error) {
	fmt.Println()
	fmt.Println("Re-running setup...")
	retryErr := runner.Run(setupConfig)
	if retryErr == nil {
		buckets, _ := writeAgentSummary(ctx, llm, runtimeCfg.Model, wd, execErr, decision, nil)
		_ = maybeOfferSetupFixPR(ctx, wd, buckets)
		return &agentRoundResult{
			Decision:    decision,
			Apply:       applyRes,
			SetupConfig: setupConfig,
		}, true, nil
	}

	buckets, _ := writeAgentSummary(ctx, llm, runtimeCfg.Model, wd, execErr, decision, retryErr)
	_ = maybeOfferSetupFixPR(ctx, wd, buckets)

	var nextExecErr *setup.SetupExecutionError
	if !errors.As(retryErr, &nextExecErr) || nextExecErr == nil || nextExecErr.Report == nil {
		return nil, false, fmt.Errorf("run after agent: %w", retryErr)
	}

	ok2, err := prompt.PromptYesNo("Setup still failing. Continue agent mode?")
	if err != nil {
		return nil, false, err
	}
	if !ok2 {
		fmt.Println("Re-run later: initiat setup run")
		return nil, false, fmt.Errorf("setup failed; agent stopped by user")
	}

	return &agentRoundResult{
		Decision:    decision,
		Apply:       applyRes,
		SetupConfig: setupConfig,
		RetryErr:    retryErr,
		NextExecErr: nextExecErr,
	}, false, nil
}

func buildAgentContextMsg(
	snapshotJSON string,
	snapshot *agent.ProjectSnapshot,
	lastApply *agent.ApplyResult,
) string {
	var b strings.Builder
	b.WriteString("Project snapshot (base dir = agent working directory):\n")
	b.WriteString(snapshotJSON)
	b.WriteString("\n\nInstruction:\n")
	b.WriteString("- Read README first using read_files.\n")
	b.WriteString("- Then, if needed, read ONLY important_files to determine app type.\n")
	b.WriteString("- Avoid reading arbitrary docs.\n")

	if snapshot != nil && len(snapshot.ReadmeCandidates) > 0 {
		b.WriteString("\nREADME candidates present: ")
		b.WriteString(strings.Join(snapshot.ReadmeCandidates, ", "))
		b.WriteString("\n")
	}
	if snapshot != nil && len(snapshot.ImportantFiles) > 0 {
		b.WriteString("Important files present: ")
		b.WriteString(strings.Join(snapshot.ImportantFiles, ", "))
		b.WriteString("\n")
	}
	if lastApply != nil && len(lastApply.Results) > 0 {
		b.WriteString("\nPrevious agent action results (do not repeat failures):\n")
		b.WriteString(formatApplyResults(lastApply))
		b.WriteString("\n")
	}

	return strings.TrimSpace(b.String())
}

func formatApplyResults(r *agent.ApplyResult) string {
	if r == nil || len(r.Results) == 0 {
		return "(none)"
	}
	var b strings.Builder
	for _, it := range r.Results {
		status := "ok"
		if !it.OK {
			status = "failed"
		}
		b.WriteString("- ")
		b.WriteString(string(it.Type))
		b.WriteString(": ")
		b.WriteString(it.Summary)
		b.WriteString(" => ")
		b.WriteString(status)
		if strings.TrimSpace(it.Error) != "" {
			b.WriteString(" (")
			b.WriteString(it.Error)
			b.WriteString(")")
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func decisionLikelyChangedSomething(decision *agent.Decision, apply *agent.ApplyResult) bool {
	if decision == nil || apply == nil {
		return false
	}

	results := apply.Results
	for i, a := range decision.Actions {
		if i >= len(results) {
			break
		}
		if !results[i].OK {
			continue
		}
		if a.Type == agent.ActionEditFiles {
			return true
		}
		if a.Type == agent.ActionRunCommand && agent.CommandLikelyMutatesProject(a.Command) {
			return true
		}
	}
	return false
}

func hasExecutableActions(decision *agent.Decision) bool {
	if decision == nil {
		return false
	}
	for _, a := range decision.Actions {
		if a.Type == agent.ActionRunCommand || a.Type == agent.ActionEditFiles {
			return true
		}
	}
	return false
}

type fileFP struct {
	ModTime time.Time
	Size    int64
}

func fileFingerprint(path string) (*fileFP, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return &fileFP{ModTime: fi.ModTime(), Size: fi.Size()}, nil
}

func fileFingerprintChanged(before, after *fileFP) bool {
	if before == nil && after == nil {
		return false
	}
	if before == nil || after == nil {
		return true
	}
	return !before.ModTime.Equal(after.ModTime) || before.Size != after.Size
}

func reloadSetupIfChanged(
	setupPath string,
	current *setup.SetupConfig,
	before *fileFP,
	after *fileFP,
) (*setup.SetupConfig, error) {
	if !fileFingerprintChanged(before, after) {
		return current, nil
	}
	updated, err := setup.ParseFile(setupPath)
	if err != nil {
		return nil, fmt.Errorf("re-parse updated %s: %w", setupPath, err)
	}
	if err := setup.Validate(updated); err != nil {
		return nil, fmt.Errorf("validate updated %s: %w", setupPath, err)
	}
	fmt.Println("Reloaded updated", setupPath)
	return updated, nil
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
) (*agent.IssueBuckets, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, nil
	}
	if execErr == nil || execErr.Report == nil || decision == nil {
		return nil, nil
	}

	buckets, err := agent.AssessIssues(ctx, llm, model, execErr.Report, decision)
	if err != nil {
		return nil, err
	}

	summaryPath := filepath.Join(baseDir, ".initiat", "agent-summary.md")
	if err := os.MkdirAll(filepath.Dir(summaryPath), agentSummaryDirPerm); err != nil {
		return nil, err
	}

	content := buildAgentSummaryMarkdown(execErr, decision, buckets, retryErr)
	if err := os.WriteFile(summaryPath, []byte(content), agentSummaryFilePerm); err != nil {
		return nil, err
	}

	fmt.Println()
	fmt.Println("Wrote agent summary to", summaryPath)
	return buckets, nil
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
		b.WriteString("- ")
		b.WriteString(string(a.Type))
		b.WriteString(" (")
		b.WriteString(string(a.Danger))
		b.WriteString("): ")
		b.WriteString(a.Summary)
		b.WriteString("\n")
		b.WriteString("  - Danger reason: ")
		b.WriteString(a.DangerReason)
		b.WriteString("\n")
		if strings.TrimSpace(a.Reason) != "" {
			b.WriteString("  - Why: ")
			b.WriteString(a.Reason)
			b.WriteString("\n")
		}
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
