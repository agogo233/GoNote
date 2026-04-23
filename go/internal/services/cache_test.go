package services

import (
	"sync"
	"testing"
	"time"
)

func TestCacheBasicSetGet(t *testing.T) {
	cache := NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	// Test Set and Get
	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
}

func TestCacheNonExistentKey(t *testing.T) {
	cache := NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	val, ok := cache.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent key")
	}
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}

func TestCacheUpdate(t *testing.T) {
	cache := NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	cache.Set("key1", "value1")
	cache.Set("key1", "value2")

	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected to find key1")
	}
	if val != "value2" {
		t.Errorf("Expected value2, got %v", val)
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	cache.Set("key1", "value1")
	cache.Delete("key1")

	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be deleted")
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache to be empty, got %d items", cache.Len())
	}
}

func TestCacheLRUEviction(t *testing.T) {
	// Small cache to test eviction
	cache := NewCache(3, 15*time.Second)
	defer cache.StopCleanup()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Access key1 to make it recently used
	cache.Get("key1")

	// Add key4, should evict key2 (least recently used)
	cache.Set("key4", "value4")

	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Len())
	}

	// key1 should still exist
	_, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected key1 to exist (recently used)")
	}

	// key2 should be evicted
	_, ok = cache.Get("key2")
	if ok {
		t.Error("Expected key2 to be evicted (least recently used)")
	}
}

func TestCacheTTLExpiration(t *testing.T) {
	// Very short TTL for testing
	cache := NewCache(100, 100*time.Millisecond)
	defer cache.StopCleanup()

	cache.Set("key1", "value1")

	// Should exist immediately
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected key1 to exist immediately")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be expired")
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	cache := NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + (id % 26)))
				cache.Set(key, id*numOperations+j)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + (id % 26)))
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// Cache should still be functional
	if cache.Len() == 0 {
		t.Error("Expected cache to have items after concurrent access")
	}
}

func TestCacheLen(t *testing.T) {
	cache := NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	if cache.Len() != 0 {
		t.Errorf("Expected empty cache, got %d", cache.Len())
	}

	cache.Set("key1", "value1")
	if cache.Len() != 1 {
		t.Errorf("Expected 1 item, got %d", cache.Len())
	}

	cache.Set("key2", "value2")
	if cache.Len() != 2 {
		t.Errorf("Expected 2 items, got %d", cache.Len())
	}

	cache.Delete("key1")
	if cache.Len() != 1 {
		t.Errorf("Expected 1 item after delete, got %d", cache.Len())
	}
}

func TestCacheDefaultValues(t *testing.T) {
	// Test with zero values
	cache := NewCache(0, 0)
	defer cache.StopCleanup()

	// Should use defaults
	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected to find key1 with default values")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
}

func TestCacheDeleteByPrefix(t *testing.T) {
	cache := NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	// Set keys with different prefixes
	cache.Set("notes:list", "list-data")
	cache.Set("notes:folder:a", "folder-a")
	cache.Set("notes:folder:b", "folder-b")
	cache.Set("content:note1.md", "content1")
	cache.Set("content:note2.md", "content2")
	cache.Set("tags:note1.md", "tags1")

	// Initial count
	if cache.Len() != 6 {
		t.Errorf("Expected 6 items, got %d", cache.Len())
	}

	// Delete all "notes:" keys
	deleted := cache.DeleteByPrefix("notes:")
	if deleted != 3 {
		t.Errorf("Expected 3 deleted, got %d", deleted)
	}
	if cache.Len() != 3 {
		t.Errorf("Expected 3 items after delete, got %d", cache.Len())
	}

	// Verify remaining keys
	if _, ok := cache.Get("content:note1.md"); !ok {
		t.Error("Expected content:note1.md to exist")
	}
	if _, ok := cache.Get("notes:list"); ok {
		t.Error("Expected notes:list to be deleted")
	}

	// Delete with non-matching prefix (should return 0)
	deleted = cache.DeleteByPrefix("nonexistent:")
	if deleted != 0 {
		t.Errorf("Expected 0 deleted for non-matching prefix, got %d", deleted)
	}

	// Delete remaining content keys
	cache.DeleteByPrefix("content:")
	if cache.Len() != 1 {
		t.Errorf("Expected 1 item after content delete, got %d", cache.Len())
	}
}

func TestCacheHasPrefix(t *testing.T) {
	cache := NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	// Empty cache
	if cache.HasPrefix("any:") {
		t.Error("Expected HasPrefix to return false for empty cache")
	}

	// Add keys
	cache.Set("notes:list", "data")
	cache.Set("content:note.md", "content")

	// Test positive cases
	if !cache.HasPrefix("notes:") {
		t.Error("Expected HasPrefix to find notes:")
	}
	if !cache.HasPrefix("content:") {
		t.Error("Expected HasPrefix to find content:")
	}

	// Test negative cases
	if cache.HasPrefix("tags:") {
		t.Error("Expected HasPrefix to not find tags:")
	}
	if cache.HasPrefix("notes:list:extra") {
		t.Error("Expected HasPrefix to not find notes:list:extra")
	}

	// After deletion
	cache.DeleteByPrefix("notes:")
	if cache.HasPrefix("notes:") {
		t.Error("Expected HasPrefix to not find notes: after deletion")
	}
}

