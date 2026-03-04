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
	Secrets         []string          `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	OptionalSecrets bool              `yaml:"optional_secrets,omitempty" json:"optional_secrets,omitempty"`
	ContinueOnError bool              `yaml:"continue_on_error,omitempty" json:"continue_on_error,omitempty"`
	Retries         *Retries          `yaml:"retries,omitempty" json:"retries,omitempty"`

	Run   string `yaml:"run,omitempty" json:"run,omitempty"`
	Print string `yaml:"print,omitempty" json:"print,omitempty"`
}

type Retries struct {
	Attempts int    `yaml:"attempts" json:"attempts"`
	Backoff  string `yaml:"backoff" json:"backoff"`
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
