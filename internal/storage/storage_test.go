package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: These tests use the actual keyring, which may require user interaction
// on some systems. In a production environment, you might want to create
// a mock keyring interface for more reliable testing.

func TestStorage_TokenOperations(t *testing.T) {
	storage := New()
	testToken := "test-token-12345"

	// Clean up any existing token first
	_ = storage.DeleteToken()

	// Initially should not have token
	assert.False(t, storage.HasToken())

	// Store token
	err := storage.StoreToken(testToken)
	if err != nil {
		t.Skipf("Skipping keyring test due to error: %v", err)
		return
	}

	// Should now have token
	assert.True(t, storage.HasToken())

	// Retrieve token
	retrievedToken, err := storage.GetToken()
	assert.NoError(t, err)
	assert.Equal(t, testToken, retrievedToken)

	// Delete token
	err = storage.DeleteToken()
	assert.NoError(t, err)

	// Should no longer have token
	assert.False(t, storage.HasToken())

	// Getting deleted token should fail
	_, err = storage.GetToken()
	assert.Error(t, err)
}

func TestStorage_DeviceIDOperations(t *testing.T) {
	storage := New()
	testDeviceID := "device-abc123"

	// Clean up any existing device ID first
	_ = storage.DeleteDeviceID()

	// Initially should not have device ID
	assert.False(t, storage.HasDeviceID())

	// Store device ID
	err := storage.StoreDeviceID(testDeviceID)
	if err != nil {
		t.Skipf("Skipping keyring test due to error: %v", err)
		return
	}

	// Should now have device ID
	assert.True(t, storage.HasDeviceID())

	// Retrieve device ID
	retrievedDeviceID, err := storage.GetDeviceID()
	assert.NoError(t, err)
	assert.Equal(t, testDeviceID, retrievedDeviceID)

	// Delete device ID
	err = storage.DeleteDeviceID()
	assert.NoError(t, err)

	// Should no longer have device ID
	assert.False(t, storage.HasDeviceID())

	// Getting deleted device ID should fail
	_, err = storage.GetDeviceID()
	assert.Error(t, err)
}

func TestStorage_MultipleOperations(t *testing.T) {
	storage := New()
	testToken := "multi-test-token"
	testDeviceID := "multi-test-device"

	// Store both
	err := storage.StoreToken(testToken)
	if err != nil {
		t.Skipf("Skipping keyring test due to error: %v", err)
		return
	}

	err = storage.StoreDeviceID(testDeviceID)
	if err != nil {
		t.Skipf("Skipping keyring test due to error: %v", err)
		return
	}

	// Both should exist
	assert.True(t, storage.HasToken())
	assert.True(t, storage.HasDeviceID())

	// Retrieve both
	retrievedToken, err := storage.GetToken()
	assert.NoError(t, err)
	assert.Equal(t, testToken, retrievedToken)

	retrievedDeviceID, err := storage.GetDeviceID()
	assert.NoError(t, err)
	assert.Equal(t, testDeviceID, retrievedDeviceID)

	// Clean up
	_ = storage.DeleteToken()
	_ = storage.DeleteDeviceID()
}

func TestStorage_OverwriteValues(t *testing.T) {
	storage := New()

	// Store initial values
	err := storage.StoreToken("initial-token")
	if err != nil {
		t.Skipf("Skipping keyring test due to error: %v", err)
		return
	}

	err = storage.StoreDeviceID("initial-device")
	if err != nil {
		t.Skipf("Skipping keyring test due to error: %v", err)
		return
	}

	// Overwrite with new values
	newToken := "updated-token"
	newDeviceID := "updated-device"

	err = storage.StoreToken(newToken)
	assert.NoError(t, err)

	err = storage.StoreDeviceID(newDeviceID)
	assert.NoError(t, err)

	// Verify new values
	retrievedToken, err := storage.GetToken()
	assert.NoError(t, err)
	assert.Equal(t, newToken, retrievedToken)

	retrievedDeviceID, err := storage.GetDeviceID()
	assert.NoError(t, err)
	assert.Equal(t, newDeviceID, retrievedDeviceID)

	// Clean up
	_ = storage.DeleteToken()
	_ = storage.DeleteDeviceID()
}
