package registry

import (
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/strategies/package_managers"
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

type PackageManager = types.PackageManager
type SystemPackageManager = types.SystemPackageManager
type RuntimeManager = types.RuntimeManager

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

// FindSystemPackageManager finds a system package manager (brew, apt, choco) for the given OS,
// excluding version managers like asdf
func (r *PackageManagerRegistry) FindSystemPackageManager(os string) SystemPackageManager {
	systemManagerNames := map[string]bool{
		"brew":   true,
		"apt":    true,
		"yum":    true,
		"dnf":    true,
		"pacman": true,
		"choco":  true,
		"scoop":  true,
		"winget": true,
	}

	for _, manager := range r.managers {
		if manager.SupportsOS(os) && systemManagerNames[manager.Name()] {
			if sysManager, ok := manager.(SystemPackageManager); ok {
				return sysManager
			}
		}
	}

	return nil
}

// FindByName finds a package manager by its name
func (r *PackageManagerRegistry) FindByName(name string) PackageManager {
	for _, manager := range r.managers {
		if manager.Name() == name {
			return manager
		}
	}
	return nil
}

// FindRuntimePackageManager finds a runtime/version manager (asdf) for the given OS
// Runtime managers are used for installing programming language runtimes with version control
func (r *PackageManagerRegistry) FindRuntimePackageManager(os string) RuntimeManager {
	runtimeManagerNames := map[string]bool{
		"asdf": true,
	}

	for _, manager := range r.managers {
		if manager.SupportsOS(os) && runtimeManagerNames[manager.Name()] {
			if runtimeManager, ok := manager.(RuntimeManager); ok {
				return runtimeManager
			}
		}
	}

	return nil
}
