package cache

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
	value, ok := c.itemsMap[key]

	return value, ok
}

func (c *Cache[T]) Remove(key string) {
	delete(c.itemsMap, key)
}
