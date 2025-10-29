package service_managers

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// SystemdServiceManager implements service management via systemctl
type SystemdServiceManager struct{}

func (s *SystemdServiceManager) Name() string              { return "systemd" }
func (s *SystemdServiceManager) SupportsOS(os string) bool { return os == types.OSLinux }

func (s *SystemdServiceManager) StartServiceCommand(service string) types.Command {
	return types.Command{
		Command:     "sudo",
		Args:        []string{"systemctl", "start", service},
		Description: fmt.Sprintf("Start %s service via systemd", service),
	}
}

func (s *SystemdServiceManager) StopServiceCommand(service string) types.Command {
	return types.Command{
		Command:     "sudo",
		Args:        []string{"systemctl", "stop", service},
		Description: fmt.Sprintf("Stop %s service via systemd", service),
	}
}

func (s *SystemdServiceManager) EnableServiceCommand(service string) types.Command {
	return types.Command{
		Command:     "sudo",
		Args:        []string{"systemctl", "enable", service},
		Description: fmt.Sprintf("Enable %s service via systemd", service),
	}
}
