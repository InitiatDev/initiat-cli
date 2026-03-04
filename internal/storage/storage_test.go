package storage

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage_TokenOperations(t *testing.T) {
	storage := NewWithKeyring("initiat-cli-test-token", NewMemKeyring())
	testToken := "test-token-12345"

	_ = storage.DeleteToken()

	assert.False(t, storage.HasToken())

	err := storage.StoreToken(testToken)
	assert.NoError(t, err)

	assert.True(t, storage.HasToken())

	retrievedToken, err := storage.GetToken()
	assert.NoError(t, err)
	assert.Equal(t, testToken, retrievedToken)

	err = storage.DeleteToken()
	assert.NoError(t, err)

	assert.False(t, storage.HasToken())

	_, err = storage.GetToken()
	assert.Error(t, err)
}

func TestStorage_DeviceIDOperations(t *testing.T) {
	storage := NewWithKeyring("initiat-cli-test-device", NewMemKeyring())
	testDeviceID := "device-abc123"

	_ = storage.DeleteDeviceID()

	assert.False(t, storage.HasDeviceID())

	err := storage.StoreDeviceID(testDeviceID)
	assert.NoError(t, err)

	assert.True(t, storage.HasDeviceID())

	retrievedDeviceID, err := storage.GetDeviceID()
	assert.NoError(t, err)
	assert.Equal(t, testDeviceID, retrievedDeviceID)

	err = storage.DeleteDeviceID()
	assert.NoError(t, err)

	assert.False(t, storage.HasDeviceID())

	_, err = storage.GetDeviceID()
	assert.Error(t, err)
}

func TestStorage_MultipleOperations(t *testing.T) {
	storage := NewWithKeyring("initiat-cli-test-multi", NewMemKeyring())
	testToken := "multi-test-token"
	testDeviceID := "multi-test-device"

	_ = storage.DeleteToken()
	_ = storage.DeleteDeviceID()

	err := storage.StoreToken(testToken)
	assert.NoError(t, err)

	err = storage.StoreDeviceID(testDeviceID)
	assert.NoError(t, err)

	assert.True(t, storage.HasToken())
	assert.True(t, storage.HasDeviceID())

	retrievedToken, err := storage.GetToken()
	assert.NoError(t, err)
	assert.Equal(t, testToken, retrievedToken)

	retrievedDeviceID, err := storage.GetDeviceID()
	assert.NoError(t, err)
	assert.Equal(t, testDeviceID, retrievedDeviceID)

	_ = storage.DeleteToken()
	_ = storage.DeleteDeviceID()
}

func TestStorage_OverwriteValues(t *testing.T) {
	storage := NewWithKeyring("initiat-cli-test-overwrite", NewMemKeyring())

	err := storage.StoreToken("initial-token")
	assert.NoError(t, err)

	err = storage.StoreDeviceID("initial-device")
	assert.NoError(t, err)

	newToken := "updated-token"
	newDeviceID := "updated-device"

	err = storage.StoreToken(newToken)
	assert.NoError(t, err)

	err = storage.StoreDeviceID(newDeviceID)
	assert.NoError(t, err)

	retrievedToken, err := storage.GetToken()
	assert.NoError(t, err)
	assert.Equal(t, newToken, retrievedToken)

	retrievedDeviceID, err := storage.GetDeviceID()
	assert.NoError(t, err)
	assert.Equal(t, newDeviceID, retrievedDeviceID)

	_ = storage.DeleteToken()
	_ = storage.DeleteDeviceID()
}

func TestStorage_SigningPrivateKeyOperations(t *testing.T) {
	storage := NewWithKeyring("initiat-cli-test-signing", NewMemKeyring())

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate test keypair: %v", err)
	}

	_ = storage.DeleteSigningPrivateKey()

	assert.False(t, storage.HasSigningPrivateKey())

	err = storage.StoreSigningPrivateKey(privateKey)
	assert.NoError(t, err)

	assert.True(t, storage.HasSigningPrivateKey())

	retrievedKey, err := storage.GetSigningPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, privateKey, retrievedKey)

	message := []byte("test message")
	signature := ed25519.Sign(retrievedKey, message)
	assert.True(t, ed25519.Verify(publicKey, message, signature))

	err = storage.DeleteSigningPrivateKey()
	assert.NoError(t, err)

	assert.False(t, storage.HasSigningPrivateKey())

	_, err = storage.GetSigningPrivateKey()
	assert.Error(t, err)
}

