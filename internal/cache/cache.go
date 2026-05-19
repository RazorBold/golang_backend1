package cache

import (
	"context"
	"time"
)

// Cache adalah interface untuk cache client.
// Abstraksi ini memungkinkan unit test menggunakan fake/mock
// tanpa perlu koneksi Redis yang sebenarnya.
type Cache interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Ping(ctx context.Context) error
	Close() error
}
