package setup

import "github.com/InitiatDev/initiat-cli/internal/types"

type StorageInterface interface {
	HasDeviceID() bool
	GetEncryptionPrivateKey() ([]byte, error)
}

type APIClientInterface interface {
	GetSecret(orgSlug, projectSlug, secretKey string) (*types.SecretWithValue, error)
	GetWrappedProjectKey(orgSlug, projectSlug string) (string, error)
}
