package ratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiago-kimura/rate-limiter/internal/storage"
)

func TestRateLimiter_IPLimiting(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	config := Config{
		Limit:     3,
		Window:    time.Second,
		BlockTime: time.Minute,
	}

	rateLimiter := NewRateLimiter(mockStorage, config)
	ctx := context.Background()
	ip := "192.168.1.1"

	for i := 0; i < 3; i++ {
		result, err := rateLimiter.CheckLimit(ctx, ip, "")
		require.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, IPLimit, result.LimitType)
		assert.Equal(t, int64(3-i-1), result.Remaining)
	}

	result, err := rateLimiter.CheckLimit(ctx, ip, "")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)
	assert.Equal(t, IPLimit, result.LimitType)

	result, err = rateLimiter.CheckLimit(ctx, ip, "")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
}

func TestRateLimiter_TokenLimiting(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	ipConfig := Config{
		Limit:     3,
		Window:    time.Second,
		BlockTime: time.Minute,
	}

	rateLimiter := NewRateLimiter(mockStorage, ipConfig)

	tokenConfig := Config{
		Limit:     5,
		Window:    time.Second,
		BlockTime: time.Minute,
	}
	rateLimiter.SetTokenConfig("test_token", tokenConfig)

	ctx := context.Background()
	ip := "192.168.1.1"
	token := "test_token"

	for i := 0; i < 5; i++ {
		result, err := rateLimiter.CheckLimit(ctx, ip, token)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, TokenLimit, result.LimitType)
		assert.Equal(t, int64(5-i-1), result.Remaining)
	}

	result, err := rateLimiter.CheckLimit(ctx, ip, token)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)
	assert.Equal(t, TokenLimit, result.LimitType)
}

func TestRateLimiter_TokenPrecedence(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	ipConfig := Config{
		Limit:     2,
		Window:    time.Second,
		BlockTime: time.Minute,
	}

	rateLimiter := NewRateLimiter(mockStorage, ipConfig)

	tokenConfig := Config{
		Limit:     10,
		Window:    time.Second,
		BlockTime: time.Minute,
	}
	rateLimiter.SetTokenConfig("vip_token", tokenConfig)

	ctx := context.Background()
	ip := "192.168.1.1"

	result, err := rateLimiter.CheckLimit(ctx, ip, "")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, IPLimit, result.LimitType)

	result, err = rateLimiter.CheckLimit(ctx, ip, "")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, IPLimit, result.LimitType)

	result, err = rateLimiter.CheckLimit(ctx, ip, "")
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, IPLimit, result.LimitType)

	result, err = rateLimiter.CheckLimit(ctx, "192.168.1.2", "vip_token")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, TokenLimit, result.LimitType)
}

func TestRateLimiter_UnknownToken(t *testing.T) {
	mockStorage := storage.NewMockStorage()
	ipConfig := Config{
		Limit:     3,
		Window:    time.Second,
		BlockTime: time.Minute,
	}

	rateLimiter := NewRateLimiter(mockStorage, ipConfig)
	ctx := context.Background()
	ip := "192.168.1.1"

	result, err := rateLimiter.CheckLimit(ctx, ip, "unknown_token")
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, IPLimit, result.LimitType)
}
