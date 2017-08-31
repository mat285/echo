package main

import "time"

type cacheEntry struct {
	data      interface{}
	timeStamp time.Time
}

type cache struct {
	entries []*cacheEntry
	lru     []*cacheEntry
	size    int
	max     int
}

func new(max int) *cache {
	return &cache{
		entries: make([]*cacheEntry, 0, max),
		lru:     make([]*cacheEntry, 0, max),
		size:    0,
		max:     max,
	}
}

func (c *cache) put(data interface) {
	if cache.size < cache.max {
		
	}
}
