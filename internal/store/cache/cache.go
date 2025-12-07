package cache

import (
	"errors"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type Cache[T any] struct {
	itemsMap map[string]T
}

func New[T any]() *Cache[T] {
	return &Cache[T]{
		itemsMap: map[string]T{},
	}
}

func (c *Cache[T]) Set(key string, value T) {
	c.itemsMap[key] = value
}

func (c *Cache[T]) Get(key string) (T, bool) {
	value, exists := c.itemsMap[key]
	if !exists {
		return value, false
	}

	return value, true
}

func (c *Cache[T]) Remove(key string) {
	delete(c.itemsMap, key)
}
