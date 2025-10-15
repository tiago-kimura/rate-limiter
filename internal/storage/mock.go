package storage

import (
	"context"
	"time"
)

type MockStorage struct {
	data map[string]int64
	ttl  map[string]time.Time
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]int64),
		ttl:  make(map[string]time.Time),
	}
}

func (m *MockStorage) Get(ctx context.Context, key string) (int64, error) {
	if expiry, exists := m.ttl[key]; exists && !time.Now().Before(expiry) {
		delete(m.data, key)
		delete(m.ttl, key)
		return 0, nil
	}

	if val, exists := m.data[key]; exists {
		return val, nil
	}
	return 0, nil
}

func (m *MockStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	if expiry, exists := m.ttl[key]; exists && !time.Now().Before(expiry) {
		delete(m.data, key)
		delete(m.ttl, key)
	}

	val := m.data[key] + 1
	m.data[key] = val

	if _, exists := m.ttl[key]; !exists {
		m.ttl[key] = time.Now().Add(expiration)
	}

	return val, nil
}

func (m *MockStorage) Set(ctx context.Context, key string, count int64, expiration time.Duration) error {
	m.data[key] = count
	m.ttl[key] = time.Now().Add(expiration)
	return nil
}

func (m *MockStorage) TTL(ctx context.Context, key string) (time.Duration, error) {
	if expiry, exists := m.ttl[key]; exists {
		remaining := time.Until(expiry)
		if remaining <= 0 {
			delete(m.data, key)
			delete(m.ttl, key)
			return 0, nil
		}
		return remaining, nil
	}
	return 0, nil
}

func (m *MockStorage) Close() error {
	return nil
}
