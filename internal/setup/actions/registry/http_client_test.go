package registry

import (
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

func TestHTTPClientRegistry(t *testing.T) {
	registry := NewHTTPClientRegistry()

	t.Run("FindClient", func(t *testing.T) {
		client := registry.FindClient(types.OSMacOS)
		if client == nil {
			t.Log("No HTTP client found for macOS (curl might not be available)")
		}

		client = registry.FindClient(types.OSLinux)
		if client == nil {
			t.Log("No HTTP client found for Linux (curl or wget might not be available)")
		}

		client = registry.FindClient(types.OSWindows)
		if client == nil {
			t.Log("No HTTP client found for Windows (PowerShell might not be available)")
		}
	})
}
