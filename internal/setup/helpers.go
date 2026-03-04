package setup

import (
	"os"
	"runtime"
)

func collectSecretNames(config *SetupConfig) []string {
	secretMap := make(map[string]bool)

	for _, phase := range GetAllPhases(config) {
		for _, step := range phase.Steps {
			if step.Secrets != nil {
				for _, secretName := range step.Secrets {
					secretMap[secretName] = true
				}
			}
		}
	}

	secretNames := make([]string, 0, len(secretMap))
	for name := range secretMap {
		secretNames = append(secretNames, name)
	}

	return secretNames
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		return shell
	}

	switch runtime.GOOS {
	case goOSWindows:
		return "powershell"
	case goOSDarwin, goOSLinux:
		return "/bin/bash"
	default:
		return "/bin/sh"
	}
}
