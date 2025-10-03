package testutil

import (
	"fmt"
	"testing"
)

func TestStdoutCapture(t *testing.T) {
	capture := CaptureStdout()
	defer capture.Restore()

	fmt.Println("Hello, World!")
	fmt.Println("This is a test")

	capture.AssertContains(t, "Hello, World!")
	capture.AssertContains(t, "This is a test")
	capture.AssertNotContains(t, "Goodbye")
}

func TestStdoutCapture_GetOutput(t *testing.T) {
	capture := CaptureStdout()
	defer capture.Restore()

	fmt.Print("Line 1")
	fmt.Print("Line 2")

	output := capture.GetOutput()
	expected := "Line 1Line 2"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}
