package setup

import (
	"fmt"
	"testing"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions"
)

func TestActionType_String(t *testing.T) {
	testCases := []struct {
		actionType actions.ActionType
		expected   string
	}{
		{actions.ActionTypeRun, "run"},
		{actions.ActionTypePrint, "print"},
		{actions.ActionTypeEnsurePackageManager, "ensure_package_manager"},
		{actions.ActionTypeEnsureTool, "ensure_tool"},
		{actions.ActionTypeEnsureRuntime, "ensure_runtime"},
		{actions.ActionTypeEnsureDatabase, "ensure_database"},
		{actions.ActionTypeAssertCommand, "assert_command"},
		{actions.ActionTypeAssertHTTP, "assert_http"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.actionType), func(t *testing.T) {
			if string(tc.actionType) != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, string(tc.actionType))
			}
		})
	}
}

func TestActionContext_Structure(t *testing.T) {
	ctx := &actions.ActionContext{
		OS:              "linux",
		Arch:            "x86_64",
		Env:             map[string]string{"TEST": "value"},
		Secrets:         map[string]string{"SECRET": "secret_value"},
		WorkingDir:      "/tmp",
		Shell:           "/bin/bash",
		Timeout:         30 * time.Second,
		ContinueOnError: false,
	}

	if ctx.OS != "linux" {
		t.Errorf("Expected OS 'linux', got '%s'", ctx.OS)
	}
	if ctx.Arch != "x86_64" {
		t.Errorf("Expected Arch 'x86_64', got '%s'", ctx.Arch)
	}
	if ctx.Env["TEST"] != "value" {
		t.Errorf("Expected Env TEST=value, got %s", ctx.Env["TEST"])
	}
	if ctx.Secrets["SECRET"] != "secret_value" {
		t.Errorf("Expected Secrets SECRET=secret_value, got %s", ctx.Secrets["SECRET"])
	}
	if ctx.WorkingDir != "/tmp" {
		t.Errorf("Expected WorkingDir '/tmp', got '%s'", ctx.WorkingDir)
	}
	if ctx.Shell != "/bin/bash" {
		t.Errorf("Expected Shell '/bin/bash', got '%s'", ctx.Shell)
	}
	if ctx.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", ctx.Timeout)
	}
	if ctx.ContinueOnError != false {
		t.Errorf("Expected ContinueOnError false, got %v", ctx.ContinueOnError)
	}
}

func TestCommand_Structure(t *testing.T) {
	cmd := actions.Command{
		Command:     "echo",
		Args:        []string{"hello", "world"},
		Env:         map[string]string{"TEST": "value"},
		WorkingDir:  "/tmp",
		Timeout:     30 * time.Second,
		Description: "Test command",
	}

	if cmd.Command != "echo" {
		t.Errorf("Expected command 'echo', got '%s'", cmd.Command)
	}
	if len(cmd.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(cmd.Args))
	}
	if cmd.Env["TEST"] != "value" {
		t.Errorf("Expected env TEST=value, got %s", cmd.Env["TEST"])
	}
	if cmd.WorkingDir != "/tmp" {
		t.Errorf("Expected working dir /tmp, got %s", cmd.WorkingDir)
	}
	if cmd.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", cmd.Timeout)
	}
	if cmd.Description != "Test command" {
		t.Errorf("Expected description 'Test command', got '%s'", cmd.Description)
	}
}

func TestActionError_Error(t *testing.T) {
	err := actions.NewActionError(actions.ActionTypeRun, "test error", nil)
	expected := "run action error: test error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestActionError_ErrorWithWrapped(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := actions.NewActionError(actions.ActionTypeRun, "test error", originalErr)
	expected := "run action error: test error: original error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestActionError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	err := actions.NewActionError(actions.ActionTypeRun, "test error", originalErr)
	if err.Unwrap() != originalErr {
		t.Error("Expected unwrap to return original error")
	}
}

func TestBaseAction_Type(t *testing.T) {
	action := actions.NewBaseAction(actions.ActionTypeRun)
	if action.Type() != actions.ActionTypeRun {
		t.Errorf("Expected type %s, got %s", actions.ActionTypeRun, action.Type())
	}
}

func TestBaseAction_Validate(t *testing.T) {
	action := actions.NewBaseAction(actions.ActionTypeRun)
	if err := action.Validate(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestActionRegistry_RegisterAndGet(t *testing.T) {
	registry := NewActionRegistry()
	action := actions.NewBaseAction(actions.ActionTypeRun)

	registry.Register(action)

	retrieved, exists := registry.Get(actions.ActionTypeRun)
	if !exists {
		t.Error("Expected action to exist")
	}
	if retrieved.Type() != action.Type() {
		t.Error("Expected retrieved action to have same type")
	}
}

func TestActionRegistry_GetNonExistent(t *testing.T) {
	registry := NewActionRegistry()

	_, exists := registry.Get(actions.ActionTypeRun)
	if exists {
		t.Error("Expected action to not exist")
	}
}

func TestActionRegistry_List(t *testing.T) {
	registry := NewActionRegistry()
	action1 := actions.NewBaseAction(actions.ActionTypeRun)
	action2 := actions.NewBaseAction(actions.ActionTypePrint)

	registry.Register(action1)
	registry.Register(action2)

	types := registry.List()
	if len(types) != 2 {
		t.Errorf("Expected 2 action types, got %d", len(types))
	}

	hasRun := false
	hasPrint := false
	for _, actionType := range types {
		if actionType == actions.ActionTypeRun {
			hasRun = true
		}
		if actionType == actions.ActionTypePrint {
			hasPrint = true
		}
	}

	if !hasRun {
		t.Error("Expected ActionTypeRun to be in list")
	}
	if !hasPrint {
		t.Error("Expected ActionTypePrint to be in list")
	}
}

func TestActionRegistry_GetActionFromStep_Run(t *testing.T) {
	registry := NewActionRegistry()
	action := actions.NewBaseAction(actions.ActionTypeRun)
	registry.Register(action)

	step := &Step{
		Run: "echo hello",
	}

	retrieved, err := registry.getActionFromStep(step)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved.Type() != actions.ActionTypeRun {
		t.Errorf("Expected ActionTypeRun, got %s", retrieved.Type())
	}
}

func TestActionRegistry_GetActionFromStep_Print(t *testing.T) {
	registry := NewActionRegistry()
	action := actions.NewBaseAction(actions.ActionTypePrint)
	registry.Register(action)

	step := &Step{
		Print: "Hello world",
	}

	retrieved, err := registry.getActionFromStep(step)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retrieved.Type() != actions.ActionTypePrint {
		t.Errorf("Expected ActionTypePrint, got %s", retrieved.Type())
	}
}

func TestActionRegistry_GetActionFromStep_NoAction(t *testing.T) {
	registry := NewActionRegistry()

	step := &Step{}

	_, err := registry.getActionFromStep(step)
	if err == nil {
		t.Error("Expected error for step with no action")
	}
}

func TestActionRegistry_GetActionFromStep_UnregisteredAction(t *testing.T) {
	registry := NewActionRegistry()

	step := &Step{
		Run: "echo hello",
	}

	_, err := registry.getActionFromStep(step)
	if err == nil {
		t.Error("Expected error for unregistered action")
	}
}
