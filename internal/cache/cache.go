package cache

import (
	"L0WB/internal/models"
	"container/list"
	"sync"
)

type Cache interface {
	Get(key string) (*models.Order, bool)
	Add(key string, value *models.Order)
	Remove(key string)
	LoadAll(orders []*models.Order)
}

type LRUCache struct {
	cache   map[string]*list.Element
	list    *list.List
	mu      sync.RWMutex
	maxSize int
}

type cacheEntry struct {
	key   string
	value *models.Order
}

func NewLRUCache(maxSize int) Cache {
	return &LRUCache{
		cache:   make(map[string]*list.Element),
		list:    list.New(),
		maxSize: maxSize,
	}
}

func (c *LRUCache) Get(key string) (*models.Order, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	c.list.MoveToFront(element)
	return element.Value.(*cacheEntry).value, true
}

func (c *LRUCache) Add(key string, value *models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.cache[key]; ok {
		element.Value.(*cacheEntry).value = value
		c.list.MoveToFront(element)
		return
	}

	if c.list.Len() >= c.maxSize {
		element := c.list.Back()
		if element != nil {
			entry := element.Value.(*cacheEntry)
			delete(c.cache, entry.key)
			c.list.Remove(element)
		}
	}

	entry := &cacheEntry{key: key, value: value}
	element := c.list.PushFront(entry)
	c.cache[key] = element
}

func (c *LRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.cache[key]
	if ok {
		c.list.Remove(element)
		delete(c.cache, key)
	}
}

func (c *LRUCache) LoadAll(orders []*models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		if c.list.Len() >= c.maxSize {
			element := c.list.Back()
			if element != nil {
				entry := element.Value.(*cacheEntry)
				delete(c.cache, entry.key)
				c.list.Remove(element)
			}
		}
		entry := &cacheEntry{key: order.OrderUID, value: order}
		element := c.list.PushFront(entry)
		c.cache[order.OrderUID] = element
	}
}
