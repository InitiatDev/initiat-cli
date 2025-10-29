package service_managers

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// WindowsServiceManager implements service management via Windows services
type WindowsServiceManager struct{}

func (s *WindowsServiceManager) Name() string              { return "windows-services" }
func (s *WindowsServiceManager) SupportsOS(os string) bool { return os == types.OSWindows }

func (s *WindowsServiceManager) StartServiceCommand(service string) types.Command {
	return types.Command{
		Command:     "net",
		Args:        []string{"start", service},
		Description: fmt.Sprintf("Start %s service via Windows services", service),
	}
}

func (s *WindowsServiceManager) StopServiceCommand(service string) types.Command {
	return types.Command{
		Command:     "net",
		Args:        []string{"stop", service},
		Description: fmt.Sprintf("Stop %s service via Windows services", service),
	}
}

func (s *WindowsServiceManager) EnableServiceCommand(service string) types.Command {
	return types.Command{
		Command:     "sc",
		Args:        []string{"config", service, "start=", "auto"},
		Description: fmt.Sprintf("Enable %s service via Windows services", service),
	}
}
