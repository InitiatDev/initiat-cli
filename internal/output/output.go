// Package output provides structured, visually clear formatting for setup
// execution and agent mode output.
package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// Formatter writes structured, visually clear output for setup execution
// and agent mode. It supports colored and plain-text modes, automatically
// detecting non-TTY environments.
type Formatter struct {
	w     io.Writer
	color bool
	fancy bool // box-drawing characters enabled
}

// NewFormatter creates a Formatter that writes to w. Color and box-drawing
// characters are enabled by default unless:
//   - The NO_COLOR environment variable is set (any value), or
//   - w is not a terminal (non-TTY).
//
// Both can be overridden with Option functions.
func NewFormatter(w io.Writer, opts ...Option) *Formatter {
	isTTY := isTerminal(w)
	noColor := os.Getenv("NO_COLOR") != ""

	f := &Formatter{
		w:     w,
		color: isTTY && !noColor,
		fancy: isTTY,
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// Option configures a Formatter.
type Option func(*Formatter)

// WithColor forces color output on or off.
func WithColor(enabled bool) Option {
	return func(f *Formatter) {
		f.color = enabled
	}
}

// WithFancy forces box-drawing characters on or off.
func WithFancy(enabled bool) Option {
	return func(f *Formatter) {
		f.fancy = enabled
	}
}

// --- Phase lifecycle ---

// PhaseStart prints the opening line of a phase block.
//
// Fancy:  ┌─ bootstrap
// Plain:  == bootstrap ==
func (f *Formatter) PhaseStart(name string) {
	if f.fancy {
		fmt.Fprintf(f.w, "%s %s\n", f.dim("┌─"), f.bold(name))
	} else {
		fmt.Fprintf(f.w, "== %s ==\n", name)
	}
}

// PhaseEnd prints the closing line of a phase block.
//
// Fancy:  └─
// Plain:  (blank line)
func (f *Formatter) PhaseEnd() {
	if f.fancy {
		fmt.Fprintln(f.w, f.dim("└─"))
	} else {
		fmt.Fprintln(f.w)
	}
}

// --- Step results ---

// StepSuccess prints a successful step with its duration.
//
// Fancy:  │  ✓ Install deps  2.3s
// Plain:  [ok] Install deps  2.3s
func (f *Formatter) StepSuccess(name string, d time.Duration) {
	dur := formatDuration(d)
	if f.fancy {
		fmt.Fprintf(f.w, "%s  %s %s  %s\n",
			f.dim("│"), f.green("✓"), name, f.dim(dur))
	} else {
		fmt.Fprintf(f.w, "[ok] %s  %s\n", name, dur)
	}
}

// StepFailure prints a failed step with its duration and error message.
//
// Fancy:  │  ✗ Run migrations  0.8s
// Fancy:  │    error: relation "users" already exists
// Plain:  [FAIL] Run migrations  0.8s
// Plain:         error: relation "users" already exists
func (f *Formatter) StepFailure(name string, d time.Duration, errMsg string) {
	dur := formatDuration(d)
	if f.fancy {
		fmt.Fprintf(f.w, "%s  %s %s  %s\n",
			f.dim("│"), f.red("✗"), name, f.dim(dur))
		if errMsg != "" {
			for _, line := range strings.Split(errMsg, "\n") {
				fmt.Fprintf(f.w, "%s    %s\n", f.dim("│"), f.red(line))
			}
		}
	} else {
		fmt.Fprintf(f.w, "[FAIL] %s  %s\n", name, dur)
		if errMsg != "" {
			for _, line := range strings.Split(errMsg, "\n") {
				fmt.Fprintf(f.w, "       %s\n", line)
			}
		}
	}
}

// StepSkipped prints a skipped step.
//
// Fancy:  │  ○ Seed database (skipped)
// Plain:  [skip] Seed database
func (f *Formatter) StepSkipped(name string) {
	if f.fancy {
		fmt.Fprintf(f.w, "%s  %s %s %s\n",
			f.dim("│"), f.yellow("○"), name, f.dim("(skipped)"))
	} else {
		fmt.Fprintf(f.w, "[skip] %s\n", name)
	}
}

// PhasesSkipped prints a summary of phases that were never reached
// (e.g. due to an earlier failure).
//
// Fancy:  └─ (setup, verify skipped)
// Plain:  -- (setup, verify skipped) --
func (f *Formatter) PhasesSkipped(names []string) {
	if len(names) == 0 {
		return
	}
	label := strings.Join(names, ", ") + " skipped"
	if f.fancy {
		fmt.Fprintf(f.w, "%s (%s)\n", f.dim("└─"), f.dim(label))
	} else {
		fmt.Fprintf(f.w, "-- (%s) --\n", label)
	}
}

// --- Agent mode ---

// AgentHeader prints the agent-mode banner with the failure context.
//
// Fancy:
//
//	╔══ Agent Mode ═══════════════════════════════════════╗
//	║ Diagnosing failure in provision → "Run migrations"  ║
//	╚═════════════════════════════════════════════════════╝
//
// Plain:
//
//	=== Agent Mode ===
//	Diagnosing failure in <phase> -> "<step>"
func (f *Formatter) AgentHeader(phase, step string) {
	desc := fmt.Sprintf("Diagnosing failure in %s", phase)
	if step != "" {
		desc += fmt.Sprintf(" → %q", step)
	}

	if !f.fancy {
		fmt.Fprintf(f.w, "=== Agent Mode ===\n%s\n\n", desc)
		return
	}

	const (
		minWidth     = 50
		sidePadding  = 2 // 1 space padding each side
		rightPadding = 1 // space before closing ║
	)
	contentWidth := len(desc) + sidePadding
	if contentWidth < minWidth {
		contentWidth = minWidth
	}

	top := "╔══ Agent Mode " + strings.Repeat("═", contentWidth-len("══ Agent Mode ")) + "╗"
	padded := "║ " + desc + strings.Repeat(" ", contentWidth-len(desc)-rightPadding) + "║"
	bottom := "╚" + strings.Repeat("═", contentWidth) + "╝"

	fmt.Fprintln(f.w, f.bold(top))
	fmt.Fprintln(f.w, f.bold(padded))
	fmt.Fprintln(f.w, f.bold(bottom))
	fmt.Fprintln(f.w)
}

// RoundSeparator prints a visual separator between agent rounds.
//
// Fancy:  ── Round 1 ──────────────────
// Plain:  -- Round 1 --
func (f *Formatter) RoundSeparator(round int) {
	label := fmt.Sprintf("Round %d", round)
	const roundLineWidth = 40
	if f.fancy {
		line := "── " + label + " " + strings.Repeat("─", roundLineWidth-len(label))
		fmt.Fprintln(f.w, f.dim(line))
	} else {
		fmt.Fprintf(f.w, "-- %s --\n", label)
	}
}

// Explanation prints the agent's diagnosis explanation in a visually
// distinct block (indented/quoted).
//
// Fancy:  │ The migration failed because …
// Plain:  > The migration failed because …
func (f *Formatter) Explanation(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	fmt.Fprintln(f.w)
	for _, line := range strings.Split(text, "\n") {
		if f.fancy {
			fmt.Fprintf(f.w, "  %s %s\n", f.dim("│"), line)
		} else {
			fmt.Fprintf(f.w, "  > %s\n", line)
		}
	}
	fmt.Fprintln(f.w)
}

// ActionItem describes a single proposed action for display purposes.
type ActionItem struct {
	Summary string
	Danger  string // "safe", "caution", or "dangerous"
	Type    string // action type like "run_command", "edit_files", etc.
	Detail  string // command string, file path, prompt, etc.
}

// ActionList prints proposed actions as a numbered list with danger-level badges.
//
// Fancy:
//
//	Proposed actions:
//	  1. [safe]    Read db/migrate/ directory listing
//	  2. [caution] Run: rails db:migrate:status
//
// Plain:
//
//	Proposed actions:
//	  1. [safe] Read db/migrate/ directory listing
//	  2. [caution] Run: rails db:migrate:status
func (f *Formatter) ActionList(actions []ActionItem) {
	if len(actions) == 0 {
		return
	}
	fmt.Fprintln(f.w, "Proposed actions:")
	for i, a := range actions {
		badge := f.dangerBadge(a.Danger)
		fmt.Fprintf(f.w, "  %d. %s %s\n", i+1, badge, a.Summary)
	}
	fmt.Fprintln(f.w)
}

// ActionResult prints the result of an executed action inline.
//
// Fancy:
//
//  1. [safe]    Read db/migrate/ directory listing  ✓
//  2. [caution] Run: rails db:migrate:status        ✓ (exit 0)
//  3. [caution] Run: rails db:drop                  ✗ (permission denied)
//
// Plain:
//
//  1. [safe] Read db/migrate/ directory listing  [ok]
//  2. [caution] Run: rails db:migrate:status  [ok] (exit 0)
//  3. [caution] Run: rails db:drop  [FAIL] (permission denied)
func (f *Formatter) ActionResult(index int, action ActionItem, ok bool, detail string) {
	badge := f.dangerBadge(action.Danger)
	suffix := ""
	if detail != "" {
		suffix = " (" + detail + ")"
	}
	if ok {
		if f.fancy {
			fmt.Fprintf(f.w, "  %d. %s %s  %s%s\n",
				index+1, badge, action.Summary, f.green("✓"), f.dim(suffix))
		} else {
			fmt.Fprintf(f.w, "  %d. %s %s  [ok]%s\n",
				index+1, badge, action.Summary, suffix)
		}
	} else {
		if f.fancy {
			fmt.Fprintf(f.w, "  %d. %s %s  %s%s\n",
				index+1, badge, action.Summary, f.red("✗"), f.red(suffix))
		} else {
			fmt.Fprintf(f.w, "  %d. %s %s  [FAIL]%s\n",
				index+1, badge, action.Summary, suffix)
		}
	}
}

const (
	dangerBadgeWidth = 11 // len("[dangerous]")
)

func (f *Formatter) dangerBadge(danger string) string {
	var raw string
	switch danger {
	case "safe":
		raw = "[safe]"
	case "caution":
		raw = "[caution]"
	case "dangerous":
		raw = "[dangerous]"
	default:
		raw = "[" + danger + "]"
	}

	if !f.fancy {
		return raw
	}

	padded := raw + strings.Repeat(" ", dangerBadgeWidth-len(raw))
	switch danger {
	case "safe":
		return f.green(padded)
	case "caution":
		return f.yellow(padded)
	case "dangerous":
		return f.red(padded)
	default:
		return padded
	}
}

// --- ANSI helpers ---

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
)

func (f *Formatter) green(s string) string {
	if !f.color {
		return s
	}
	return ansiGreen + s + ansiReset
}

func (f *Formatter) red(s string) string {
	if !f.color {
		return s
	}
	return ansiRed + s + ansiReset
}

func (f *Formatter) yellow(s string) string {
	if !f.color {
		return s
	}
	return ansiYellow + s + ansiReset
}

func (f *Formatter) bold(s string) string {
	if !f.color {
		return s
	}
	return ansiBold + s + ansiReset
}

func (f *Formatter) dim(s string) string {
	if !f.color {
		return s
	}
	return ansiDim + s + ansiReset
}

// --- Utilities ---

// formatDuration renders a duration as a human-friendly string.
func formatDuration(d time.Duration) string {
	const secondsPerMinute = 60
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.1fs", d.Seconds())
	default:
		m := int(d.Minutes())
		s := int(d.Seconds()) % secondsPerMinute
		return fmt.Sprintf("%dm%ds", m, s)
	}
}

// isTerminal reports whether w is connected to a terminal.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
