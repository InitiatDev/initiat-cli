package registry

import (
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/strategies/service_managers"
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// ServiceManager defines a simple interface for service managers
type ServiceManager = types.ServiceManager

// ServiceManagerRegistry manages available service manager strategies
type ServiceManagerRegistry struct {
	managers []ServiceManager
}

// NewServiceManagerRegistry creates a new registry with default service managers
func NewServiceManagerRegistry() *ServiceManagerRegistry {
	return &ServiceManagerRegistry{
		managers: []ServiceManager{
			&service_managers.BrewServiceManager{},
			&service_managers.SystemdServiceManager{},
			&service_managers.WindowsServiceManager{},
		},
	}
}

// FindManager finds a suitable service manager for the given OS
func (r *ServiceManagerRegistry) FindManager(os string) ServiceManager {
	for _, manager := range r.managers {
		if manager.SupportsOS(os) {
			return manager
		}
	}
	return nil
}
