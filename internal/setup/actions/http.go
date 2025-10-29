package actions

import (
	"fmt"
	"os/exec"
	"strings"
)

type AssertHTTPAction struct {
	*BaseAction
	url          string
	expectStatus int
	retries      *Retries
}

type Retries struct {
	Attempts int    `yaml:"attempts,omitempty" json:"attempts,omitempty"`
	Backoff  string `yaml:"backoff,omitempty" json:"backoff,omitempty"`
	MaxDelay string `yaml:"max_delay,omitempty" json:"max_delay,omitempty"`
}

func NewAssertHTTPAction(url string, expectStatus int, retries *Retries) *AssertHTTPAction {
	return &AssertHTTPAction{
		BaseAction:   NewBaseAction(ActionTypeAssertHTTP),
		url:          url,
		expectStatus: expectStatus,
		retries:      retries,
	}
}

func (a *AssertHTTPAction) Render(ctx *ActionContext) ([]Command, error) {
	if strings.TrimSpace(a.url) == "" {
		return nil, NewActionError(ActionTypeAssertHTTP, "URL cannot be empty", nil)
	}

	commands, err := a.getHTTPCommands(ctx)
	if err != nil {
		return nil, NewActionError(ActionTypeAssertHTTP, "failed to generate HTTP commands", err)
	}

	var result []Command
	for _, cmd := range commands {
		result = append(result, Command{
			Command:     cmd.Command,
			Args:        cmd.Args,
			Env:         ctx.Env,
			WorkingDir:  ctx.WorkingDir,
			Timeout:     ctx.Timeout,
			Description: cmd.Description,
		})
	}

	return result, nil
}

func (a *AssertHTTPAction) Validate() error {
	if strings.TrimSpace(a.url) == "" {
		return NewActionError(ActionTypeAssertHTTP, "URL cannot be empty", nil)
	}

	// Validate URL format (basic check)
	if !strings.HasPrefix(a.url, "http://") && !strings.HasPrefix(a.url, "https://") {
		return NewActionError(ActionTypeAssertHTTP, "URL must start with http:// or https://", nil)
	}

	// Set default expected status if not provided
	if a.expectStatus == 0 {
		a.expectStatus = 200
	}

	// Validate expected status code
	if a.expectStatus < 100 || a.expectStatus > 599 {
		return NewActionError(ActionTypeAssertHTTP, "expected status code must be between 100 and 599", nil)
	}

	return nil
}

type HTTPCommand struct {
	Command     string
	Args        []string
	Description string
}

// getHTTPCommands generates HTTP assertion commands
func (a *AssertHTTPAction) getHTTPCommands(ctx *ActionContext) ([]HTTPCommand, error) {
	os := strings.ToLower(ctx.OS)
	var commands []HTTPCommand

	// Choose HTTP client based on OS
	var httpClient string
	var httpArgs []string

	switch os {
	case OSMacOS, OSDarwin, OSLinux:
		// Prefer curl, fall back to wget
		switch {
		case a.isCommandAvailable("curl"):
			httpClient = "curl"
			httpArgs = a.getCurlArgs()
		case a.isCommandAvailable("wget"):
			httpClient = "wget"
			httpArgs = a.getWgetArgs()
		default:
			return nil, fmt.Errorf("neither curl nor wget is available")
		}
	case OSWindows:
		// Use PowerShell Invoke-WebRequest
		httpClient = "powershell"
		httpArgs = a.getPowerShellArgs()
	default:
		return nil, fmt.Errorf("unsupported OS for HTTP requests: %s", os)
	}

	// Add retry logic if configured
	if a.retries != nil && a.retries.Attempts > 1 {
		commands = append(commands, a.getRetryCommands(httpClient, httpArgs)...)
	} else {
		commands = append(commands, HTTPCommand{
			Command:     httpClient,
			Args:        httpArgs,
			Description: fmt.Sprintf("Check HTTP response from %s", a.url),
		})
	}

	return commands, nil
}

// getCurlArgs generates curl command arguments
func (a *AssertHTTPAction) getCurlArgs() []string {
	args := []string{
		"-s",              // Silent mode
		"-o", "/dev/null", // Discard output
		"-w", "%{http_code}", // Write HTTP status code to stdout
		"--max-time", "30", // 30 second timeout
		"--connect-timeout", "10", // 10 second connection timeout
	}

	// Add retry options if configured
	if a.retries != nil && a.retries.Attempts > 1 {
		args = append(args, "--retry", fmt.Sprintf("%d", a.retries.Attempts-1))
		args = append(args, "--retry-delay", "1")
		args = append(args, "--retry-max-time", "60")
	}

	args = append(args, a.url)
	return args
}

