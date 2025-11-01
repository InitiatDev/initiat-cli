package setup

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/setup/actions"
	"github.com/InitiatDev/initiat-cli/internal/setup/actions/types"
)

type ActionFactory struct{}

func NewActionFactory() *ActionFactory {
	return &ActionFactory{}
}

func (f *ActionFactory) BuildFromStep(step *Step) (actions.Action, error) {
	switch {
	case step.Run != "":
		return actions.NewRunAction(step.Run), nil
	case step.Print != "":
		return actions.NewPrintAction(step.Print), nil
	case step.EnsurePackageManager != nil:
		return actions.NewEnsurePackageManagerAction(step.EnsurePackageManager.Type), nil
	case step.EnsureTool != nil:
		installConfig := f.buildToolInstallConfig(step.EnsureTool.Install)
		return actions.NewEnsureToolAction(step.EnsureTool.Name, step.EnsureTool.Version, installConfig), nil
	case step.EnsureRuntime != nil:
		manager := f.buildRuntimeManager(step.EnsureRuntime.Manager)
		return actions.NewEnsureRuntimeAction(step.EnsureRuntime.Name, step.EnsureRuntime.Version, manager), nil
	case step.EnsureDatabase != nil:
		return actions.NewEnsureDatabaseAction(
			step.EnsureDatabase.Engine,
			step.EnsureDatabase.Version,
			step.EnsureDatabase.ServiceName,
			step.EnsureDatabase.EnsureRunning,
			step.EnsureDatabase.CreateDB,
		), nil
	case step.AssertCommand != "":
		return actions.NewAssertCommandAction(step.AssertCommand), nil
	case step.AssertHTTP != nil:
		retries, err := f.buildHTTPRetries(step.AssertHTTP.Retries)
		if err != nil {
			return nil, err
		}
		return actions.NewAssertHTTPAction(step.AssertHTTP.URL, step.AssertHTTP.ExpectStatus, retries), nil
	default:
		return nil, fmt.Errorf("no action found in step")
	}
}

func (f *ActionFactory) buildToolInstallConfig(install *ToolInstallConfig) *actions.ToolInstallConfig {
	if install == nil {
		return nil
	}

	return &actions.ToolInstallConfig{
		Brew:        f.convertBrewInstall(install.Brew),
		Apt:         f.convertAptInstall(install.Apt),
		Choco:       f.convertChocoInstall(install.Choco),
		FallbackURL: install.FallbackURL,
		Checksum:    install.Checksum,
	}
}

func (f *ActionFactory) buildRuntimeManager(manager *RuntimeManager) *actions.RuntimeManager {
	if manager == nil {
		return nil
	}

	return &actions.RuntimeManager{
		Asdf: manager.Asdf,
	}
}

func (f *ActionFactory) buildHTTPRetries(retries *Retries) (*types.Retries, error) {
	if retries == nil {
		return nil, nil
	}

	backoffDuration, err := ParseDuration(retries.Backoff)
	if err != nil {
		return nil, fmt.Errorf("invalid retry backoff format: %w", err)
	}

	return &types.Retries{
		Attempts: retries.Attempts,
		Backoff:  int(backoffDuration.Seconds()),
	}, nil
}

func (f *ActionFactory) convertBrewInstall(brew *BrewInstall) *actions.BrewInstall {
	if brew == nil {
		return nil
	}
	return &actions.BrewInstall{Formula: brew.Formula}
}

func (f *ActionFactory) convertAptInstall(apt *AptInstall) *actions.AptInstall {
	if apt == nil {
		return nil
	}
	return &actions.AptInstall{
		Packages: apt.Packages,
		Update:   apt.Update,
	}
}

func (f *ActionFactory) convertChocoInstall(choco *ChocoInstall) *actions.ChocoInstall {
	if choco == nil {
		return nil
	}
	return &actions.ChocoInstall{Packages: choco.Packages}
}
