package cache

import "sync"

type Cache[T any] struct {
	data map[string]T
	rw   sync.RWMutex
}

func New[T any]() *Cache[T] {
	return &Cache[T]{
		data: map[string]T{},
	}
}

func (c *Cache[T]) Set(key string, value T) {
	c.rw.Lock()
	c.data[key] = value
	c.rw.Unlock()
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.rw.RLock()
	value, ok := c.data[key]
	c.rw.RUnlock()

	return value, ok
}

func (c *Cache[T]) Remove(key string) {
	delete(c.data, key)
}