// getWgetArgs generates wget command arguments
func (a *AssertHTTPAction) getWgetArgs() []string {
	args := []string{
		"--spider",     // Don't download, just check
		"--quiet",      // Quiet mode
		"--timeout=30", // 30 second timeout
		"--tries=1",    // Single attempt (retry logic handled separately)
	}

	// Add retry options if configured
	if a.retries != nil && a.retries.Attempts > 1 {
		args = append(args, "--tries", fmt.Sprintf("%d", a.retries.Attempts))
	}

	args = append(args, a.url)
	return args
}

// getPowerShellArgs generates PowerShell Invoke-WebRequest arguments
func (a *AssertHTTPAction) getPowerShellArgs() []string {
	script := fmt.Sprintf(`
		try {
			$response = Invoke-WebRequest -Uri "%s" -TimeoutSec 30 -UseBasicParsing
			Write-Output $response.StatusCode
		} catch {
			Write-Output $_.Exception.Response.StatusCode.value__
		}
	`, a.url)

	return []string{"-Command", script}
}

// getRetryCommands generates retry logic commands
func (a *AssertHTTPAction) getRetryCommands(httpClient string, httpArgs []string) []HTTPCommand {
	var commands []HTTPCommand

	// Create a retry script
	retryScript := a.createRetryScript(httpClient, httpArgs)

	commands = append(commands, HTTPCommand{
		Command:     "bash",
		Args:        []string{"-c", retryScript},
		Description: fmt.Sprintf("Check HTTP response from %s with retry logic", a.url),
	})

	return commands
}

// createRetryScript creates a bash script for retry logic
func (a *AssertHTTPAction) createRetryScript(_ string, _ []string) string {
	attempts := 1
	backoff := 1
	maxDelay := 30

	if a.retries != nil {
		attempts = a.retries.Attempts
		if a.retries.Backoff != "" {
			// Parse backoff (e.g., "2s", "1m")
			// For simplicity, assume seconds
			backoff = 2
		}
		if a.retries.MaxDelay != "" {
			// Parse max delay
			maxDelay = 30
		}
	}

	script := fmt.Sprintf(`
		#!/bin/bash
		set -e
		
		URL="%s"
		EXPECTED_STATUS=%d
		ATTEMPTS=%d
		BACKOFF=%d
		MAX_DELAY=%d
		
		for i in $(seq 1 $ATTEMPTS); do
			echo "Attempt $i of $ATTEMPTS..."
			
			if [ "$httpClient" = "curl" ]; then
				STATUS_CODE=$(curl -s -o /dev/null -w "%%{http_code}" --max-time 30 --connect-timeout 10 "$URL")
			elif [ "$httpClient" = "wget" ]; then
				STATUS_CODE=$(wget --spider --quiet --timeout=30 "$URL" 2>&1 | grep -o '[0-9]\\{3\\}' | tail -1)
			fi
			
			if [ "$STATUS_CODE" = "$EXPECTED_STATUS" ]; then
				echo "HTTP check passed: $STATUS_CODE"
				exit 0
			fi
			
			echo "HTTP check failed: got $STATUS_CODE, expected $EXPECTED_STATUS"
			
			if [ $i -lt $ATTEMPTS ]; then
				DELAY=$((BACKOFF * i))
				if [ $DELAY -gt $MAX_DELAY ]; then
					DELAY=$MAX_DELAY
				fi
				echo "Waiting $DELAY seconds before retry..."
				sleep $DELAY
			fi
		done
		
		echo "HTTP check failed after $ATTEMPTS attempts"
		exit 1
	`, a.url, a.expectStatus, attempts, backoff, maxDelay)

	return script
}

// isCommandAvailable checks if a command is available in PATH
func (a *AssertHTTPAction) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// CheckHTTP performs the actual HTTP check
func (a *AssertHTTPAction) CheckHTTP(ctx *ActionContext) (bool, error) {
	os := strings.ToLower(ctx.OS)
	var cmd *exec.Cmd

	switch os {
	case OSMacOS, OSDarwin, OSLinux:
		switch {
		case a.isCommandAvailable("curl"):
			cmd = exec.Command("curl", a.getCurlArgs()...) // #nosec G204
		case a.isCommandAvailable("wget"):
			cmd = exec.Command("wget", a.getWgetArgs()...) // #nosec G204
		default:
			return false, fmt.Errorf("neither curl nor wget is available")
		}
	case OSWindows:
		cmd = exec.Command("powershell", a.getPowerShellArgs()...) // #nosec G204
	default:
		return false, fmt.Errorf("unsupported OS for HTTP requests: %s", os)
	}

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	statusCode := strings.TrimSpace(string(output))
	expectedStatus := fmt.Sprintf("%d", a.expectStatus)

	return statusCode == expectedStatus, nil
}
