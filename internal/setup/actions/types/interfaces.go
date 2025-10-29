package types

// Command represents a shell command to execute
type Command struct {
	Command     string   `json:"command"`
	Args        []string `json:"args"`
	Description string   `json:"description,omitempty"`
}

// Retries represents retry configuration
type Retries struct {
	Attempts int `yaml:"attempts" json:"attempts"`
	Backoff  int `yaml:"backoff" json:"backoff"`
}

// OS constants
const (
	OSLinux   = "linux"
	OSWindows = "windows"
	OSMacOS   = "macos"
	OSDarwin  = "darwin"
)

// PackageManager defines a simple interface for package managers
type PackageManager interface {
	Name() string
	SupportsOS(os string) bool
	InstallCommand(pkg, version string) Command
	CheckInstalledCommand(pkg string) Command
	// GetInstallCommands returns all commands needed for installation (e.g., install + global set for asdf)
	GetInstallCommands(pkg, version string) []Command
}

// ServiceManager defines a simple interface for service managers
type ServiceManager interface {
	Name() string
	SupportsOS(os string) bool
	StartServiceCommand(service string) Command
	StopServiceCommand(service string) Command
	EnableServiceCommand(service string) Command
}

// HTTPClient defines a simple interface for HTTP clients
type HTTPClient interface {
	Name() string
	SupportsOS(os string) bool
	IsAvailable() bool
	CheckURLCommand(url string, expectedStatus int, retries *Retries) Command
}
