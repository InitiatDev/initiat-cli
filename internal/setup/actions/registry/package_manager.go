package registry

import (
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/strategies/package_managers"
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// PackageManager defines a simple interface for package managers
type PackageManager = types.PackageManager

// PackageManagerRegistry manages available package manager strategies
type PackageManagerRegistry struct {
	managers []PackageManager
}

// NewPackageManagerRegistry creates a new registry with default package managers
func NewPackageManagerRegistry() *PackageManagerRegistry {
	return &PackageManagerRegistry{
		managers: []PackageManager{
			&package_managers.AsdfPackageManager{},
			&package_managers.BrewPackageManager{},
			&package_managers.AptPackageManager{},
			&package_managers.ChocoPackageManager{},
		},
	}
}

// FindManager finds a suitable package manager for the given OS
func (r *PackageManagerRegistry) FindManager(os string) PackageManager {
	for _, manager := range r.managers {
		if manager.SupportsOS(os) {
			return manager
		}
	}
	return nil
}

// HasAvailableManagers checks if any package managers are available for the given OS
func (r *PackageManagerRegistry) HasAvailableManagers(os string) bool {
	return r.FindManager(os) != nil
}
