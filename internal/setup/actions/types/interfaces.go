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

// PackageManager defines base interface for all package managers
type PackageManager interface {
	Name() string
	SupportsOS(os string) bool
	InstallSelfCommand() Command
}

// SystemPackageManager handles system-level tools (brew, apt, choco)
type SystemPackageManager interface {
	PackageManager
	InstallCommand(pkg, version string) Command
	CheckInstalledCommand(pkg string) Command
	ExtractToolInstallCommands(config interface{}) ([]Command, bool)
}

// RuntimeManager handles language runtime installations (asdf)
type RuntimeManager interface {
	PackageManager
	GetInstallCommands(runtime, version string) []Command
	CheckInstalledCommand(runtime string) Command
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

// ToolInstallConfig represents installation configuration for tools
type ToolInstallConfig struct {
	Brew        *BrewInstall  `yaml:"brew,omitempty" json:"brew,omitempty"`
	Apt         *AptInstall   `yaml:"apt,omitempty" json:"apt,omitempty"`
	Choco       *ChocoInstall `yaml:"choco,omitempty" json:"choco,omitempty"`
	FallbackURL string        `yaml:"fallback_url,omitempty" json:"fallback_url,omitempty"`
	Checksum    string        `yaml:"checksum,omitempty" json:"checksum,omitempty"`
}

// BrewInstall represents Homebrew installation configuration
type BrewInstall struct {
	Formula string `yaml:"formula" json:"formula"`
}

// AptInstall represents APT installation configuration
type AptInstall struct {
	Packages []string `yaml:"packages" json:"packages"`
	Update   bool     `yaml:"update,omitempty" json:"update,omitempty"`
}

// ChocoInstall represents Chocolatey installation configuration
type ChocoInstall struct {
	Packages []string `yaml:"packages" json:"packages"`
}
