package module

import (
	"go.uber.org/zap"
)

type Cache struct {
	entries map[string]*CacheEntry
	logger  *zap.Logger
}

type CacheEntry struct {
	validUntil int
	value      *string
}
