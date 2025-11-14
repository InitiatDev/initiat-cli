package setup

import (
	"testing"
)

func TestConditionEvaluator_Evaluate_EmptyCondition(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []string{"", "   ", "\t\n"}

	for _, condition := range testCases {
		t.Run("empty_"+condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(condition)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if !result {
				t.Error("Expected empty condition to evaluate to true")
			}
		})
	}
}

func TestConditionEvaluator_Evaluate_OSConditions(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []struct {
		condition string
		expected  bool
	}{
		{`os("macos")`, true},
		{`os("linux")`, false},
		{`!os("windows")`, true},
		{`!os("macos")`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(tc.condition)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestConditionEvaluator_Evaluate_ArchConditions(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []struct {
		condition string
		expected  bool
	}{
		{`arch("x86_64")`, true},
		{`arch("arm64")`, false},
		{`!arch("arm64")`, true},
		{`!arch("x86_64")`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(tc.condition)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestConditionEvaluator_Evaluate_EnvConditions(t *testing.T) {
	env := map[string]string{
		"NODE_ENV": "development",
		"VERSION":  "1.2.3",
		"DEBUG":    "true",
	}
	evaluator := NewConditionEvaluator("macos", "x86_64", env)

	testCases := []struct {
		condition string
		expected  bool
	}{
		{`env.NODE_ENV == "development"`, true},
		{`env.NODE_ENV == "production"`, false},
		{`env.VERSION == "1.2.3"`, true},
		{`env.VERSION != "2.0.0"`, true},
		{`env.DEBUG == "true"`, true},
		{`env.MISSING == "value"`, false},
		{`env.MISSING == ""`, true},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(tc.condition)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestConditionEvaluator_Evaluate_StringConditions(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []struct {
		condition string
		expected  bool
	}{
		{`"hello" == "hello"`, true},
		{`"hello" == "world"`, false},
		{`'hello' == 'hello'`, true},
		{`'hello' == "hello"`, true},
		{`"hello" != "world"`, true},
		{`"hello" != "hello"`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(tc.condition)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestConditionEvaluator_Evaluate_NumberConditions(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []struct {
		condition string
		expected  bool
	}{
		{`5 == 5`, true},
		{`5 == 3`, false},
		{`5 != 3`, true},
		{`5 != 5`, false},
		{`5 > 3`, true},
		{`5 > 5`, false},
		{`5 >= 5`, true},
		{`5 >= 3`, true},
		{`5 >= 7`, false},
		{`5 < 7`, true},
		{`5 < 5`, false},
		{`5 <= 5`, true},
		{`5 <= 7`, true},
		{`5 <= 3`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(tc.condition)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestConditionEvaluator_Evaluate_BooleanConditions(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []struct {
		condition string
		expected  bool
	}{
		{`true == true`, true},
		{`true == false`, false},
		{`false == false`, true},
		{`true != false`, true},
		{`true != true`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(tc.condition)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestConditionEvaluator_Evaluate_StringContains(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []struct {
		condition string
		expected  bool
	}{
		{`os("macos")`, true},
		{`os("linux")`, false},
		{`arch("x86_64")`, true},
		{`arch("arm64")`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(tc.condition)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestConditionEvaluator_Evaluate_InvalidConditions(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []string{
		"os ==",
		"== os",
		"os ===",
	}

	for _, condition := range testCases {
		t.Run(condition, func(t *testing.T) {
			_, err := evaluator.Evaluate(condition)
			if err == nil {
				t.Errorf("Expected error for invalid condition: %s", condition)
			}
		})
	}
}

func TestConditionEvaluator_ShouldExecuteStep_NoCondition(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)
	step := &Step{
		Name: "Test step",
		Run:  "echo hello",
	}

	shouldExecute, err := evaluator.ShouldExecuteStep(step)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !shouldExecute {
		t.Error("Expected step without condition to execute")
	}
}

func TestConditionEvaluator_ShouldExecuteStep_WithCondition(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []struct {
		condition string
		expected  bool
	}{
		{`os("macos")`, true},
		{`os("linux")`, false},
		{`arch("x86_64")`, true},
		{`arch("arm64")`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			step := &Step{
				Name: "Test step",
				Run:  "echo hello",
				If:   tc.condition,
			}

			shouldExecute, err := evaluator.ShouldExecuteStep(step)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if shouldExecute != tc.expected {
				t.Errorf("Expected %v, got %v for condition: %s", tc.expected, shouldExecute, tc.condition)
			}
		})
	}
}

func TestConditionEvaluator_ComplexConditions(t *testing.T) {
	env := map[string]string{
		"NODE_ENV": "development",
		"VERSION":  "1.2.3",
	}
	evaluator := NewConditionEvaluator("macos", "x86_64", env)

	testCases := []struct {
		condition string
		expected  bool
	}{
		{`os("macos") && arch("x86_64")`, true},
		{`os("macos") || arch("arm64")`, true},
		{`env.NODE_ENV == "development" && os("macos")`, true},
		{`env.VERSION contains "1.2" && arch("x86_64")`, true},
		{`os("linux") && arch("x86_64")`, false},
		{`os("macos") && arch("arm64")`, false},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(tc.condition)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for condition: %s", tc.expected, result, tc.condition)
			}
		})
	}
}

func TestConditionEvaluator_EdgeCases(t *testing.T) {
	evaluator := NewConditionEvaluator("macos", "x86_64", nil)

	testCases := []struct {
		condition string
		expected  bool
		expectErr bool
	}{
		{`"" == ""`, true, false},
		{`"" != "hello"`, true, false},
		{`0 == 0`, true, false},
		{`0 != 1`, true, false},
		{`"0" == "0"`, true, false},
		{`"0" == 0`, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.condition, func(t *testing.T) {
			result, err := evaluator.Evaluate(tc.condition)
			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error for condition: %s", tc.condition)
				}
				return
			}
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
