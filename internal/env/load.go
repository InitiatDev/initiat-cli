package env

import (
	"fmt"
	"strings"

	"github.com/InitiatDev/initiat-cli/internal/client"
	"github.com/InitiatDev/initiat-cli/internal/crypto"
	"github.com/InitiatDev/initiat-cli/internal/storage"
)

func LoadEnvironmentSecrets(orgSlug, projectSlug string) (string, error) {
	activeEnv, err := GetActiveEnvironment()
	if err != nil {
		return "", fmt.Errorf("no active environment set")
	}

	apiClient := client.New()
	environment, err := apiClient.GetEnvironment(orgSlug, projectSlug, activeEnv)
	if err != nil {
		return "", fmt.Errorf("failed to get environment %s: %w", activeEnv, err)
	}

	if len(environment.Secrets) == 0 {
		return fmt.Sprintf("export INITIAT_ENV=%s\n", shellEscape(activeEnv)), nil
	}

	projectKey, err := getProjectKey(orgSlug, projectSlug)
	if err != nil {
		return "", fmt.Errorf("failed to get project key: %w", err)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("export INITIAT_ENV=%s\n", shellEscape(activeEnv)))

	for _, secret := range environment.Secrets {
		secretWithValue, err := apiClient.GetSecret(orgSlug, projectSlug, secret.Key)
		if err != nil {
			return "", fmt.Errorf("failed to get secret %s: %w", secret.Key, err)
		}

		encryptedValue, err := crypto.Decode(secretWithValue.EncryptedValue)
		if err != nil {
			return "", fmt.Errorf("failed to decode encrypted value for %s: %w", secret.Key, err)
		}

		nonce, err := crypto.Decode(secretWithValue.Nonce)
		if err != nil {
			return "", fmt.Errorf("failed to decode nonce for %s: %w", secret.Key, err)
		}

		decryptedValue, err := crypto.DecryptSecretValue(encryptedValue, nonce, projectKey)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt secret %s: %w", secret.Key, err)
		}

		output.WriteString(fmt.Sprintf("export %s=%s\n", secret.Key, shellEscape(decryptedValue)))
	}

	return output.String(), nil
}

func shellEscape(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "$", "\\$")
	value = strings.ReplaceAll(value, "`", "\\`")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	return fmt.Sprintf("\"%s\"", value)
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
