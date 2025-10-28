package setup

import "time"

type SetupConfig struct {
	Version     int          `yaml:"version" json:"version"`
	Name        string       `yaml:"name,omitempty" json:"name,omitempty"`
	Description string       `yaml:"description,omitempty" json:"description,omitempty"`
	Matrix      *Matrix      `yaml:"matrix,omitempty" json:"matrix,omitempty"`
	Defaults    *Defaults    `yaml:"defaults,omitempty" json:"defaults,omitempty"`
	Env         *Environment `yaml:"env,omitempty" json:"env,omitempty"`
	Bootstrap   []Step       `yaml:"bootstrap,omitempty" json:"bootstrap,omitempty"`
	Provision   []Step       `yaml:"provision,omitempty" json:"provision,omitempty"`
	Setup       []Step       `yaml:"setup,omitempty" json:"setup,omitempty"`
	Verify      []Step       `yaml:"verify,omitempty" json:"verify,omitempty"`
	Post        []Step       `yaml:"post,omitempty" json:"post,omitempty"`
}

type Matrix struct {
	OS   []string `yaml:"os,omitempty" json:"os,omitempty"`
	Arch []string `yaml:"arch,omitempty" json:"arch,omitempty"`
}

type Defaults struct {
	Timeout         string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Shell           string `yaml:"shell,omitempty" json:"shell,omitempty"`
	ContinueOnError bool   `yaml:"continue_on_error,omitempty" json:"continue_on_error,omitempty"`
	CWD             string `yaml:"cwd,omitempty" json:"cwd,omitempty"`
}

type Environment struct {
	Vars    map[string]string `yaml:",inline" json:",inline"`
	Secrets []string          `yaml:"secrets,omitempty" json:"secrets,omitempty"`
}

type Step struct {
	Name            string            `yaml:"name,omitempty" json:"name,omitempty"`
	If              string            `yaml:"if,omitempty" json:"if,omitempty"`
	Timeout         string            `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	CWD             string            `yaml:"cwd,omitempty" json:"cwd,omitempty"`
	Env             map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	EnvFromSecrets  []string          `yaml:"env_from_secrets,omitempty" json:"env_from_secrets,omitempty"`
	OptionalSecrets bool              `yaml:"optional_secrets,omitempty" json:"optional_secrets,omitempty"`
	ContinueOnError bool              `yaml:"continue_on_error,omitempty" json:"continue_on_error,omitempty"`
	Retries         *Retries          `yaml:"retries,omitempty" json:"retries,omitempty"`

	Run                  string                `yaml:"run,omitempty" json:"run,omitempty"`
	Print                string                `yaml:"print,omitempty" json:"print,omitempty"`
	EnsurePackageManager *EnsurePackageManager `yaml:"ensure_package_manager,omitempty"`
	EnsureTool           *EnsureTool           `yaml:"ensure_tool,omitempty" json:"ensure_tool,omitempty"`
	EnsureRuntime        *EnsureRuntime        `yaml:"ensure_runtime,omitempty" json:"ensure_runtime,omitempty"`
	EnsureDatabase       *EnsureDatabase       `yaml:"ensure_database,omitempty" json:"ensure_database,omitempty"`
	AssertCommand        string                `yaml:"assert_command,omitempty" json:"assert_command,omitempty"`
	AssertHTTP           *AssertHTTP           `yaml:"assert_http,omitempty" json:"assert_http,omitempty"`
}

type Retries struct {
	Attempts int    `yaml:"attempts" json:"attempts"`
	Backoff  string `yaml:"backoff" json:"backoff"`
}

type EnsurePackageManager struct {
	Type string `yaml:"type" json:"type"`
}

type EnsureTool struct {
	Name    string             `yaml:"name" json:"name"`
	Version string             `yaml:"version,omitempty" json:"version,omitempty"`
	Install *ToolInstallConfig `yaml:"install,omitempty" json:"install,omitempty"`
}

type ToolInstallConfig struct {
	Brew        *BrewInstall  `yaml:"brew,omitempty" json:"brew,omitempty"`
	Apt         *AptInstall   `yaml:"apt,omitempty" json:"apt,omitempty"`
	Choco       *ChocoInstall `yaml:"choco,omitempty" json:"choco,omitempty"`
	FallbackURL string        `yaml:"fallback_url,omitempty" json:"fallback_url,omitempty"`
	Checksum    string        `yaml:"checksum,omitempty" json:"checksum,omitempty"`
}

type BrewInstall struct {
	Formula string `yaml:"formula" json:"formula"`
}

type AptInstall struct {
	Packages []string `yaml:"packages" json:"packages"`
	Update   bool     `yaml:"update,omitempty" json:"update,omitempty"`
}

type ChocoInstall struct {
	Packages []string `yaml:"packages" json:"packages"`
}

type EnsureRuntime struct {
	Name               string                     `yaml:"name" json:"name"`
	Version            string                     `yaml:"version,omitempty" json:"version,omitempty"`
	Manager            *RuntimeManager            `yaml:"manager,omitempty" json:"manager,omitempty"`
	FallbackInstallers []RuntimeFallbackInstaller `yaml:"fallback_installers,omitempty"`
}

type RuntimeManager struct {
	Asdf bool `yaml:"asdf,omitempty" json:"asdf,omitempty"`
}

type RuntimeFallbackInstaller struct {
	Brew  *BrewInstall  `yaml:"brew,omitempty" json:"brew,omitempty"`
	Apt   *AptInstall   `yaml:"apt,omitempty" json:"apt,omitempty"`
	Choco *ChocoInstall `yaml:"choco,omitempty" json:"choco,omitempty"`
}

type EnsureDatabase struct {
	Engine        string   `yaml:"engine" json:"engine"`
	Version       string   `yaml:"version,omitempty" json:"version,omitempty"`
	ServiceName   string   `yaml:"service_name,omitempty" json:"service_name,omitempty"`
	EnsureRunning bool     `yaml:"ensure_running,omitempty" json:"ensure_running,omitempty"`
	CreateDB      []string `yaml:"create_db,omitempty" json:"create_db,omitempty"`
}

type AssertHTTP struct {
	URL          string   `yaml:"url" json:"url"`
	ExpectStatus int      `yaml:"expect_status,omitempty" json:"expect_status,omitempty"`
	Retries      *Retries `yaml:"retries,omitempty" json:"retries,omitempty"`
}

type ParsedDuration struct {
	Duration time.Duration
	Original string
}

type ParsedRetries struct {
	Attempts int
	Backoff  time.Duration
	Original Retries
}
