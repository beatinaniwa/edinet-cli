package cache

import (
	"errors"
	"time"
)

// ErrCacheMiss is returned when the requested key is not in the cache.
var ErrCacheMiss = errors.New("cache miss")

// Cache defines the interface for key-value caching with TTL.
type Cache interface {
	// Get retrieves data for the given key.
	// Returns ErrCacheMiss if the key is not found or has expired.
	// Returns other errors for read failures (corruption, permissions, etc.).
	Get(key string, maxAge time.Duration) ([]byte, error)

	// Set stores data for the given key. Implementations should write atomically.
	Set(key string, data []byte) error
}

// NoopCache is a cache that always misses. Used when caching is disabled.
type NoopCache struct{}

func (NoopCache) Get(_ string, _ time.Duration) ([]byte, error) {
	return nil, ErrCacheMiss
}

func (NoopCache) Set(_ string, _ []byte) error {
	return nil
}
