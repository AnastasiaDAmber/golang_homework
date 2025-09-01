package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type cacheItem struct {
	key   Key
	value interface{}
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mu       sync.Mutex
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity <= 0 {
		// кэш с нулевой ёмкостостью, то ничего не хранится
		return false
	}

	if it, ok := c.items[key]; ok {
		// обновление значения и передвижение вперёд
		it.Value.(*cacheItem).value = value
		c.queue.MoveToFront(it)
		return true
	}

	// вставка нового
	node := c.queue.PushFront(&cacheItem{key: key, value: value})
	c.items[key] = node

	// если переполнено, то  выселяем хвост
	if c.queue.Len() > c.capacity {
		tail := c.queue.Back()
		if tail != nil {
			c.queue.Remove(tail)
			ci := tail.Value.(*cacheItem)
			delete(c.items, ci.key)
		}
	}

	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	it, ok := c.items[key]
	if !ok {
		return nil, false
	}
	c.queue.MoveToFront(it)
	return it.Value.(*cacheItem).value, true
}

func (c *lruCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// создание новых структур
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}
