package env

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/storage"
	"github.com/InitiatDev/initiat-cli/internal/types"
)

func SyncEnvironment(envSlug, orgSlug, projectSlug string) error {
	apiClient := client.New()

	environment, err := apiClient.GetEnvironment(orgSlug, projectSlug, envSlug)
	if err != nil {
		return fmt.Errorf("failed to get environment %s: %w", envSlug, err)
	}

	if len(environment.Secrets) == 0 {
		return WriteSecrets(envSlug, "")
	}

	projectKey, err := getProjectKey(orgSlug, projectSlug)
	if err != nil {
		return fmt.Errorf("failed to get project key: %w", err)
	}

	var envContent strings.Builder
	for _, secret := range environment.Secrets {
		secretWithValue, err := apiClient.GetSecret(orgSlug, projectSlug, secret.Key)
		if err != nil {
			return fmt.Errorf("failed to get secret %s: %w", secret.Key, err)
		}

		encryptedValue, err := crypto.Decode(secretWithValue.EncryptedValue)
		if err != nil {
			return fmt.Errorf("failed to decode encrypted value for %s: %w", secret.Key, err)
		}

		nonce, err := crypto.Decode(secretWithValue.Nonce)
		if err != nil {
			return fmt.Errorf("failed to decode nonce for %s: %w", secret.Key, err)
		}

		decryptedValue, err := crypto.DecryptSecretValue(encryptedValue, nonce, projectKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt secret %s: %w", secret.Key, err)
		}

		envContent.WriteString(fmt.Sprintf("%s=%s\n", secret.Key, decryptedValue))
	}

	return WriteSecrets(envSlug, envContent.String())
}

func SyncAllEnvironments(orgSlug, projectSlug string) error {
	apiClient := client.New()

	environments, err := apiClient.ListEnvironments(orgSlug, projectSlug)
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	var errors []string
	for _, env := range environments {
		if err := SyncEnvironment(env.Slug, orgSlug, projectSlug); err != nil {
			errors = append(errors, fmt.Sprintf("failed to sync %s: %v", env.Slug, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("sync errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

func GetEnvironmentSecrets(envSlug string) ([]types.Secret, error) {
	content, err := ReadSecrets(envSlug)
	if err != nil {
		return nil, err
	}

	if content == "" {
		return []types.Secret{}, nil
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")
	var secrets []types.Secret

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		secrets = append(secrets, types.Secret{
			Key: strings.TrimSpace(parts[0]),
			ID:  i + 1,
		})
	}

	return secrets, nil
}

func GetLastSyncTime(envSlug string) (time.Time, error) {
	secretsPath, err := GetSecretsPath(envSlug)
	if err != nil {
		return time.Time{}, err
	}

	info, err := os.Stat(secretsPath)
	if err != nil {
		return time.Time{}, err
	}

	return info.ModTime(), nil
}

func getProjectKey(orgSlug, projectSlug string) ([]byte, error) {
	apiClient := client.New()
	wrappedKey, err := apiClient.GetWrappedProjectKey(orgSlug, projectSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wrapped project key: %w", err)
	}

	store := storage.New()
	devicePrivateKey, err := store.GetEncryptionPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get device private key: %w", err)
	}

	projectKey, err := crypto.UnwrapProjectKey(wrappedKey, devicePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap project key: %w", err)
	}

	return projectKey, nil
}
