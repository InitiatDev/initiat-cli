package scaffold

import (
	"os"
	"path/filepath"
)

const (
	TemplateGo      = "go"
	TemplateNode    = "node"
	TemplatePython  = "python"
	TemplatePhoenix = "phoenix"
	TemplateRails   = "rails"
	TemplateDocker  = "docker"
)

func Detect(dir string) ([]string, error) {
	var out []string
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	if fileExists(dir, "go.mod") {
		out = append(out, TemplateGo)
	}
	if fileExists(dir, "package.json") {
		out = append(out, TemplateNode)
	}
	if fileExists(dir, "assets", "package.json") && !contains(out, TemplateNode) {
		out = append(out, TemplateNode)
	}
	if fileExists(dir, "pyproject.toml") || fileExists(dir, "requirements.txt") {
		out = append(out, TemplatePython)
	}
	if fileExists(dir, "mix.exs") {
		out = append(out, TemplatePhoenix)
	}
	if fileExists(dir, "Gemfile") {
		out = append(out, TemplateRails)
	}
	if fileExists(dir, "docker-compose.yml") || fileExists(dir, "compose.yml") {
		out = append(out, TemplateDocker)
	}
	return out, nil
}

func fileExists(dir string, parts ...string) bool {
	path := filepath.Join(dir, filepath.Join(parts...))
	_, err := os.Stat(path)
	return err == nil
}

func contains(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}
