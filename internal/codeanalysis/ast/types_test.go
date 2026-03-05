package ast

import (
	"encoding/json"
	"testing"
)

func TestDocument_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	doc := Document{
		Version:  "ast-v1",
		Language: "go",
		Path:     "main.go",
		Root: Node{
			Type:  "source_file",
			Named: true,
			Range: Range{
				Start: Point{Byte: 0, Row: 0, Col: 0},
				End:   Point{Byte: 10, Row: 0, Col: 10},
			},
			Children: []Node{
				{
					Type:  "package_clause",
					Named: true,
					Range: Range{
						Start: Point{Byte: 0, Row: 0, Col: 0},
						End:   Point{Byte: 7, Row: 0, Col: 7},
					},
				},
			},
			Fields: map[string][]int{
				"package": {0},
			},
		},
		Errors: []ParseError{
			{
				Message: "example error",
				Range: Range{
					Start: Point{Byte: 2, Row: 0, Col: 2},
					End:   Point{Byte: 3, Row: 0, Col: 3},
				},
			},
		},
		Meta: map[string]any{
			"tool": "initiat",
		},
	}

	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Document
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Version != doc.Version || got.Language != doc.Language || got.Path != doc.Path {
		t.Fatalf("unexpected header: got=%+v want=%+v", got, doc)
	}
	if got.Root.Type != "source_file" || len(got.Root.Children) != 1 {
		t.Fatalf("unexpected root: %+v", got.Root)
	}
	if got.Root.Fields["package"][0] != 0 {
		t.Fatalf("unexpected fields mapping: %+v", got.Root.Fields)
	}
	if len(got.Errors) != 1 || got.Errors[0].Message != "example error" {
		t.Fatalf("unexpected errors: %+v", got.Errors)
	}
	if got.Meta["tool"] != "initiat" {
		t.Fatalf("unexpected meta: %+v", got.Meta)
	}
}
