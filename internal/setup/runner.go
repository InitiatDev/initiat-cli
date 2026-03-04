package setup

import (
	"errors"
	"fmt"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/config"
	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/storage"
)

var ErrNoCommandsToExecute = errors.New("no commands to execute")

type SetupRunner struct {
	projectCtx *config.ProjectContext
	store      StorageInterface
	apiClient  APIClientInterface
}

func NewSetupRunner(projectCtx *config.ProjectContext) *SetupRunner {
	return NewSetupRunnerWithDeps(projectCtx, storage.New(), client.New())
}

func NewSetupRunnerWithDeps(
	projectCtx *config.ProjectContext,
	store StorageInterface,
	apiClient APIClientInterface,
) *SetupRunner {
	return &SetupRunner{
		projectCtx: projectCtx,
		store:      store,
		apiClient:  apiClient,
	}
}

func (r *SetupRunner) Run(config *SetupConfig) error {
	secrets, err := r.fetchRequiredSecrets(config)
	if err != nil {
		return fmt.Errorf("failed to fetch secrets: %w", err)
	}

	ctx, err := r.createRenderContext(secrets)
	if err != nil {
		return fmt.Errorf("failed to create execution context: %w", err)
	}

	plan, err := Render(config, ctx)
	if err != nil {
		return fmt.Errorf("failed to generate execution plan: %w", err)
	}

	if len(plan.Commands) == 0 {
		return ErrNoCommandsToExecute
	}

	executor := NewExecutor(secrets)
	if err := executor.Execute(plan); err != nil {
		return fmt.Errorf("setup execution failed: %w", err)
	}

	return nil
}

func (r *SetupRunner) fetchRequiredSecrets(config *SetupConfig) (map[string]string, error) {
	secretNames := collectSecretNames(config)
	if len(secretNames) == 0 {
		return nil, nil
	}

	if r.projectCtx == nil {
		return nil, fmt.Errorf(
			"project context required for secrets; run 'initiat project init' or set org/project in .initiat/config.yml")
	}

	if !r.store.HasDeviceID() {
		return nil, fmt.Errorf("device not registered. Please run 'initiat device register <name>' first")
	}

	projectKey, err := r.getProjectKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get project key: %w", err)
	}

	secrets := make(map[string]string)

	for _, secretName := range secretNames {
		secretData, err := r.apiClient.GetSecret(r.projectCtx.OrgSlug, r.projectCtx.ProjectSlug, secretName)
		if err != nil {
			return nil, fmt.Errorf("failed to get secret '%s': %w", secretName, err)
		}

		encryptedValue, err := crypto.Decode(secretData.EncryptedValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decode encrypted value for '%s': %w", secretName, err)
		}

		nonce, err := crypto.Decode(secretData.Nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to decode nonce for '%s': %w", secretName, err)
		}

		decryptedValue, err := crypto.DecryptSecretValue(encryptedValue, nonce, projectKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secret '%s': %w", secretName, err)
		}

		secrets[secretName] = decryptedValue
	}

	return secrets, nil
}

func (r *SetupRunner) getProjectKey() ([]byte, error) {
	wrappedKey, err := r.apiClient.GetWrappedProjectKey(r.projectCtx.OrgSlug, r.projectCtx.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wrapped project key: %w", err)
	}

	devicePrivateKey, err := r.store.GetEncryptionPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get device private key: %w", err)
	}

	projectKey, err := crypto.UnwrapProjectKey(wrappedKey, devicePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap project key: %w", err)
	}

	return projectKey, nil
}

func (r *SetupRunner) createRenderContext(secrets map[string]string) (*RenderContext, error) {
	ctx, err := NewContext()
	if err != nil {
		return nil, err
	}

	shell := detectShell()

	const defaultTimeoutMinutes = 5

	return &RenderContext{
		OS:                     ctx.OS,
		Arch:                   ctx.Arch,
		WorkingDir:             ctx.WorkingDir,
		Shell:                  shell,
		Secrets:                secrets,
		GlobalEnv:              make(map[string]string),
		DefaultTimeout:         time.Minute * defaultTimeoutMinutes,
		DefaultContinueOnError: false,
	}, nil
}