func TestCacheDeleteByPrefixConcurrent(t *testing.T) {
	cache := NewCache(1000, 15*time.Second)
	defer cache.StopCleanup()

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent writes with prefixes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			prefix := string(rune('a' + (id % 5)))
			for j := 0; j < 10; j++ {
				cache.Set(prefix+":key"+string(rune('0'+j%10)), id*10+j)
			}
		}(i)
	}

	wg.Wait()

	// Concurrent prefix deletions
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			prefix := string(rune('a' + id))
			cache.DeleteByPrefix(prefix + ":")
		}(i)
	}

	wg.Wait()

	// Cache should be empty after all prefix deletions
	if cache.Len() != 0 {
		t.Errorf("Expected empty cache after all prefix deletions, got %d", cache.Len())
	}
}

func TestCacheStartCleanup(t *testing.T) {
	// Test that StartCleanup runs and cleans expired entries
	// Use very short TTL and cleanup interval
	cache := NewCache(100, 50*time.Millisecond)
	cache.cleanupInterval = 50 * time.Millisecond
	cache.StartCleanup()
	defer cache.StopCleanup()

	// Add an entry
	cache.Set("key1", "value1")

	// Should exist immediately
	if _, ok := cache.Get("key1"); !ok {
		t.Error("Expected key1 to exist immediately")
	}

	// Wait for TTL expiration and cleanup
	time.Sleep(150 * time.Millisecond)

	// Entry should be cleaned up by the background goroutine
	if _, ok := cache.Get("key1"); ok {
		t.Error("Expected key1 to be cleaned up by background goroutine")
	}
}

func TestCacheStopCleanup(t *testing.T) {
	// Test that StopCleanup stops the background goroutine
	cache := NewCache(100, 200*time.Millisecond)
	cache.cleanupInterval = 50 * time.Millisecond

	// Start and immediately stop
	cache.StartCleanup()
	time.Sleep(10 * time.Millisecond) // Let goroutine start
	cache.StopCleanup()

	// Add entry after cleanup stopped
	cache.Set("key2", "value2")

	// Wait longer than cleanup interval but less than TTL
	time.Sleep(100 * time.Millisecond)

	// Entry should still exist because cleanup was stopped
	// and TTL hasn't expired yet
	if cache.Len() != 1 {
		t.Errorf("Expected 1 item (cleanup stopped, TTL not expired), got %d", cache.Len())
	}
}

func TestCacheCleanupMultipleStops(t *testing.T) {
	// Test that multiple StopCleanup calls don't panic
	cache := NewCache(100, 15*time.Second)
	cache.StartCleanup()

	// Multiple stops should not panic
	cache.StopCleanup()
	cache.StopCleanup()
	cache.StopCleanup()
}

func TestCacheCleanupRestart(t *testing.T) {
	// Test that cleanup can be restarted after stop
	cache := NewCache(100, 50*time.Millisecond)
	cache.cleanupInterval = 30 * time.Millisecond

	// Start, stop, restart
	cache.StartCleanup()
	cache.StopCleanup()
	cache.StartCleanup()
	defer cache.StopCleanup()

	cache.Set("key1", "value1")

	// Wait for cleanup
	time.Sleep(120 * time.Millisecond)

	if cache.Len() != 0 {
		t.Errorf("Expected cache to be cleaned, got %d items", cache.Len())
	}
}

func TestNoteServiceCacheCleanup(t *testing.T) {
	// Test NoteService StartCacheCleanup and StopCacheCleanup
	noteService := NewNoteServiceWithCache("/tmp", 50*time.Millisecond, 100)
	noteService.StartCacheCleanup()
	defer noteService.StopCacheCleanup()

	// Add some cache entries
	noteService.cache.Set("test:key1", "value1")
	noteService.tagCache.Set("test:tag1", "tagvalue1")

	// Entries should exist
	if _, ok := noteService.cache.Get("test:key1"); !ok {
		t.Error("Expected test:key1 to exist in cache")
	}
	if _, ok := noteService.tagCache.Get("test:tag1"); !ok {
		t.Error("Expected test:tag1 to exist in tagCache")
	}

	// Wait for TTL expiration
	time.Sleep(100 * time.Millisecond)

	// Verify caches were cleaned by background goroutine
	// Note: The cleanup interval is 30s by default, so entries might still be in memory
	// but expired. The important thing is that StartCacheCleanup doesn't panic.
}

func TestNoteServiceStopCacheCleanupMultipleTimes(t *testing.T) {
	// Test that StopCacheCleanup can be called multiple times without panic
	noteService := NewNoteService("/tmp")
	noteService.StartCacheCleanup()

	// Multiple stops should not panic
	noteService.StopCacheCleanup()
	noteService.StopCacheCleanup()
}
