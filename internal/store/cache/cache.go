package cache

import (
	"errors"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type Cache[T any] struct {
	keys     []string
	itemsMap map[string]T
}

func New[T any]() *Cache[T] {
	return &Cache[T]{
		keys:     []string{},
		itemsMap: map[string]T{},
	}
}

func (c *Cache[T]) Set(key string, value T) {
	if _, exists := c.itemsMap[key]; !exists {
		c.keys = append(c.keys, key)
	}

	c.itemsMap[key] = value
}

func (c *Cache[T]) Get(key string) (T, error) {
	value, exists := c.itemsMap[key]
	if !exists {
		return value, ErrCacheMiss
	}

	return value, nil
}

func (c *Cache[T]) Remove(key string) {
	for i, cKey := range c.keys {
		if cKey == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			break
		}
	}

	delete(c.itemsMap, key)
}
