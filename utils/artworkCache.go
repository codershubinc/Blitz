package utils

import (
	"sync"
)

// ArtworkCache is a tiny in-memory cache for artwork data URIs.
type ArtworkCache struct {
	mu    sync.RWMutex
	cache map[string]string
}

// NewArtworkCache creates a new ArtworkCache
func NewArtworkCache() *ArtworkCache {
	return &ArtworkCache{cache: make(map[string]string)}
}

// Get returns a cached value and whether it was present
func (a *ArtworkCache) Get(key string) (string, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	v, ok := a.cache[key]
	return v, ok
}

// Set stores a value in the cache
func (a *ArtworkCache) Set(key, val string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cache[key] = val
}

// GetOrFetch returns the cached value if present, otherwise calls fetch, stores and returns it.
func (a *ArtworkCache) GetOrFetch(key string, fetch func() (string, error)) (string, error) {
	if v, ok := a.Get(key); ok {
		return v, nil
	}
	val, err := fetch()
	if err != nil {
		return "", err
	}
	a.Set(key, val)
	return val, nil
}
