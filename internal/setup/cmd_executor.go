package setup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type CommandExecutor interface {
	Execute(ctx context.Context, req *CommandRequest) error
}

type CommandRequest struct {
	Command    string
	Args       []string
	Env        map[string]string
	WorkingDir string
	Timeout    time.Duration
}

type realCommandExecutor struct{}

func NewRealCommandExecutor() CommandExecutor {
	return &realCommandExecutor{}
}

func (e *realCommandExecutor) Execute(ctx context.Context, req *CommandRequest) error {
	var shellCmd *exec.Cmd
	commandLine := buildCommandLine(req.Command, req.Args)

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

	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr

	return shellCmd.Run()
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
