package testutil

import (
	"io"
	"os"
	"strings"
)

type StdinMock struct {
	original *os.File
	reader   *os.File
	writer   *os.File
}

func MockStdin(input string) (*StdinMock, error) {
	original := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	mock := &StdinMock{
		original: original,
		reader:   reader,
		writer:   writer,
	}

	go func() {
		defer writer.Close()
		_, _ = io.WriteString(writer, input)
	}()

	os.Stdin = reader

	return mock, nil
}

func (m *StdinMock) Restore() {
	_ = m.writer.Close()
	_ = m.reader.Close()
	os.Stdin = m.original
}

func MockStdinWithLines(lines ...string) (*StdinMock, error) {
	return MockStdin(strings.Join(lines, "\n") + "\n")
}
