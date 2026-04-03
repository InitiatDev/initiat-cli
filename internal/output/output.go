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
