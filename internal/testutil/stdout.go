package testutil

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

type StdoutCapture struct {
	original *os.File
	writer   *os.File
	reader   *os.File
	buffer   *bytes.Buffer
}

func CaptureStdout() *StdoutCapture {
	original := os.Stdout
	reader, writer, _ := os.Pipe()
	os.Stdout = writer

	return &StdoutCapture{
		original: original,
		writer:   writer,
		reader:   reader,
		buffer:   &bytes.Buffer{},
	}
}

func (c *StdoutCapture) Restore() {
	_ = c.writer.Close()
	os.Stdout = c.original
}

func (c *StdoutCapture) GetOutput() string {
	_ = c.writer.Close()
	_, _ = c.buffer.ReadFrom(c.reader)
	return c.buffer.String()
}

func (c *StdoutCapture) Contains(text string) bool {
	output := c.GetOutput()
	return strings.Contains(output, text)
}

func (c *StdoutCapture) AssertContains(t *testing.T, text string) {
	if !c.Contains(text) {
		t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", text, c.GetOutput())
	}
}

func (c *StdoutCapture) AssertNotContains(t *testing.T, text string) {
	if c.Contains(text) {
		t.Errorf("Expected output to NOT contain '%s', but it did. Output: %s", text, c.GetOutput())
	}
}
