package scaffold

import (
	"testing"
)

func TestListTemplates(t *testing.T) {
	list := ListTemplates()
	want := map[string]bool{TemplateGo: true, TemplateNode: true, TemplatePython: true, TemplatePhoenix: true, TemplateRails: true, TemplateDocker: true}
	if len(list) != len(want) {
		t.Errorf("ListTemplates: got %d, want %d", len(list), len(want))
	}
	for _, id := range list {
		if !want[id] {
			t.Errorf("unexpected template: %s", id)
		}
	}
}

func TestLoadTemplate_Each(t *testing.T) {
	for _, id := range ListTemplates() {
		t.Run(id, func(t *testing.T) {
			cfg, err := LoadTemplate(id)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Version != 1 {
				t.Errorf("Version: got %d", cfg.Version)
			}
			if cfg.Name == "" {
				t.Error("Name is empty")
			}
		})
	}
}

func TestLoadTemplate_Unknown(t *testing.T) {
	_, err := LoadTemplate("unknown")
	if err == nil {
		t.Error("expected error for unknown template")
	}
}
