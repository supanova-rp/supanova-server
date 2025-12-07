package cache

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type Cache[T any] struct {
	keys     []uuid.UUID
	itemsMap map[uuid.UUID]T
}

func New[T any](data T) *Cache[T] {
	return &Cache[T]{
		keys:     []uuid.UUID{},
		itemsMap: map[uuid.UUID]T{},
	}
}

func (c *Cache[T]) Set(key uuid.UUID, value T) {
	if _, exists := c.itemsMap[key]; !exists {
		c.keys = append(c.keys, key)
	}

	c.itemsMap[key] = value
}

func (c *Cache[T]) Get(key uuid.UUID) (T, error) {
	value, exists := c.itemsMap[key]
	if !exists {
		return value, ErrCacheMiss
	}

	return value, nil
}

func (c *Cache[T]) Remove(key uuid.UUID) {
	for i, cKey := range c.keys {
		if cKey == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			break
		}
	}

	delete(c.itemsMap, key)
}
