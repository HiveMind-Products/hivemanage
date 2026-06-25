package cache

import (
	"sync"
	"time"
)

const (
	TokenCacheKey = "tokenHash_%s.data"
)

type Item struct {
	Object     any
	Expiration int64
}

type memcache struct {
	items             map[string]Item
	defaultExpiration time.Duration
	mutex             sync.Mutex
}

type Cache struct {
	*memcache
}

type mop struct {
	interval time.Duration
}

func NewMemcache(defaultExpiration time.Duration) *Cache {
	c := &Cache{
		memcache: &memcache{
			items:             make(map[string]Item),
			defaultExpiration: defaultExpiration,
		},
	}

	go (&mop{interval: time.Minute}).start(c)

	return c
}

func (c *Cache) Set(key string, value any, expiration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if expiration <= 0 {
		expiration = c.defaultExpiration
	}

	c.items[key] = Item{
		Object:     value,
		Expiration: time.Now().Add(expiration).Unix(),
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, found := c.items[key]
	if !found || (item.Expiration > 0 && item.Expiration < time.Now().Unix()) {
		return nil, false
	}

	return item.Object, true
}

func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

func (m *mop) start(c *Cache) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now().Unix()
		for key, item := range c.items {
			if item.Expiration > 0 && item.Expiration < now {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}
