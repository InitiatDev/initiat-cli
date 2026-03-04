package setup

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/expr-lang/expr"
)

type ConditionEvaluator struct {
	OS   string
	Arch string
	Env  map[string]string
}

func NewConditionEvaluator(os, arch string, env map[string]string) *ConditionEvaluator {
	if env == nil {
		env = make(map[string]string)
	}
	return &ConditionEvaluator{
		OS:   os,
		Arch: arch,
		Env:  env,
	}
}

func (e *ConditionEvaluator) Evaluate(condition string) (bool, error) {
	if condition == "" {
		return true, nil
	}

	condition = strings.TrimSpace(condition)
	if condition == "" {
		return true, nil
	}

	env := map[string]interface{}{
		"env": e.Env,
	}

	options := []expr.Option{
		expr.Env(env),
		expr.Function("os", e.osFunction, new(func(string) bool)),
		expr.Function("arch", e.archFunction, new(func(string) bool)),
		expr.Function("file_exists", e.fileExistsFunction, new(func(string) bool)),
		expr.Function("cmd_ok", e.cmdOkFunction, new(func(string) bool)),
	}

	program, err := expr.Compile(condition, options...)
	if err != nil {
		return false, fmt.Errorf("failed to compile condition '%s': %w", condition, err)
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate condition '%s': %w", condition, err)
	}

	// Convert result to bool
	switch v := result.(type) {
	case bool:
		return v, nil
	case string:
		return v != "", nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", v) != "0", nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v) != "0", nil
	case float32, float64:
		return fmt.Sprintf("%v", v) != "0", nil
	default:
		return fmt.Sprintf("%v", v) != "", nil
	}
}

func (e *ConditionEvaluator) ShouldExecuteStep(step *Step) (bool, error) {
	if step.If == "" {
		return true, nil
	}

	return e.Evaluate(step.If)
}

func (e *ConditionEvaluator) osFunction(params ...interface{}) (interface{}, error) {
	if len(params) != 1 {
		return false, fmt.Errorf("os() expects 1 argument")
	}
	name, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("os() argument must be a string")
	}
	return e.OS == name, nil
}

func (e *ConditionEvaluator) archFunction(params ...interface{}) (interface{}, error) {
	if len(params) != 1 {
		return false, fmt.Errorf("arch() expects 1 argument")
	}
	name, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("arch() argument must be a string")
	}
	return e.Arch == name, nil
}

func (e *ConditionEvaluator) fileExistsFunction(params ...interface{}) (interface{}, error) {
	if len(params) != 1 {
		return false, fmt.Errorf("file_exists() expects 1 argument")
	}
	path, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("file_exists() argument must be a string")
	}
	_, err := os.Stat(path)
	return err == nil, nil
}

func (e *ConditionEvaluator) cmdOkFunction(params ...interface{}) (interface{}, error) {
	if len(params) != 1 {
		return false, fmt.Errorf("cmd_ok() expects 1 argument")
	}
	command, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("cmd_ok() argument must be a string")
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return false, nil
	}
	var cmd *exec.Cmd
	if runtime.GOOS == goOSWindows {
		// #nosec G204 -- command is from repo's setup.yml (author-controlled), not external input
		cmd = exec.Command("powershell", "-NoProfile", "-Command", command)
	} else {
		// #nosec G204 -- command is from repo's setup.yml (author-controlled), not external input
		cmd = exec.Command("/bin/sh", "-c", command)
	}
	cmd.Env = os.Environ()
	err := cmd.Run()
	return err == nil, nil
}
