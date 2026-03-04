package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInferServiceSteps_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	out, err := InferServiceSteps(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Provision) != 0 {
		t.Errorf("empty dir: got %d provision steps", len(out.Provision))
	}
}

func TestInferServiceSteps_ComposePreference(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("services:\n  db:\n    image: postgres\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := InferServiceSteps(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Provision) < 2 {
		t.Errorf("compose: expected at least 2 provision steps, got %d", len(out.Provision))
	}
	if out.PostMessage == "" {
		t.Error("compose: expected PostMessage")
	}
}

func TestInferServiceSteps_GemfilePostgres(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Gemfile"), []byte("gem 'pg'\ngem 'rails'\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := InferServiceSteps(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Provision) == 0 {
		t.Error("Gemfile with pg: expected provision steps")
	}
	found := false
	for _, s := range out.Provision {
		if s.Name == "Verify PostgreSQL" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Verify PostgreSQL step")
	}
}

func TestInferServiceSteps_GemfileRedis(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Gemfile"), []byte("gem 'redis'\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := InferServiceSteps(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range out.Provision {
		if s.Name == "Verify Redis" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Verify Redis step")
	}
}

func TestInferServiceSteps_MixPostgrex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mix.exs"), []byte("defp deps do\n  [{:postgrex, \"~> 0.16\"}]\nend\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := InferServiceSteps(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range out.Provision {
		if s.Name == "Verify PostgreSQL" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Verify PostgreSQL step for mix.exs postgrex")
	}
}

func TestInferServiceSteps_PackageJsonPg(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"pg":"^8.0.0"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := InferServiceSteps(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range out.Provision {
		if s.Name == "Verify PostgreSQL" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Verify PostgreSQL step for package.json pg")
	}
}

func TestInferServiceSteps_RequirementsTxtPsycopg(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("psycopg2-binary\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := InferServiceSteps(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range out.Provision {
		if s.Name == "Verify PostgreSQL" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Verify PostgreSQL step for requirements.txt psycopg")
	}
}
