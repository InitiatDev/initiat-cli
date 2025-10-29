package registry

import (
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

func TestServiceManagerRegistry(t *testing.T) {
	registry := NewServiceManagerRegistry()

	t.Run("FindManager", func(t *testing.T) {
		// Test macOS
		manager := registry.FindManager(types.OSMacOS)
		if manager == nil {
			t.Error("Expected to find service manager for macOS")
		}
		if manager != nil && manager.Name() != "brew-services" {
			t.Errorf("Expected brew-services manager, got %s", manager.Name())
		}

		// Test Linux
		manager = registry.FindManager(types.OSLinux)
		if manager == nil {
			t.Error("Expected to find service manager for Linux")
		}
		if manager != nil && manager.Name() != "systemd" {
			t.Errorf("Expected systemd manager, got %s", manager.Name())
		}

		// Test Windows
		manager = registry.FindManager(types.OSWindows)
		if manager == nil {
			t.Error("Expected to find service manager for Windows")
		}
		if manager != nil && manager.Name() != "windows-services" {
			t.Errorf("Expected windows-services manager, got %s", manager.Name())
		}
	})
}
