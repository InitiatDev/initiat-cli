package service_managers

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// BrewServiceManager implements service management via Homebrew services
type BrewServiceManager struct{}

func (s *BrewServiceManager) Name() string { return "brew-services" }
func (s *BrewServiceManager) SupportsOS(os string) bool {
	return os == types.OSMacOS || os == types.OSDarwin
}

func (s *BrewServiceManager) StartServiceCommand(service string) types.Command {
	return types.Command{
		Command:     "brew",
		Args:        []string{"services", "start", service},
		Description: fmt.Sprintf("Start %s service via Homebrew", service),
	}
}

func (s *BrewServiceManager) StopServiceCommand(service string) types.Command {
	return types.Command{
		Command:     "brew",
		Args:        []string{"services", "stop", service},
		Description: fmt.Sprintf("Stop %s service via Homebrew", service),
	}
}

func (s *BrewServiceManager) EnableServiceCommand(service string) types.Command {
	return types.Command{
		Command:     "brew",
		Args:        []string{"services", "start", service},
		Description: fmt.Sprintf("Enable %s service via Homebrew", service),
	}
}
