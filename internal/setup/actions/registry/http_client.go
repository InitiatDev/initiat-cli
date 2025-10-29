package registry

import (
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/strategies/http_clients"
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

// HTTPClient defines a simple interface for HTTP clients
type HTTPClient = types.HTTPClient

// HTTPClientRegistry manages available HTTP client strategies
type HTTPClientRegistry struct {
	clients []HTTPClient
}

// NewHTTPClientRegistry creates a new registry with default HTTP clients
func NewHTTPClientRegistry() *HTTPClientRegistry {
	return &HTTPClientRegistry{
		clients: []HTTPClient{
			&http_clients.CurlHTTPClient{},
			&http_clients.WgetHTTPClient{},
			&http_clients.PowerShellHTTPClient{},
		},
	}
}

// FindClient finds a suitable HTTP client for the given OS
func (r *HTTPClientRegistry) FindClient(os string) HTTPClient {
	for _, client := range r.clients {
		if client.SupportsOS(os) && client.IsAvailable() {
			return client
		}
	}
	return nil
}
