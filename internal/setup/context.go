package setup

import (
	"os"
	"path/filepath"
	"runtime"
)

type Context struct {
	OS         string
	Arch       string
	WorkingDir string
	Secrets    map[string]string
}

func NewContext() (*Context, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	osName := normalizeOS(runtime.GOOS)
	arch := normalizeArch(runtime.GOARCH)

	return &Context{
		OS:         osName,
		Arch:       arch,
		WorkingDir: wd,
		Secrets:    make(map[string]string),
	}, nil
}

func (c *Context) WithSecrets(secrets map[string]string) *Context {
	newSecrets := make(map[string]string)
	for k, v := range c.Secrets {
		newSecrets[k] = v
	}
	for k, v := range secrets {
		newSecrets[k] = v
	}

	return &Context{
		OS:         c.OS,
		Arch:       c.Arch,
		WorkingDir: c.WorkingDir,
		Secrets:    newSecrets,
	}
}

func (c *Context) WithWorkingDir(dir string) (*Context, error) {
	if filepath.IsAbs(dir) {
		return &Context{
			OS:         c.OS,
			Arch:       c.Arch,
			WorkingDir: dir,
			Secrets:    c.Secrets,
		}, nil
	}

	absPath, err := filepath.Abs(filepath.Join(c.WorkingDir, dir))
	if err != nil {
		return nil, err
	}

	return &Context{
		OS:         c.OS,
		Arch:       c.Arch,
		WorkingDir: absPath,
		Secrets:    c.Secrets,
	}, nil
}

func (c *Context) ShouldExecuteStep(step *Step) (bool, error) {
	evaluator := NewConditionEvaluator(c.OS, c.Arch, nil)
	return evaluator.ShouldExecuteStep(step)
}

//nolint:unparam // goos can be any runtime.GOOS value, not just "darwin"
func normalizeOS(goos string) string {
	switch goos {
	case goOSDarwin:
		return osMacOS
	case goOSLinux:
		return osLinux
	case goOSWindows:
		return osWindows
	default:
		return goos
	}
}

func normalizeArch(goarch string) string {
	switch goarch {
	case goArchAMD64:
		return archX86_64
	case goArchARM64:
		return archARM64
	default:
		return goarch
	}
}
