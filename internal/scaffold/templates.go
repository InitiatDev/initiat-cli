package scaffold

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

//go:embed templates/go.yml
var goYAML []byte

//go:embed templates/node.yml
var nodeYAML []byte

//go:embed templates/python.yml
var pythonYAML []byte

//go:embed templates/phoenix.yml
var phoenixYAML []byte

//go:embed templates/rails.yml
var railsYAML []byte

//go:embed templates/docker.yml
var dockerYAML []byte

var templateData = map[string][]byte{
	TemplateGo:      goYAML,
	TemplateNode:    nodeYAML,
	TemplatePython:  pythonYAML,
	TemplatePhoenix: phoenixYAML,
	TemplateRails:   railsYAML,
	TemplateDocker:  dockerYAML,
}

func LoadTemplate(id string) (*setup.SetupConfig, error) {
	data, ok := templateData[id]
	if !ok {
		return nil, fmt.Errorf("unknown template: %s", id)
	}
	var cfg setup.SetupConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("template %s: %w", id, err)
	}
	cfg.Version = 1
	return &cfg, nil
}

func ListTemplates() []string {
	return []string{TemplateGo, TemplateNode, TemplatePython, TemplatePhoenix, TemplateRails, TemplateDocker}
}
