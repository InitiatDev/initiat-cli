package ast

type Document struct {
	Version  string         `json:"version"`
	Language string         `json:"language"`
	Path     string         `json:"path"`
	Root     Node           `json:"root"`
	Errors   []ParseError   `json:"errors,omitempty"`
	Meta     map[string]any `json:"meta,omitempty"`
}

type Node struct {
	Type     string           `json:"type"`
	Named    bool             `json:"named"`
	Range    Range            `json:"range"`
	Children []Node           `json:"children,omitempty"`
	Fields   map[string][]int `json:"fields,omitempty"`
}

type ParseError struct {
	Message string `json:"message"`
	Range   Range  `json:"range"`
}

type Range struct {
	Start Point `json:"start"`
	End   Point `json:"end"`
}

type Point struct {
	Byte uint32 `json:"byte"`
	Row  uint32 `json:"row"`
	Col  uint32 `json:"col"`
}
