package table

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

type Table struct {
	headers []string
	rows    [][]string
	writer  io.Writer
}

func New() *Table {
	return &Table{
		writer: os.Stdout,
	}
}

func NewWithWriter(w io.Writer) *Table {
	return &Table{
		writer: w,
	}
}

func (t *Table) SetHeaders(headers ...string) *Table {
	t.headers = headers
	return t
}

func (t *Table) AddRow(values ...string) *Table {
	t.rows = append(t.rows, values)
	return t
}

func (t *Table) AddRows(rows [][]string) *Table {
	t.rows = append(t.rows, rows...)
	return t
}

func (t *Table) Render() error {
	w := tabwriter.NewWriter(t.writer, 0, 0, 1, ' ', 0)

	if len(t.headers) > 0 {
		fmt.Fprintln(w, strings.Join(t.headers, "\t"))

		colWidths := t.calculateColumnWidths()
		separator := make([]string, len(t.headers))
		for i, width := range colWidths {
			separator[i] = strings.Repeat("─", width)
		}
		fmt.Fprintln(w, strings.Join(separator, "\t"))
	}

	for _, row := range t.rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return w.Flush()
}

func (t *Table) calculateColumnWidths() []int {
	if len(t.rows) == 0 {
		return []int{}
	}

	colWidths := make([]int, len(t.headers))
	for i, header := range t.headers {
		colWidths[i] = len(header)
	}

	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	return colWidths
}

func QuickTable(headers []string, rows [][]string) error {
	table := New()
	if len(headers) > 0 {
		table.SetHeaders(headers...)
	}
	table.AddRows(rows)
	return table.Render()
}
