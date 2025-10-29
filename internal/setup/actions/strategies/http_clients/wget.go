package http_clients

import (
	"fmt"
	"os/exec"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// WgetHTTPClient implements HTTP checking via wget
type WgetHTTPClient struct{}

func (c *WgetHTTPClient) Name() string { return "wget" }
func (c *WgetHTTPClient) SupportsOS(os string) bool {
	return os == types.OSLinux
}

func (c *WgetHTTPClient) IsAvailable() bool {
	_, err := exec.LookPath("wget")
	return err == nil
}

func (c *WgetHTTPClient) CheckURLCommand(url string, expectedStatus int, retries *types.Retries) types.Command {
	args := []string{
		"--spider",
		"--timeout=30",
		"--tries=1",
	}

	if retries != nil && retries.Attempts > 1 {
		args = append(args, "--tries", fmt.Sprintf("%d", retries.Attempts))
		if retries.Backoff > 0 {
			args = append(args, "--waitretry", fmt.Sprintf("%d", retries.Backoff))
		}
	}

	args = append(args, url)

	return types.Command{
		Command:     "wget",
		Args:        args,
		Description: fmt.Sprintf("Check URL %s returns status %d", url, expectedStatus),
	}
}
