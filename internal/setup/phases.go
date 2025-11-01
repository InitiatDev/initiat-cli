package setup

type Phase struct {
	Name  string
	Steps []Step
}

func GetAllPhases(config *SetupConfig) []Phase {
	return []Phase{
		{"bootstrap", config.Bootstrap},
		{"provision", config.Provision},
		{"setup", config.Setup},
		{"verify", config.Verify},
		{"post", config.Post},
	}
}
