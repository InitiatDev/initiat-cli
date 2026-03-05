package ast

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestEncode_JSONL(t *testing.T) {
	t.Parallel()

	docs := []Document{
		{
			Version:  "ast-v1",
			Language: "go",
			Path:     "a.go",
			Root: Node{
				Type:  "source_file",
				Named: true,
				Range: Range{Start: Point{}, End: Point{Byte: 1, Row: 0, Col: 1}},
			},
		},
		{
			Version:  "ast-v1",
			Language: "python",
			Path:     "b.py",
			Root: Node{
				Type:  "module",
				Named: true,
				Range: Range{Start: Point{}, End: Point{Byte: 2, Row: 0, Col: 2}},
			},
		},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, FormatJSONL, docs); err != nil {
		t.Fatalf("encode: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}

	var got0, got1 Document
	if err := json.Unmarshal([]byte(lines[0]), &got0); err != nil {
		t.Fatalf("unmarshal line0: %v\n%s", err, lines[0])
	}
	if err := json.Unmarshal([]byte(lines[1]), &got1); err != nil {
		t.Fatalf("unmarshal line1: %v\n%s", err, lines[1])
	}

	if got0.Path != "a.go" || got1.Path != "b.py" {
		t.Fatalf("unexpected order/paths: got0=%q got1=%q", got0.Path, got1.Path)
	}
}

func TestEncode_JSON_Array(t *testing.T) {
	t.Parallel()

	docs := []Document{
		{Version: "ast-v1", Language: "go", Path: "a.go", Root: Node{Type: "source_file", Named: true}},
		{Version: "ast-v1", Language: "go", Path: "b.go", Root: Node{Type: "source_file", Named: true}},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, FormatJSON, docs); err != nil {
		t.Fatalf("encode: %v", err)
	}

	var got []Document
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, buf.String())
	}
	if len(got) != 2 || got[0].Path != "a.go" || got[1].Path != "b.go" {
		t.Fatalf("unexpected docs: %+v", got)
	}
}

func TestEncode_InvalidFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, "nope", []Document{{Version: "ast-v1", Language: "go", Path: "a.go"}})
	if err == nil {
		t.Fatalf("expected error")
	}
}
