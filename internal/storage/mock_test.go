package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockStorage_Get(t *testing.T) {
	storage := NewMockStorage()
	ctx := context.Background()

	val, err := storage.Get(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), val)

	storage.data["test"] = 5
	val, err = storage.Get(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), val)
}

func TestMockStorage_Increment(t *testing.T) {
	storage := NewMockStorage()
	ctx := context.Background()

	val, err := storage.Increment(ctx, "test", time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = storage.Increment(ctx, "test", time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)
}

func TestMockStorage_Set(t *testing.T) {
	storage := NewMockStorage()
	ctx := context.Background()

	err := storage.Set(ctx, "test", 10, time.Minute)
	assert.NoError(t, err)

	val, err := storage.Get(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), val)
}

func TestMockStorage_TTL(t *testing.T) {
	storage := NewMockStorage()
	ctx := context.Background()

	ttl, err := storage.TTL(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), ttl)

	err = storage.Set(ctx, "test", 1, time.Minute)
	assert.NoError(t, err)

	ttl, err = storage.TTL(ctx, "test")
	assert.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= time.Minute)
}

func TestMockStorage_Expiration(t *testing.T) {
	storage := NewMockStorage()
	ctx := context.Background()

	err := storage.Set(ctx, "test", 1, time.Millisecond)
	require.NoError(t, err)

	time.Sleep(2 * time.Millisecond)

	val, err := storage.Get(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), val)
}
