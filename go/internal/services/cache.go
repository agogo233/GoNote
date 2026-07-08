package services

import (
	"container/list"
	"strings"
	"sync"
	"time"
)

// Default values for Cache
const (
	DefaultCapacity        = 1000
	DefaultTTL             = 15 * time.Second
	DefaultCleanupInterval = 30 * time.Second
)

// Cache is a thread-safe LRU cache with TTL support
type Cache struct {
	mu              sync.RWMutex
	items           map[string]*list.Element
	evictList       *list.List
	capacity        int
	ttl             time.Duration
	cleanupInterval time.Duration
	stopMu          sync.Mutex      // Protects stopCleanup channel operations
	stopCleanup     chan struct{}
	cleanupDone     chan struct{}   // Signaled when cleanup goroutine exits
	cleanupRunning  bool            // Track if cleanup goroutine is running
}

// cacheEntry holds the cached value and expiration time
type cacheEntry struct {
	key       string
	value     interface{}
	expiresAt time.Time
}

// NewCache creates a new Cache with the specified capacity and TTL.
// If capacity is <= 0, DefaultCapacity (1000) is used.
// If ttl is <= 0, DefaultTTL (15 seconds) is used.
// The cleanup goroutine is NOT started by default - call StartCleanup() to enable it.
func NewCache(capacity int, ttl time.Duration) *Cache {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	return &Cache{
		items:           make(map[string]*list.Element),
		evictList:       list.New(),
		capacity:        capacity,
		ttl:             ttl,
		cleanupInterval: DefaultCleanupInterval,
		stopCleanup:     make(chan struct{}),
	}
}

// Set stores a value in the cache with the specified key.
// If the cache is at capacity, the least recently used item is evicted.
// Thread-safe for concurrent use.
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key already exists
	if elem, ok := c.items[key]; ok {
		// Update existing entry
		entry := elem.Value.(*cacheEntry)
		entry.value = value
		entry.expiresAt = time.Now().Add(c.ttl)
		// Move to back (most recently used)
		c.evictList.MoveToBack(elem)
		return
	}

	// Add new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	elem := c.evictList.PushBack(entry)
	c.items[key] = elem

	// Evict if over capacity
	if c.evictList.Len() > c.capacity {
		c.evictOldest()
	}
}

// Get retrieves a value from the cache.
// Returns (value, true) if found and not expired, (nil, false) otherwise.
// Thread-safe for concurrent use.
// Uses RLock for hot path to avoid serializing concurrent reads (I-2).
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	elem, ok := c.items[key]
	if !ok {
		c.mu.RUnlock()
		return nil, false
	}
	entry := elem.Value.(*cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.mu.RUnlock()
		// Remove expired entry under write lock
		c.mu.Lock()
		if elem, ok = c.items[key]; ok {
			entry = elem.Value.(*cacheEntry)
			if time.Now().After(entry.expiresAt) {
				c.evictList.Remove(elem)
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
		return nil, false
	}
	val := entry.value
	c.mu.RUnlock()
	return val, true
}

// Delete removes an item from the cache.
// Thread-safe for concurrent use.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return
	}

	c.evictList.Remove(elem)
	delete(c.items, key)
}

// DeleteByPrefix removes all items with keys matching the given prefix.
// This enables fine-grained cache invalidation without clearing the entire cache.
// Thread-safe for concurrent use.
func (c *Cache) DeleteByPrefix(prefix string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	deleted := 0
	for key, elem := range c.items {
		if strings.HasPrefix(key, prefix) {
			c.evictList.Remove(elem)
			delete(c.items, key)
			deleted++
		}
	}
	return deleted
}

// HasPrefix checks if any key in the cache starts with the given prefix.
// Thread-safe for concurrent use.
func (c *Cache) HasPrefix(prefix string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for key := range c.items {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// Clear removes all items from the cache.
// Thread-safe for concurrent use.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.evictList.Init()
}

// Len returns the number of items in the cache (including expired ones).
// Thread-safe for concurrent use.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// LenValid returns the number of non-expired items in the cache.
// Thread-safe for concurrent use.
func (c *Cache) LenValid() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	count := 0
	for _, elem := range c.items {
		entry := elem.Value.(*cacheEntry)
		if now.Before(entry.expiresAt) {
			count++
		}
	}
	return count
}

// evictOldest removes the least recently used item from the cache.
// Must be called with the lock held.
func (c *Cache) evictOldest() {
	elem := c.evictList.Front()
	if elem != nil {
		c.evictList.Remove(elem)
		entry := elem.Value.(*cacheEntry)
		delete(c.items, entry.key)
	}
}

// StartCleanup starts a background goroutine that periodically removes expired entries.
// The cleanup runs every cleanupInterval (default 30 seconds).
// Safe to call multiple times - only one cleanup goroutine will run.
func (c *Cache) StartCleanup() {
	c.stopMu.Lock()
	if c.cleanupRunning {
		c.stopMu.Unlock()
		return // Already running
	}
	c.stopCleanup = make(chan struct{})
	c.cleanupDone = make(chan struct{})
	c.cleanupRunning = true
	c.stopMu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic - cleanup will stop but cache remains functional.
				// This prevents a panic in cleanup from crashing the server.
			}
		}()
		ticker := time.NewTicker(c.cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-c.stopCleanup:
				// Signal that we're done
				close(c.cleanupDone)
				return
			case <-ticker.C:
				c.cleanup()
			}
		}
	}()
}

// StopCleanup stops the background cleanup goroutine.
// Call this when shutting down the application.
// Safe to call multiple times.
// After stopping, you can call StartCleanup() again to restart.
func (c *Cache) StopCleanup() {
	c.stopMu.Lock()
	if !c.cleanupRunning {
		c.stopMu.Unlock()
		return // Not running, nothing to do
	}

	// Close stop channel to signal goroutine to exit
	close(c.stopCleanup)
	cleanupDone := c.cleanupDone
	c.stopMu.Unlock()

	// Wait for goroutine to exit (with timeout to prevent deadlock)
	select {
	case <-cleanupDone:
		// Goroutine exited
	case <-time.After(5 * time.Second):
		// Timeout, but we still mark as stopped
	}

	// Mark as stopped after goroutine has exited
	c.stopMu.Lock()
	c.cleanupRunning = false
	c.stopMu.Unlock()
}

// cleanup removes all expired entries from the cache.
// Called periodically by the background goroutine.
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for _, elem := range c.items {
		entry := elem.Value.(*cacheEntry)
		if now.After(entry.expiresAt) {
			c.evictList.Remove(elem)
			delete(c.items, entry.key)
		}
	}
}

// Capacity returns the maximum number of items the cache can hold.
func (c *Cache) Capacity() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.capacity
}

// TTL returns the time-to-live duration for cache entries.
func (c *Cache) TTL() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ttl
}
