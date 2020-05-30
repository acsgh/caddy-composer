package module

import (
	"go.uber.org/zap"
	"time"
)

type Cache struct {
	entries map[string]*CacheEntry
	logger  *zap.Logger
}

type CacheEntry struct {
	source     *WebSource
	validUntil *time.Time
}

func (w *WebComposer) createCache() *Cache {
	cache := new(Cache)
	cache.logger = w.logger
	cache.entries = make(map[string]*CacheEntry)
	return cache
}

func (c Cache) get(id *string) (*WebSource, *time.Time) {
	entry := c.entries[*id]

	if entry != nil {
		if entry.validUntil != nil {
			if entry.validUntil.After(time.Now()) {
				return entry.source, entry.validUntil
			} else {
				return nil, nil
			}
		} else {
			return entry.source, entry.validUntil
		}
	}
	return nil, nil
}

func (c Cache) set(source *WebSource, validUntil *time.Time) {
	entry := new(CacheEntry)
	entry.source = source
	entry.validUntil = validUntil
	c.entries[*source.id] = entry
}
