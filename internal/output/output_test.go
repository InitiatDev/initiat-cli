package output

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func newPlainFormatter(buf *bytes.Buffer) *Formatter {
	return NewFormatter(buf, WithColor(false), WithFancy(false))
}

func newFancyFormatter(buf *bytes.Buffer) *Formatter {
	return NewFormatter(buf, WithColor(false), WithFancy(true))
}

func newColorFormatter(buf *bytes.Buffer) *Formatter {
	return NewFormatter(buf, WithColor(true), WithFancy(true))
}

func TestPhaseStart_Plain(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.PhaseStart("bootstrap")
	if got := buf.String(); got != "== bootstrap ==\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestPhaseStart_Fancy(t *testing.T) {
	var buf bytes.Buffer
	f := newFancyFormatter(&buf)
	f.PhaseStart("bootstrap")
	got := buf.String()
	if !strings.Contains(got, "┌─") || !strings.Contains(got, "bootstrap") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestPhaseEnd_Plain(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.PhaseEnd()
	if got := buf.String(); got != "\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestPhaseEnd_Fancy(t *testing.T) {
	var buf bytes.Buffer
	f := newFancyFormatter(&buf)
	f.PhaseEnd()
	if !strings.Contains(buf.String(), "└─") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestStepSuccess_Plain(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.StepSuccess("Install deps", 2300*time.Millisecond)
	got := buf.String()
	if !strings.Contains(got, "[ok]") || !strings.Contains(got, "Install deps") || !strings.Contains(got, "2.3s") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestStepSuccess_Fancy(t *testing.T) {
	var buf bytes.Buffer
	f := newFancyFormatter(&buf)
	f.StepSuccess("Install deps", 2300*time.Millisecond)
	got := buf.String()
	if !strings.Contains(got, "✓") || !strings.Contains(got, "Install deps") || !strings.Contains(got, "2.3s") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestStepFailure_Plain(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.StepFailure("Run migrations", 800*time.Millisecond, "relation already exists")
	got := buf.String()
	if !strings.Contains(got, "[FAIL]") || !strings.Contains(got, "Run migrations") {
		t.Fatalf("unexpected output: %q", got)
	}
	if !strings.Contains(got, "relation already exists") {
		t.Fatalf("expected error message in output: %q", got)
	}
}

func TestStepFailure_Fancy(t *testing.T) {
	var buf bytes.Buffer
	f := newFancyFormatter(&buf)
	f.StepFailure("Run migrations", 800*time.Millisecond, "relation already exists")
	got := buf.String()
	if !strings.Contains(got, "✗") || !strings.Contains(got, "Run migrations") {
		t.Fatalf("unexpected output: %q", got)
	}
	if !strings.Contains(got, "relation already exists") {
		t.Fatalf("expected error message in output: %q", got)
	}
}

func TestStepFailure_MultilineError(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.StepFailure("Build", 100*time.Millisecond, "line 1\nline 2")
	got := buf.String()
	if strings.Count(got, "line") != 2 {
		t.Fatalf("expected both error lines in output: %q", got)
	}
}

func TestStepFailure_EmptyError(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.StepFailure("Build", 100*time.Millisecond, "")
	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected single line for empty error, got %d: %q", len(lines), got)
	}
}

func TestStepSkipped_Plain(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.StepSkipped("Seed database")
	got := buf.String()
	if !strings.Contains(got, "[skip]") || !strings.Contains(got, "Seed database") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestStepSkipped_Fancy(t *testing.T) {
	var buf bytes.Buffer
	f := newFancyFormatter(&buf)
	f.StepSkipped("Seed database")
	got := buf.String()
	if !strings.Contains(got, "○") || !strings.Contains(got, "Seed database") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestPhasesSkipped_Plain(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.PhasesSkipped([]string{"setup", "verify"})
	got := buf.String()
	if !strings.Contains(got, "setup, verify skipped") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestPhasesSkipped_Fancy(t *testing.T) {
	var buf bytes.Buffer
	f := newFancyFormatter(&buf)
	f.PhasesSkipped([]string{"setup", "verify"})
	got := buf.String()
	if !strings.Contains(got, "└─") || !strings.Contains(got, "setup, verify skipped") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestPhasesSkipped_Empty(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.PhasesSkipped(nil)
	if buf.Len() != 0 {
		t.Fatalf("expected no output for empty slice, got: %q", buf.String())
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d      time.Duration
		expect string
	}{
		{50 * time.Millisecond, "50ms"},
		{999 * time.Millisecond, "999ms"},
		{1500 * time.Millisecond, "1.5s"},
		{59900 * time.Millisecond, "59.9s"},
		{90 * time.Second, "1m30s"},
		{125 * time.Second, "2m5s"},
	}
	for _, tc := range cases {
		t.Run(tc.expect, func(t *testing.T) {
			got := formatDuration(tc.d)
			if got != tc.expect {
				t.Fatalf("formatDuration(%v) = %q, want %q", tc.d, got, tc.expect)
			}
		})
	}
}

func TestColorOutput(t *testing.T) {
	var buf bytes.Buffer
	f := newColorFormatter(&buf)
	f.StepSuccess("test", time.Second)
	got := buf.String()
	if !strings.Contains(got, "\033[") {
		t.Fatalf("expected ANSI codes in color output: %q", got)
	}
}

func TestNoColorOutput(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)
	f.StepSuccess("test", time.Second)
	got := buf.String()
	if strings.Contains(got, "\033[") {
		t.Fatalf("unexpected ANSI codes in plain output: %q", got)
	}
}

func TestNewFormatter_NonTTY(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf) // bytes.Buffer is not a TTY
	if f.color {
		t.Fatal("expected color=false for non-TTY writer")
	}
	if f.fancy {
		t.Fatal("expected fancy=false for non-TTY writer")
	}
}

func TestNewFormatter_NO_COLOR_Env(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// Even with a real file (which could be a TTY), NO_COLOR disables color.
	f := NewFormatter(os.Stdout)
	if f.color {
		t.Fatal("expected color=false when NO_COLOR is set")
	}
}

func TestFullPhaseSequence(t *testing.T) {
	var buf bytes.Buffer
	f := newPlainFormatter(&buf)

	f.PhaseStart("bootstrap")
	f.StepSuccess("Install deps", 2*time.Second)
	f.StepSuccess("Check runtime", 150*time.Millisecond)
	f.PhaseEnd()

	f.PhaseStart("provision")
	f.StepFailure("Run migrations", 800*time.Millisecond, "connection refused")
	f.PhaseEnd()

	f.PhasesSkipped([]string{"setup", "verify"})

	got := buf.String()
	if !strings.Contains(got, "== bootstrap ==") {
		t.Fatal("missing bootstrap phase")
	}
	if !strings.Contains(got, "[ok] Install deps") {
		t.Fatal("missing success step")
	}
	if !strings.Contains(got, "[FAIL] Run migrations") {
		t.Fatal("missing failure step")
	}
	if !strings.Contains(got, "setup, verify skipped") {
		t.Fatal("missing skipped phases")
	}
}
