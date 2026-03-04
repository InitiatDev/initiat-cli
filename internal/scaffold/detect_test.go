package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	ids, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Errorf("empty dir: got %v", ids)
	}
}

func TestDetect_Go(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	ids, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != TemplateGo {
		t.Errorf("go.mod: got %v", ids)
	}
}

func TestDetect_Node(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	ids, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != TemplateNode {
		t.Errorf("package.json: got %v", ids)
	}
}

func TestDetect_NodeAssetsOnly(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	ids, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != TemplateNode {
		t.Errorf("assets/package.json: got %v", ids)
	}
}

func TestDetect_Python(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	ids, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != TemplatePython {
		t.Errorf("requirements.txt: got %v", ids)
	}
}

func TestDetect_Phoenix(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mix.exs"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	ids, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != TemplatePhoenix {
		t.Errorf("mix.exs: got %v", ids)
	}
}

func TestDetect_Rails(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Gemfile"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	ids, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != TemplateRails {
		t.Errorf("Gemfile: got %v", ids)
	}
}

func TestDetect_Docker(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	ids, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != TemplateDocker {
		t.Errorf("docker-compose.yml: got %v", ids)
	}
}

func TestDetect_Multiple(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"go.mod", "package.json"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}
	ids, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 templates, got %v", ids)
	}
}
