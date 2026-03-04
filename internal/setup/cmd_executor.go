package setup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type CommandExecutor interface {
	Execute(ctx context.Context, req *CommandRequest) (*CommandResult, error)
}

type CommandRequest struct {
	Command    string
	Args       []string
	Env        map[string]string
	WorkingDir string
	Timeout    time.Duration
}

type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	TimedOut bool
}

type realCommandExecutor struct{}

func NewRealCommandExecutor() CommandExecutor {
	return &realCommandExecutor{}
}

func (e *realCommandExecutor) Execute(ctx context.Context, req *CommandRequest) (*CommandResult, error) {
	var shellCmd *exec.Cmd
	commandLine := buildCommandLine(req.Command, req.Args)
	startedAt := time.Now()

	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	if runtime.GOOS == goOSWindows {
		// #nosec G204 -- commandLine is from trusted setup.yml configuration
		shellCmd = exec.CommandContext(ctx, "powershell", "-Command", commandLine)
	} else {
		// #nosec G204 -- commandLine is from trusted setup.yml configuration
		shellCmd = exec.CommandContext(ctx, "/bin/sh", "-c", commandLine)
	}

	shellCmd.Dir = req.WorkingDir
	if shellCmd.Dir == "" {
		wd, err := os.Getwd()
		if err == nil {
			shellCmd.Dir = wd
		}
	}

	env := os.Environ()
	for k, v := range req.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	shellCmd.Env = env

	const outputLimitBytes = 8 * 1024
	stdoutBuf := newLimitedBuffer(outputLimitBytes)
	stderrBuf := newLimitedBuffer(outputLimitBytes)

	shellCmd.Stdout = io.MultiWriter(os.Stdout, stdoutBuf)
	shellCmd.Stderr = io.MultiWriter(os.Stderr, stderrBuf)

	err := shellCmd.Run()
	duration := time.Since(startedAt)

	timedOut := ctx.Err() == context.DeadlineExceeded
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &CommandResult{
		ExitCode: exitCode,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: duration,
		TimedOut: timedOut,
	}, err
}

type limitedBuffer struct {
	buf   bytes.Buffer
	limit int
}

func newLimitedBuffer(limit int) *limitedBuffer {
	return &limitedBuffer{limit: limit}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		return len(p), nil
	}
	remaining := b.limit - b.buf.Len()
	if remaining <= 0 {
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		return len(p), nil
	}
	_, _ = b.buf.Write(p)
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}

func buildCommandLine(command string, args []string) string {
	if len(args) == 0 {
		return command
	}

	var builder strings.Builder
	builder.WriteString(command)
	for _, arg := range args {
		builder.WriteString(" ")
		if strings.Contains(arg, " ") {
			builder.WriteString(`"`)
			builder.WriteString(arg)
			builder.WriteString(`"`)
		} else {
			builder.WriteString(arg)
		}
	}
	return builder.String()
}
