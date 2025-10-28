package setup

import (
	"fmt"
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
		"os":   e.OS,
		"arch": e.Arch,
		"env":  e.Env,
	}

	result, err := expr.Eval(condition, env)
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
