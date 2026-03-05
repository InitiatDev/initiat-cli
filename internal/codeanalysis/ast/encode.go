package ast

import (
	"encoding/json"
	"fmt"
	"io"
)

type Format string

const (
	FormatJSON  Format = "json"
	FormatJSONL Format = "jsonl"
)

func Encode(w io.Writer, format Format, docs []Document) error {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		return enc.Encode(docs)
	case FormatJSONL:
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		for _, doc := range docs {
			if err := enc.Encode(doc); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported format %q", string(format))
	}
}
