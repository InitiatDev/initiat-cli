package treesitter

import "github.com/InitiatDev/initiat-cli/internal/codeanalysis/ast"

func Normalize(root Node) ast.Node {
	return normalizeNode(root)
}

func normalizeNode(n Node) ast.Node {
	out := ast.Node{
		Type:  n.Type(),
		Named: n.IsNamed(),
		Range: ast.Range{
			Start: ast.Point{Byte: n.StartByte(), Row: n.StartPoint().Row, Col: n.StartPoint().Col},
			End:   ast.Point{Byte: n.EndByte(), Row: n.EndPoint().Row, Col: n.EndPoint().Col},
		},
	}

	childCount := n.ChildCount()
	if childCount == 0 {
		return out
	}

	out.Children = make([]ast.Node, 0, childCount)
	fields := make(map[string][]int)

	for i := uint32(0); i < childCount; i++ {
		child := n.Child(i)
		out.Children = append(out.Children, normalizeNode(child))

		field := n.FieldNameForChild(i)
		if field != "" {
			fields[field] = append(fields[field], int(i))
		}
	}

	if len(fields) > 0 {
		out.Fields = fields
	}
	return out
}
