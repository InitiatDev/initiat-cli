package env

import "time"

type LocalEnvironment struct {
	Slug       string
	Path       string
	Synced     time.Time
	HasSecrets bool
}

type State struct {
	Active string
}

type EnvironmentInfo struct {
	Slug         string
	Name         string
	IsActive     bool
	Synced       time.Time
	HasSecrets   bool
	SecretsCount int
}
