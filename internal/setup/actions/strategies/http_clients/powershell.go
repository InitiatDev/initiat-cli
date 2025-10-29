package http_clients

import (
	"fmt"
	"os/exec"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// PowerShellHTTPClient implements HTTP checking via PowerShell
type PowerShellHTTPClient struct{}

func (c *PowerShellHTTPClient) Name() string { return "powershell" }
func (c *PowerShellHTTPClient) SupportsOS(os string) bool {
	return os == types.OSWindows
}

func (c *PowerShellHTTPClient) IsAvailable() bool {
	_, err := exec.LookPath("powershell")
	return err == nil
}

func (c *PowerShellHTTPClient) CheckURLCommand(url string, expectedStatus int, retries *types.Retries) types.Command {
	script := fmt.Sprintf(`
try {
    $response = Invoke-WebRequest -Uri "%s" -UseBasicParsing -TimeoutSec 30
    Write-Output $response.StatusCode
    exit 0
} catch {
    Write-Output "ERROR"
    exit 1
}`, url)

	return types.Command{
		Command:     "powershell",
		Args:        []string{"-Command", script},
		Description: fmt.Sprintf("Check URL %s returns status %d", url, expectedStatus),
	}
}
