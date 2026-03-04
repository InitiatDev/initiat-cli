package scaffold

import (
	"path/filepath"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

func Merge(dir string, templateIDs []string, projectName string) (*setup.SetupConfig, error) {
	var configs []*setup.SetupConfig
	for _, id := range templateIDs {
		cfg, err := LoadTemplate(id)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	if len(configs) == 0 {
		return minimalConfig(projectName), nil
	}
	merged := &setup.SetupConfig{
		Version:     1,
		Name:        projectName,
		Description: "Generated setup for this project",
		Matrix:      &setup.Matrix{OS: []string{"macos", "linux", "windows"}, Arch: []string{"x86_64", "arm64"}},
		Defaults:    &setup.Defaults{Timeout: "15m", Shell: "auto", ContinueOnError: false, CWD: "."},
	}
	if configs[0].Defaults != nil {
		merged.Defaults = configs[0].Defaults
	}
	for _, c := range configs {
		merged.Bootstrap = append(merged.Bootstrap, c.Bootstrap...)
		merged.Provision = append(merged.Provision, c.Provision...)
		merged.Setup = append(merged.Setup, c.Setup...)
		merged.Verify = append(merged.Verify, c.Verify...)
		merged.Post = append(merged.Post, c.Post...)
	}
	serviceSteps, err := InferServiceSteps(dir)
	if err == nil && len(serviceSteps.Provision) > 0 {
		merged.Provision = append(merged.Provision, serviceSteps.Provision...)
		if serviceSteps.PostMessage != "" {
			merged.Post = append(merged.Post, setup.Step{Name: "Services", Print: serviceSteps.PostMessage})
		}
	}
	return merged, nil
}

func minimalConfig(projectName string) *setup.SetupConfig {
	return &setup.SetupConfig{
		Version:     1,
		Name:        projectName,
		Description: "Generated setup (no stack detected)",
		Matrix:      &setup.Matrix{OS: []string{"macos", "linux", "windows"}, Arch: []string{"x86_64", "arm64"}},
		Defaults:    &setup.Defaults{Timeout: "15m", Shell: "auto", ContinueOnError: false, CWD: "."},
		Post: []setup.Step{
			{Name: "Next", Print: "Add steps to .initiat/setup.yml for your project. See docs/SETUP_SCRIPTS.md."},
		},
	}
}

func ProjectName(dir string) string {
	if dir == "" {
		return "project"
	}
	return filepath.Base(dir)
}
