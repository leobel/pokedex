package pokecache

import (
	"sync"
	"time"
)

type Entry interface {
	Compare(time.Time) int
	GetVal() []byte
}

type PokeEntry struct {
	CreatedAt time.Time
	Val       []byte
}

func (e PokeEntry) Compare(t time.Time) int {
	return e.CreatedAt.Compare(t)
}

func (e PokeEntry) GetVal() []byte {
	return e.Val
}

type Cache interface {
	Get(key string) ([]byte, bool)
	Add(key string, val []byte)
	Stop()
}

type PokeCache struct {
	items map[string]PokeEntry
	done  chan bool
	wg    sync.WaitGroup
	mux   *sync.RWMutex
}

func NewPokeCache(interval time.Duration) *PokeCache {
	cache := &PokeCache{
		items: make(map[string]PokeEntry),
		done:  make(chan bool),
		mux:   &sync.RWMutex{},
	}
	cache.wg.Add(1)
	go cache.reapLoop(interval)
	return cache
}

func (c *PokeCache) Add(key string, val []byte) {
	withWriteLock(c.mux, func() {
		c.items[key] = PokeEntry{time.Now(), val}
	})
}

// tuple wrapper
type GetResult struct {
	entry Entry
	ok    bool
}

func (c *PokeCache) Get(key string) ([]byte, bool) {
	r := withReadLock(c.mux, func() GetResult {
		entry, ok := c.items[key]
		return GetResult{entry, ok}
	})
	if r.ok {
		return r.entry.GetVal(), true
	} else {
		return nil, false
	}
}

func (c *PokeCache) Stop() {
	close(c.done)
	c.wg.Wait()
	withWriteLock(c.mux, func() {
		c.items = map[string]PokeEntry{}
	})
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
	withWriteLock(c.mux, func() {
		for key, entry := range c.items {
			if entry.Compare(t) <= 0 {
				delete(c.items, key)
			}
		}
	})
}

func withReadLock[T any](mux *sync.RWMutex, f func() T) T {
	mux.RLock()
	defer mux.RUnlock()
	return f()
}

func withWriteLock(mux *sync.RWMutex, f func()) {
	mux.Lock()
	defer mux.Unlock()
	f()
}
