package pokecache

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Add(key string, val []byte)
}

type PokeCache struct {
	items map[string]cacheEntry
	done  chan bool
	wg    sync.WaitGroup
	mux   *sync.RWMutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *PokeCache {
	cache := &PokeCache{
		items: make(map[string]cacheEntry),
		done:  make(chan bool),
		mux:   &sync.RWMutex{},
	}
	cache.wg.Add(1)
	go cache.reapLoop(interval)
	return cache
}

func (c *PokeCache) Add(key string, val []byte) {
	c.mux.Lock()
	c.items[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.mux.Unlock()
}

func (c *PokeCache) Get(key string) ([]byte, bool) {
	c.mux.RLock()
	entry, ok := c.items[key]
	c.mux.RUnlock()
	if ok {
		return entry.val, true
	} else {
		return nil, false
	}
}

func (c *PokeCache) Stop() {
	close(c.done)
	c.wg.Wait()
}

func (c *PokeCache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer c.wg.Done()

	for {
		select {
		case <-c.done:
			return
		case t := <-ticker.C:
			threshold := t.Add(-1 * interval)
			c.cleanCache(threshold)
		}
	}

}

func (c *PokeCache) cleanCache(t time.Time) {
	c.mux.RLock()
	for key, entry := range c.items {
		if entry.createdAt.Compare(t) <= 0 {
			delete(c.items, key)
		}
	}
	c.mux.RUnlock()
}
