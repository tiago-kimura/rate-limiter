package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/tiago-kimura/rate-limiter/internal/storage"
)

type LimitType string

const (
	IPLimit    LimitType = "ip"
	TokenLimit LimitType = "token"
)

type Config struct {
	Limit     int64
	Window    time.Duration
	BlockTime time.Duration
}

type RateLimiter struct {
	storage      storage.Storage
	ipConfig     Config
	tokenConfigs map[string]Config
}

func NewRateLimiter(storage storage.Storage, ipConfig Config) *RateLimiter {
	return &RateLimiter{
		storage:      storage,
		ipConfig:     ipConfig,
		tokenConfigs: make(map[string]Config),
	}
}

func (rl *RateLimiter) SetTokenConfig(token string, config Config) {
	rl.tokenConfigs[token] = config
}

type CheckResult struct {
	Allowed   bool
	Remaining int64
	ResetTime time.Time
	LimitType LimitType
	Limit     int64
}

func (rl *RateLimiter) CheckLimit(ctx context.Context, ip string, token string) (*CheckResult, error) {
	if token != "" {
		if config, exists := rl.tokenConfigs[token]; exists {
			return rl.checkLimitForKey(ctx, fmt.Sprintf("token:%s", token), config, TokenLimit)
		}
	}

	return rl.checkLimitForKey(ctx, fmt.Sprintf("ip:%s", ip), rl.ipConfig, IPLimit)
}

func (rl *RateLimiter) checkLimitForKey(ctx context.Context, key string, config Config, limitType LimitType) (*CheckResult, error) {
	blockedKey := fmt.Sprintf("blocked:%s", key)
	blocked, err := rl.storage.Get(ctx, blockedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check blocked status: %w", err)
	}

	if blocked > 0 {
		ttl, err := rl.storage.TTL(ctx, blockedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get block TTL: %w", err)
		}

		return &CheckResult{
			Allowed:   false,
			Remaining: 0,
			ResetTime: time.Now().Add(ttl),
			LimitType: limitType,
			Limit:     config.Limit,
		}, nil
	}

	count, err := rl.storage.Increment(ctx, key, config.Window)
	if err != nil {
		return nil, fmt.Errorf("failed to increment counter: %w", err)
	}

	if count > config.Limit {
		if err := rl.storage.Set(ctx, blockedKey, 1, config.BlockTime); err != nil {
			return nil, fmt.Errorf("failed to set block: %w", err)
		}

		return &CheckResult{
			Allowed:   false,
			Remaining: 0,
			ResetTime: time.Now().Add(config.BlockTime),
			LimitType: limitType,
			Limit:     config.Limit,
		}, nil
	}

	ttl, err := rl.storage.TTL(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get TTL: %w", err)
	}

	remaining := config.Limit - count
	if remaining < 0 {
		remaining = 0
	}

	return &CheckResult{
		Allowed:   true,
		Remaining: remaining,
		ResetTime: time.Now().Add(ttl),
		LimitType: limitType,
		Limit:     config.Limit,
	}, nil
}
