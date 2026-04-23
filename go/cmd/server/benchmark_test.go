package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gonote/internal/models"
	"gonote/internal/services"
)

// BenchmarkSearchLatency benchmarks search functionality with different dataset sizes
func BenchmarkSearchLatency(b *testing.B) {
	b.ReportAllocs()
	sizes := []int{100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("notes-%d", size), func(b *testing.B) {
			// Setup: create temporary notes directory
			tempDir, err := os.MkdirTemp("", "gonote-bench-*")
			if err != nil {
				b.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			// Generate test notes
			if err := generateTestNotes(size, tempDir); err != nil {
				b.Fatal(err)
			}

			// Create search service
			searchService := services.NewSearchService(tempDir)

			// Benchmark
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				query := fmt.Sprintf("test%d", i%100)
				_, err := searchService.Search(query)
				if err != nil {
					b.Logf("Search error: %v", err)
				}
			}
		})
	}
}

// BenchmarkIndexedSearchLatency benchmarks indexed search functionality
func BenchmarkIndexedSearchLatency(b *testing.B) {
	b.ReportAllocs()
	sizes := []int{100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("notes-%d", size), func(b *testing.B) {
			// Setup: create temporary notes directory
			tempDir, err := os.MkdirTemp("", "gonote-bench-*")
			if err != nil {
				b.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			// Generate test notes
			if err := generateTestNotes(size, tempDir); err != nil {
				b.Fatal(err)
			}

			// Create and build search index
			noteService := services.NewNoteService(tempDir)
			searchIndex := services.NewSearchIndex(tempDir, noteService)
			if err := searchIndex.BuildIndex(); err != nil {
				b.Fatal(err)
			}

			// Benchmark
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				query := fmt.Sprintf("test%d", i%100)
				_, err := searchIndex.Search(query)
				if err != nil {
					b.Logf("Search error: %v", err)
				}
			}
		})
	}
}

// BenchmarkNoteListLatency benchmarks note listing with different dataset sizes
func BenchmarkNoteListLatency(b *testing.B) {
	b.ReportAllocs()
	sizes := []int{100, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("notes-%d", size), func(b *testing.B) {
			// Setup: create temporary notes directory
			tempDir, err := os.MkdirTemp("", "gonote-bench-*")
			if err != nil {
				b.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			// Generate test notes
			if err := generateTestNotes(size, tempDir); err != nil {
				b.Fatal(err)
			}

			// Create note service
			noteService := services.NewNoteService(tempDir)

			// Benchmark
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, err := noteService.ScanNotes(false)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkCacheOperations benchmarks cache Set and Get operations
func BenchmarkCacheOperations(b *testing.B) {
	b.ReportAllocs()
	cache := services.NewCache(1000, 15*time.Second)
	defer cache.StopCleanup()

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", i)
			cache.Set(key, i)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key%d", i)
			cache.Set(key, i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", i%1000)
			cache.Get(key)
		}
	})
}

// BenchmarkCacheHitRate benchmarks cache hit rate under different access patterns
func BenchmarkCacheHitRate(b *testing.B) {
	b.ReportAllocs()
	cache := services.NewCache(100, 15*time.Second)
	defer cache.StopCleanup()

	// Pre-populate cache with 100 items
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, i)
	}

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key%d", i%100)
			cache.Get(key)
		}
	})

	b.Run("Random", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate random access pattern
			key := fmt.Sprintf("key%d", (i*7)%100)
			cache.Get(key)
		}
	})
}

// BenchmarkPagination benchmarks pagination overhead
func BenchmarkPagination(b *testing.B) {
	b.ReportAllocs()
	// Create a large slice of notes
	notes := make([]models.Note, 10000)
	for i := range notes {
		notes[i] = models.Note{
			Name: fmt.Sprintf("note%d", i),
			Path: fmt.Sprintf("note%d.md", i),
		}
	}

	limits := []int{10, 50, 100, 500}

	for _, limit := range limits {
		b.Run(fmt.Sprintf("limit-%d", limit), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				page := (i % 100) + 1
				services.Paginate(notes, page, limit)
			}
		})
	}
}

// generateTestNotes creates test note files for benchmarking
func generateTestNotes(count int, notesDir string) error {
	for i := 0; i < count; i++ {
		filename := fmt.Sprintf("note%d.md", i)
		filepath := filepath.Join(notesDir, filename)

		content := generateNoteContent(i)
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// generateNoteContent creates realistic markdown content
func generateNoteContent(index int) string {
	return fmt.Sprintf(`# Note %d

This is a test note for benchmarking purposes.

## Section 1

Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.

## Section 2

- Item 1
- Item 2
- Item 3

## Search Terms

test%d keyword%d sample%d data%d

More content here to make the file realistic in size.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.
`, index, index, index, index, index)
}
