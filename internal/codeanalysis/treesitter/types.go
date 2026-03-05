package treesitter

import (
	"fmt"
	"math"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type ParserHandle struct {
	parser *sitter.Parser
}

func NewParserHandle(lang *sitter.Language) (*ParserHandle, error) {
	if lang == nil {
		return nil, fmt.Errorf("language cannot be nil")
	}
	p := sitter.NewParser()
	if p == nil {
		return nil, fmt.Errorf("failed to create parser")
	}
	if err := p.SetLanguage(lang); err != nil {
		p.Close()
		return nil, fmt.Errorf("set language: %w", err)
	}
	return &ParserHandle{parser: p}, nil
}

func (h *ParserHandle) Close() {
	if h == nil || h.parser == nil {
		return
	}
	h.parser.Close()
	h.parser = nil
}

func (h *ParserHandle) Parse(source []byte) (*TreeHandle, error) {
	if h == nil || h.parser == nil {
		return nil, fmt.Errorf("parser is closed")
	}
	t := h.parser.Parse(source, nil)
	if t == nil {
		return nil, fmt.Errorf("parse returned nil tree")
	}
	return &TreeHandle{tree: t}, nil
}

type TreeHandle struct {
	tree *sitter.Tree
}

func (t *TreeHandle) Close() {
	if t == nil || t.tree == nil {
		return
	}
	t.tree.Close()
	t.tree = nil
}

func (t *TreeHandle) Root() Node {
	if t == nil || t.tree == nil {
		return Node{}
	}
	return Node{n: t.tree.RootNode()}
}

type Node struct {
	n *sitter.Node
}

func uintToUint32Clamped(v uint) uint32 {
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}

func (n Node) Type() string {
	if n.n == nil {
		return ""
	}
	return n.n.Kind()
}
func (n Node) IsNamed() bool {
	if n.n == nil {
		return false
	}
	return n.n.IsNamed()
}

func (n Node) ChildCount() uint32 {
	if n.n == nil {
		return 0
	}
	return uintToUint32Clamped(n.n.ChildCount())
}
func (n Node) Child(i uint32) Node {
	if n.n == nil {
		return Node{}
	}
	return Node{n: n.n.Child(uint(i))}
}

func (n Node) StartByte() uint32 {
	if n.n == nil {
		return 0
	}
	return uintToUint32Clamped(n.n.StartByte())
}
func (n Node) EndByte() uint32 {
	if n.n == nil {
		return 0
	}
	return uintToUint32Clamped(n.n.EndByte())
}

func (n Node) StartPoint() Point {
	if n.n == nil {
		return Point{}
	}
	p := n.n.StartPosition()
	return Point{
		Row: uintToUint32Clamped(p.Row),
		Col: uintToUint32Clamped(p.Column),
	}
}

func (n Node) EndPoint() Point {
	if n.n == nil {
		return Point{}
	}
	p := n.n.EndPosition()
	return Point{
		Row: uintToUint32Clamped(p.Row),
		Col: uintToUint32Clamped(p.Column),
	}
}

func (n Node) FieldNameForChild(childIndex uint32) string {
	if n.n == nil {
		return ""
	}
	return n.n.FieldNameForChild(childIndex)
}

type Point struct {
	Row uint32
	Col uint32
}
