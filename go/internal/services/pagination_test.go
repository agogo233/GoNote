package services

import (
	"testing"

	"gonote/internal/models"
)

func createTestNotes(count int) []models.Note {
	notes := make([]models.Note, count)
	for i := 0; i < count; i++ {
		notes[i] = models.Note{
			Name:     "note",
			Path:    "note.md",
			Folder:  "",
			Modified: "2024-01-01T00:00:00Z",
			Size:     100,
			Type:     "note",
			Tags:     []string{},
		}
	}
	return notes
}

func TestPaginate_NormalPagination(t *testing.T) {
	notes := createTestNotes(25)

	result := Paginate(notes, 1, 10)

	if len(result.Notes) != 10 {
		t.Errorf("expected 10 notes, got %d", len(result.Notes))
	}

	if result.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Pagination.Page)
	}

	if result.Pagination.Limit != 10 {
		t.Errorf("expected limit 10, got %d", result.Pagination.Limit)
	}

	if result.Pagination.Total != 25 {
		t.Errorf("expected total 25, got %d", result.Pagination.Total)
	}

	if result.Pagination.TotalPages != 3 {
		t.Errorf("expected total pages 3, got %d", result.Pagination.TotalPages)
	}

	if !result.Pagination.HasNext {
		t.Error("expected has next to be true")
	}

	if result.Pagination.HasPrev {
		t.Error("expected has prev to be false")
	}
}

func TestPaginate_LastPage(t *testing.T) {
	notes := createTestNotes(25)

	result := Paginate(notes, 3, 10)

	if len(result.Notes) != 5 {
		t.Errorf("expected 5 notes on last page, got %d", len(result.Notes))
	}

	if result.Pagination.Page != 3 {
		t.Errorf("expected page 3, got %d", result.Pagination.Page)
	}

	if result.Pagination.TotalPages != 3 {
		t.Errorf("expected total pages 3, got %d", result.Pagination.TotalPages)
	}

	if result.Pagination.HasNext {
		t.Error("expected has next to be false")
	}

	if !result.Pagination.HasPrev {
		t.Error("expected has prev to be true")
	}
}

func TestPaginate_PageBeyondTotal(t *testing.T) {
	notes := createTestNotes(25)

	result := Paginate(notes, 10, 10)

	if len(result.Notes) != 0 {
		t.Errorf("expected 0 notes, got %d", len(result.Notes))
	}

	if result.Pagination.Total != 25 {
		t.Errorf("expected total 25, got %d", result.Pagination.Total)
	}

	if result.Pagination.HasNext {
		t.Error("expected has next to be false")
	}

	if !result.Pagination.HasPrev {
		t.Error("expected has prev to be true")
	}
}

func TestPaginate_ZeroLimit(t *testing.T) {
	notes := createTestNotes(25)

	result := Paginate(notes, 1, 0)

	if result.Pagination.Limit != DefaultLimit {
		t.Errorf("expected limit %d, got %d", DefaultLimit, result.Pagination.Limit)
	}

	if len(result.Notes) != 25 {
		t.Errorf("expected 25 notes with default limit, got %d", len(result.Notes))
	}
}

func TestPaginate_NegativeLimit(t *testing.T) {
	notes := createTestNotes(25)

	result := Paginate(notes, 1, -5)

	if result.Pagination.Limit != DefaultLimit {
		t.Errorf("expected limit %d, got %d", DefaultLimit, result.Pagination.Limit)
	}
}

func TestPaginate_NegativePage(t *testing.T) {
	notes := createTestNotes(25)

	result := Paginate(notes, -1, 10)

	if result.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Pagination.Page)
	}

	if len(result.Notes) != 10 {
		t.Errorf("expected 10 notes, got %d", len(result.Notes))
	}
}

func TestPaginate_ZeroPage(t *testing.T) {
	notes := createTestNotes(25)

	result := Paginate(notes, 0, 10)

	if result.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Pagination.Page)
	}
}

func TestPaginate_LargeLimitCapped(t *testing.T) {
	notes := createTestNotes(50)

	// Test with limit exceeding MaxLimit (10000)
	result := Paginate(notes, 1, 20000)

	if result.Pagination.Limit != MaxLimit {
		t.Errorf("expected limit capped at %d, got %d", MaxLimit, result.Pagination.Limit)
	}

	if len(result.Notes) != 50 {
		t.Errorf("expected 50 notes (all), got %d", len(result.Notes))
	}
}

func TestPaginate_LimitAtMax(t *testing.T) {
	notes := createTestNotes(100)

	result := Paginate(notes, 1, 1000)

	if result.Pagination.Limit != 1000 {
		t.Errorf("expected limit 1000, got %d", result.Pagination.Limit)
	}
}

func TestPaginate_EmptyNotes(t *testing.T) {
	notes := []models.Note{}

	result := Paginate(notes, 1, 10)

	if len(result.Notes) != 0 {
		t.Errorf("expected 0 notes, got %d", len(result.Notes))
	}

	if result.Pagination.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Pagination.Total)
	}

	if result.Pagination.TotalPages != 0 {
		t.Errorf("expected total pages 0, got %d", result.Pagination.TotalPages)
	}

	if result.Pagination.HasNext {
		t.Error("expected has next to be false")
	}

	if result.Pagination.HasPrev {
		t.Error("expected has prev to be false")
	}
}

func TestPaginate_ExactPageMatch(t *testing.T) {
	notes := createTestNotes(25)

	result := Paginate(notes, 2, 10)

	if len(result.Notes) != 10 {
		t.Errorf("expected 10 notes, got %d", len(result.Notes))
	}

	if result.Pagination.TotalPages != 3 {
		t.Errorf("expected total pages 3, got %d", result.Pagination.TotalPages)
	}

	if !result.Pagination.HasNext {
		t.Error("expected has next to be true")
	}

	if !result.Pagination.HasPrev {
		t.Error("expected has prev to be true")
	}
}

func TestPaginate_DefaultValues(t *testing.T) {
	notes := createTestNotes(100)

	result := Paginate(notes, 0, 0)

	if result.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Pagination.Page)
	}

	if result.Pagination.Limit != DefaultLimit {
		t.Errorf("expected limit %d, got %d", DefaultLimit, result.Pagination.Limit)
	}

	if len(result.Notes) != DefaultLimit {
		t.Errorf("expected %d notes, got %d", DefaultLimit, len(result.Notes))
	}
}
