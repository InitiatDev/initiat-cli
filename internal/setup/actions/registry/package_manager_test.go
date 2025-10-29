package registry

import (
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

func TestPackageManagerRegistry(t *testing.T) {
	registry := NewPackageManagerRegistry()

	t.Run("FindManager", func(t *testing.T) {
		// Test macOS - should find asdf first
		manager := registry.FindManager(types.OSMacOS)
		if manager == nil {
			t.Error("Expected to find package manager for macOS")
		}
		if manager != nil && manager.Name() != "asdf" {
			t.Errorf("Expected asdf manager, got %s", manager.Name())
		}

		// Test Linux - should find asdf first
		manager = registry.FindManager(types.OSLinux)
		if manager == nil {
			t.Error("Expected to find package manager for Linux")
		}
		if manager != nil && manager.Name() != "asdf" {
			t.Errorf("Expected asdf manager, got %s", manager.Name())
		}

		// Test Windows - should find asdf first
		manager = registry.FindManager(types.OSWindows)
		if manager == nil {
			t.Error("Expected to find package manager for Windows")
		}
		if manager != nil && manager.Name() != "asdf" {
			t.Errorf("Expected asdf manager, got %s", manager.Name())
		}
	})
}
