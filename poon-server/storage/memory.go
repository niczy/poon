package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
)

// MemoryBackend implements StorageBackend using in-memory storage
type MemoryBackend struct {
	data map[string][]byte
	mu   sync.RWMutex
}

// NewMemoryBackend creates a new in-memory storage backend
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		data: make(map[string][]byte),
	}
}

// Put stores data at the given key
func (m *MemoryBackend) Put(ctx context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Make a copy of the data to avoid external modifications
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	m.data[key] = dataCopy
	
	return nil
}

// Get retrieves data for the given key
func (m *MemoryBackend) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	data, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	
	// Return a copy to avoid external modifications
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// Exists checks if a key exists
func (m *MemoryBackend) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	_, exists := m.data[key]
	return exists, nil
}

// Delete removes data for the given key
func (m *MemoryBackend) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.data[key]; !exists {
		return fmt.Errorf("key not found: %s", key)
	}
	
	delete(m.data, key)
	return nil
}

// List returns all keys with the given prefix
func (m *MemoryBackend) List(ctx context.Context, prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var keys []string
	for key := range m.data {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	
	return keys, nil
}

// Stream returns a reader for the data (implements io.ReadCloser)
func (m *MemoryBackend) Stream(ctx context.Context, key string) (io.ReadCloser, error) {
	data, err := m.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	
	return &memoryReader{data: data}, nil
}

// Close closes the backend (no-op for memory backend)
func (m *MemoryBackend) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Clear all data
	m.data = make(map[string][]byte)
	return nil
}

// Size returns the number of stored keys
func (m *MemoryBackend) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return len(m.data)
}

// memoryReader implements io.ReadCloser for streaming data
type memoryReader struct {
	data   []byte
	offset int
}

func (r *memoryReader) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func (r *memoryReader) Close() error {
	return nil
}