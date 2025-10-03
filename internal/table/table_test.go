package table

import (
	"fmt"
	"testing"

	"github.com/DylanBlakemore/initiat-cli/internal/testutil"
)

func TestTable_Basic(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

	table := New()
	table.SetHeaders("Name", "Age", "City")
	table.AddRow("Alice", "25", "New York")
	table.AddRow("Bob", "30", "London")

	err := table.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	capture.AssertContains(t, "Name")
	capture.AssertContains(t, "Age")
	capture.AssertContains(t, "City")
	capture.AssertContains(t, "Alice")
	capture.AssertContains(t, "Bob")
}

func TestTable_NoHeaders(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

	table := New()
	table.AddRow("Alice", "25")
	table.AddRow("Bob", "30")

	err := table.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	capture.AssertContains(t, "Alice")
}

func TestTable_EmptyTable(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

	table := New()
	table.SetHeaders("Name", "Age")

	err := table.Render()
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	capture.AssertContains(t, "Name")
	capture.AssertContains(t, "Age")
}

func TestQuickTable(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"Alice", "25"},
		{"Bob", "30"},
	}

	err := QuickTable(headers, rows)
	if err != nil {
		t.Fatalf("QuickTable failed: %v", err)
	}

	capture.AssertContains(t, "Alice")
	capture.AssertContains(t, "Bob")
}

func TestTable_AddRows(t *testing.T) {
	table := New()

	rows := [][]string{
		{"Alice", "25"},
		{"Bob", "30"},
		{"Charlie", "35"},
	}

	table.AddRows(rows)

	if len(table.rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(table.rows))
	}

	table.AddRow("David", "40")

	if len(table.rows) != 4 {
		t.Errorf("Expected 4 rows, got %d", len(table.rows))
	}
}

func TestTable_WorkspaceScenario(t *testing.T) {
	capture := testutil.CaptureStdout()
	defer capture.Restore()

	table := New()
	table.SetHeaders("Name", "Composite Slug", "Key Initialized", "Role")

	compositeSlug1 := ""
	if compositeSlug1 == "" {
		compositeSlug1 = fmt.Sprintf("%s/%s", "test-organization", "my-project")
	}
	keyStatus1 := "❌ No"

	compositeSlug2 := ""
	if compositeSlug2 == "" {
		compositeSlug2 = fmt.Sprintf("%s/%s", "team-organization", "team-secrets")
	}
	keyStatus2 := "✅ Yes"

	table.AddRow("My Project", compositeSlug1, keyStatus1, "Owner")
	table.AddRow("Team Secrets", compositeSlug2, keyStatus2, "Member")

	err := table.Render()
	if err != nil {
		t.Fatalf("Table render failed: %v", err)
	}

	capture.AssertContains(t, "My Project")
	capture.AssertContains(t, "test-organization/my-project")
}