func TestStorage_EncryptionPrivateKeyOperations(t *testing.T) {
	storage := NewWithKeyring("initiat-cli-test-encryption", NewMemKeyring())

	testKey := make([]byte, 32)
	_, err := rand.Read(testKey)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	_ = storage.DeleteEncryptionPrivateKey()

	assert.False(t, storage.HasEncryptionPrivateKey())

	err = storage.StoreEncryptionPrivateKey(testKey)
	assert.NoError(t, err)

	assert.True(t, storage.HasEncryptionPrivateKey())

	retrievedKey, err := storage.GetEncryptionPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, testKey, retrievedKey)

	err = storage.DeleteEncryptionPrivateKey()
	assert.NoError(t, err)

	assert.False(t, storage.HasEncryptionPrivateKey())

	_, err = storage.GetEncryptionPrivateKey()
	assert.Error(t, err)
}

func TestURLBasedServiceNameUniqueness(t *testing.T) {
	urls := []string{
		"https://www.initiat.dev",
		"http://localhost:8080",
		"https://staging.initiat.dev",
		"http://192.168.1.100:3000",
	}

	serviceNames := make(map[string]bool)

	for _, url := range urls {
		serviceName := generateServiceNameFromURL(url)
		assert.False(t, serviceNames[serviceName],
			"Service name %q should be unique for URL %q", serviceName, url)
		serviceNames[serviceName] = true
	}
}

func TestStorageNewWithDifferentAPIURLs(t *testing.T) {
	serviceName1 := generateServiceNameFromURL("https://www.initiat.dev")
	serviceName2 := generateServiceNameFromURL("http://localhost:8080")

	assert.NotEqual(t, serviceName1, serviceName2,
		"Different API URLs should generate different service names")

	assert.Contains(t, serviceName1, "initiat-cli")
	assert.Contains(t, serviceName2, "initiat-cli")
}

func TestStorageNewWithExplicitServiceName(t *testing.T) {
	storage := NewWithServiceName("my-custom-service")
	assert.Equal(t, "my-custom-service", storage.serviceName,
		"Explicit service name should be used as-is")
}

func TestStorageNewWithDefaultServiceName(t *testing.T) {
	expectedServiceName := generateServiceNameFromURL("https://www.initiat.dev")
	storage := New()

	assert.Equal(t, expectedServiceName, storage.serviceName,
		"Default service name should trigger URL-based generation")
	assert.Contains(t, storage.serviceName, "initiat-cli",
		"URL-based service name should contain base name")
}

func TestStorage_DeviceNameOperations(t *testing.T) {
	storage := NewWithKeyring("initiat-cli-test-device-name", NewMemKeyring())
	testDeviceName := "my-test-device"

	_ = storage.DeleteDeviceName()

	assert.False(t, storage.HasDeviceName())

	err := storage.StoreDeviceName(testDeviceName)
	assert.NoError(t, err)

	assert.True(t, storage.HasDeviceName())

	retrievedDeviceName, err := storage.GetDeviceName()
	assert.NoError(t, err)
	assert.Equal(t, testDeviceName, retrievedDeviceName)

	err = storage.DeleteDeviceName()
	assert.NoError(t, err)

	assert.False(t, storage.HasDeviceName())

	_, err = storage.GetDeviceName()
	assert.Error(t, err)
}

func TestStorage_ClearDeviceCredentialsIncludesDeviceName(t *testing.T) {
	storage := NewWithKeyring("initiat-cli-test-clear", NewMemKeyring())

	_ = storage.DeleteDeviceID()
	_ = storage.DeleteDeviceName()
	_ = storage.DeleteSigningPrivateKey()
	_ = storage.DeleteEncryptionPrivateKey()

	err := storage.StoreDeviceID("test-device-id")
	assert.NoError(t, err)

	err = storage.StoreDeviceName("test-device-name")
	assert.NoError(t, err)

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate signing key: %v", err)
	}

	err = storage.StoreSigningPrivateKey(privateKey)
	assert.NoError(t, err)

	testKey := make([]byte, 32)
	_, err = rand.Read(testKey)
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	err = storage.StoreEncryptionPrivateKey(testKey)
	assert.NoError(t, err)

	assert.True(t, storage.HasDeviceID())
	assert.True(t, storage.HasDeviceName())
	assert.True(t, storage.HasSigningPrivateKey())
	assert.True(t, storage.HasEncryptionPrivateKey())

	err = storage.ClearDeviceCredentials()
	assert.NoError(t, err)

	assert.False(t, storage.HasDeviceID())
	assert.False(t, storage.HasDeviceName())
	assert.False(t, storage.HasSigningPrivateKey())
	assert.False(t, storage.HasEncryptionPrivateKey())
}
