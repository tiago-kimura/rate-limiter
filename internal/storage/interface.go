package storage

import (
	"context"
	"time"
)

type Storage interface {
	Get(ctx context.Context, key string) (int64, error)
	Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)
	Set(ctx context.Context, key string, count int64, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Close() error
}
