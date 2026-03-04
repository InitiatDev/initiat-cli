package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

const (
	initiatDir = ".initiat"
	configFile = "config.yml"
	setupFile  = "setup.yml"
	configPerm = 0644
	dirPerm    = 0755
)

type WriteOptions struct {
	Dir         string
	ProjectName string
	ForceSetup  bool
	ForceConfig bool
}

func Write(config *setup.SetupConfig, opts WriteOptions) (wroteSetup, wroteConfig bool, err error) {
	if opts.Dir == "" {
		opts.Dir, err = os.Getwd()
		if err != nil {
			return false, false, err
		}
	}
	if opts.ProjectName == "" {
		opts.ProjectName = ProjectName(opts.Dir)
	}
	initiatPath := filepath.Join(opts.Dir, initiatDir)
	if err := os.MkdirAll(initiatPath, dirPerm); err != nil {
		return false, false, fmt.Errorf("create %s: %w", initiatDir, err)
	}

	setupPath := filepath.Join(initiatPath, setupFile)
	if _, statErr := os.Stat(setupPath); statErr == nil && !opts.ForceSetup {
		// skip writing setup
	} else {
		setupYAML, marshalErr := yaml.Marshal(config)
		if marshalErr != nil {
			return false, false, fmt.Errorf("marshal setup: %w", marshalErr)
		}
		if writeErr := os.WriteFile(setupPath, setupYAML, configPerm); writeErr != nil {
			return false, false, fmt.Errorf("write %s: %w", setupPath, writeErr)
		}
		wroteSetup = true
	}

	configPath := filepath.Join(initiatPath, configFile)
	if _, statErr := os.Stat(configPath); statErr == nil && !opts.ForceConfig {
		return wroteSetup, false, nil
	}
	configContent := []byte(fmt.Sprintf("name: %q\n", opts.ProjectName))
	if err := os.WriteFile(configPath, configContent, configPerm); err != nil {
		return wroteSetup, false, fmt.Errorf("write %s: %w", configPath, err)
	}
	wroteConfig = true
	return wroteSetup, wroteConfig, nil
}

func SetupExists(dir string) (bool, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return false, err
		}
	}
	path := filepath.Join(dir, initiatDir, setupFile)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
