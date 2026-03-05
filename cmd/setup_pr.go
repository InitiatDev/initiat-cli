package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/agent"
	"github.com/InitiatDev/initiat-cli/internal/git"
	"github.com/InitiatDev/initiat-cli/internal/prompt"
	"github.com/InitiatDev/initiat-cli/internal/setupfixpr"
)

const (
	prBranchTimeFormat = "20060102-150405"
)

func maybeOfferSetupFixPR(ctx context.Context, baseDir string, buckets *agent.IssueBuckets) error {
	gitRoot, eligible, ok, err := detectSetupFixPREligibility(ctx, baseDir, buckets)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	fmt.Println()
	fmt.Println("Agent made setup-related changes that look PR-eligible:")
	for _, p := range eligible {
		fmt.Println(" -", p)
	}

	shouldCreatePR, err := prompt.PromptYesNo("Create a PR with these setup fixes?")
	if err != nil {
		return err
	}
	if !shouldCreatePR {
		return nil
	}

	okToProceed, err := prompt.PromptYesNo("Proceed with git operations (branch, commit, push, PR)?")
	if err != nil {
		return err
	}
	if !okToProceed {
		return nil
	}

	return createSetupFixPR(ctx, gitRoot, eligible, buckets)
}

func detectSetupFixPREligibility(
	ctx context.Context,
	baseDir string,
	buckets *agent.IssueBuckets,
) (string, []string, bool, error) {
	if strings.TrimSpace(baseDir) == "" {
		return "", nil, false, nil
	}

	handler := git.NewHandler()
	gitRoot, ok := handler.FindGitRoot(filepath.Join(baseDir, "."))
	if !ok || strings.TrimSpace(gitRoot) == "" {
		return "", nil, false, nil
	}

	statusOut, err := runGit(ctx, gitRoot, "status", "--porcelain")
	if err != nil {
		return "", nil, false, err
	}

	changed := setupfixpr.ParseGitPorcelainPaths(statusOut)
	eligible := setupfixpr.EligibleSetupFixPaths(changed)
	if len(eligible) == 0 {
		return "", nil, false, nil
	}

	if buckets != nil && len(buckets.SetupOrApp) == 0 {
		return "", nil, false, nil
	}

	return gitRoot, eligible, true, nil
}

func createSetupFixPR(ctx context.Context, gitRoot string, eligible []string, buckets *agent.IssueBuckets) error {
	baseBranch := detectBaseBranch(ctx, gitRoot)

	branch := fmt.Sprintf("initiat-agent-setup-fix-%s", time.Now().Format(prBranchTimeFormat))

	if _, err := runGit(ctx, gitRoot, "checkout", "-b", branch); err != nil {
		return err
	}
	if _, err := runGit(ctx, gitRoot, append([]string{"add", "--"}, eligible...)...); err != nil {
		return err
	}
	if _, err := runGit(ctx, gitRoot, "commit", "-m", "chore(setup): agent-guided setup fixes"); err != nil {
		return err
	}
	if _, err := runGit(ctx, gitRoot, "push", "-u", "origin", "HEAD"); err != nil {
		return err
	}

	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Println()
		fmt.Println("GitHub CLI (gh) not found. You can open a PR manually from your pushed branch.")
		return nil
	}

	title := "Setup fixes (agent-guided)"
	body := buildSetupFixPRBody(buckets)
	out, err := runGH(ctx, gitRoot, "pr", "create", "--base", baseBranch, "--title", title, "--body", body)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(strings.TrimSpace(out))
	return nil
}

func buildSetupFixPRBody(buckets *agent.IssueBuckets) string {
	var b strings.Builder
	b.WriteString("## Summary\n")
	b.WriteString("- Agent-guided setup changes to improve onboarding reliability.\n\n")

	if buckets == nil {
		return b.String()
	}

	if len(buckets.SetupOrApp) > 0 {
		b.WriteString("## Setup/App issues addressed\n")
		for _, it := range buckets.SetupOrApp {
			b.WriteString("- " + it + "\n")
		}
		b.WriteString("\n")
	}

	if strings.TrimSpace(buckets.Notes) != "" {
		b.WriteString("## Notes\n")
		b.WriteString(buckets.Notes)
		b.WriteString("\n")
	}

	return b.String()
}

func detectBaseBranch(ctx context.Context, gitRoot string) string {
	out, err := runGit(ctx, gitRoot, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		return setupfixpr.BaseBranchFromOriginHeadRef(out)
	}
	return "main"
}

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	return runBin(ctx, dir, "git", args...)
}

func runGH(ctx context.Context, dir string, args ...string) (string, error) {
	return runBin(ctx, dir, "gh", args...)
}

func runBin(ctx context.Context, dir string, bin string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w\n%s", bin, strings.Join(args, " "), err, strings.TrimSpace(buf.String()))
	}
	return buf.String(), nil
}
