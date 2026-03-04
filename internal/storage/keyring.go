package storage

import (
	"fmt"
	"sync"

	"github.com/zalando/go-keyring"
)

type Keyring interface {
	Get(service, key string) (string, error)
	Set(service, key, value string) error
	Delete(service, key string) error
}

type systemKeyring struct{}

func (systemKeyring) Get(service, key string) (string, error) {
	return keyring.Get(service, key)
}

func (systemKeyring) Set(service, key, value string) error {
	return keyring.Set(service, key, value)
}

func (systemKeyring) Delete(service, key string) error {
	return keyring.Delete(service, key)
}

type MemKeyring struct {
	mu   sync.Mutex
	data map[string]string
}

func NewMemKeyring() *MemKeyring {
	return &MemKeyring{data: make(map[string]string)}
}

func (m *MemKeyring) key(service, k string) string {
	return service + ":" + k
}

func (m *MemKeyring) Get(service, k string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[m.key(service, k)]
	if !ok {
		return "", fmt.Errorf("not found")
	}
	return v, nil
}

func (m *MemKeyring) Set(service, k, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[m.key(service, k)] = value
	return nil
}

func (m *MemKeyring) Delete(service, k string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, m.key(service, k))
	return nil
}
