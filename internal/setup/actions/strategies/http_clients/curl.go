package http_clients

import (
	"fmt"
	"os/exec"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// CurlHTTPClient implements HTTP checking via curl
type CurlHTTPClient struct{}

func (c *CurlHTTPClient) Name() string { return "curl" }
func (c *CurlHTTPClient) SupportsOS(os string) bool {
	return os == types.OSMacOS || os == types.OSDarwin || os == types.OSLinux
}

func (c *CurlHTTPClient) IsAvailable() bool {
	_, err := exec.LookPath("curl")
	return err == nil
}

func (c *CurlHTTPClient) CheckURLCommand(url string, expectedStatus int, retries *types.Retries) types.Command {
	args := []string{
		"-s", "-o", "/dev/null", "-w", "%{http_code}",
		"--max-time", "30",
		"--connect-timeout", "10",
	}

	if retries != nil && retries.Attempts > 1 {
		args = append(args, "--retry", fmt.Sprintf("%d", retries.Attempts-1))
		if retries.Backoff > 0 {
			args = append(args, "--retry-delay", fmt.Sprintf("%d", retries.Backoff))
		}
	}

	args = append(args, url)

	return types.Command{
		Command:     "curl",
		Args:        args,
		Description: fmt.Sprintf("Check URL %s returns status %d", url, expectedStatus),
	}
}
