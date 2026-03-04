package scaffold

import (
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

func TestProjectName(t *testing.T) {
	if got := ProjectName(""); got != "project" {
		t.Errorf("ProjectName(\"\"): got %q", got)
	}
	if got := ProjectName("/foo/bar/baz"); got != "baz" {
		t.Errorf("ProjectName: got %q", got)
	}
}

func TestMerge_NoTemplates(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Merge(dir, nil, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 {
		t.Errorf("Version: got %d", cfg.Version)
	}
	if cfg.Name != "myapp" {
		t.Errorf("Name: got %q", cfg.Name)
	}
	if len(cfg.Post) == 0 {
		t.Error("minimal config should have post step")
	}
	if err := setup.Validate(cfg); err != nil {
		t.Errorf("merged config invalid: %v", err)
	}
}

func TestMerge_SingleTemplate(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Merge(dir, []string{TemplateGo}, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 {
		t.Errorf("Version: got %d", cfg.Version)
	}
	if len(cfg.Setup) == 0 {
		t.Error("Go template should have setup steps")
	}
	if err := setup.Validate(cfg); err != nil {
		t.Errorf("merged config invalid: %v", err)
	}
}

func TestMerge_MultipleTemplates(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Merge(dir, []string{TemplateGo, TemplateNode}, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 {
		t.Errorf("Version: got %d", cfg.Version)
	}
	goSteps := 0
	for _, s := range cfg.Setup {
		if s.Run != "" && (s.Run == "go mod download" || s.Run == "npm ci" || s.Run == "npm install") {
			goSteps++
		}
	}
	if goSteps < 2 {
		t.Error("expected both Go and Node setup steps")
	}
	if err := setup.Validate(cfg); err != nil {
		t.Errorf("merged config invalid: %v", err)
	}
}

func TestMerge_InvalidTemplateID(t *testing.T) {
	dir := t.TempDir()
	_, err := Merge(dir, []string{"invalid"}, "myapp")
	if err == nil {
		t.Error("expected error for invalid template id")
	}
}
